package server

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devsecurity/sast-joern-adapter/internal/analyzer"
)

type Config struct {
	Scanner         analyzer.Scanner
	ServiceToken    string
	MaxArchiveBytes int64
	MaxUncompressed int64
	MaxEntries      int
}

type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func New(config Config) http.Handler {
	if config.MaxArchiveBytes <= 0 {
		config.MaxArchiveBytes = 100 * 1024 * 1024
	}
	if config.MaxUncompressed <= 0 {
		config.MaxUncompressed = 500 * 1024 * 1024
	}
	if config.MaxEntries <= 0 {
		config.MaxEntries = 20_000
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(writer http.ResponseWriter, request *http.Request) {
		available := config.Scanner.Available()
		status := http.StatusOK
		state := "READY"
		if !available {
			status = http.StatusServiceUnavailable
			state = "JOERN_NOT_AVAILABLE"
		}
		writeJSON(writer, status, map[string]any{
			"status":        state,
			"engine":        "joern",
			"engineVersion": analyzer.EngineVersion,
			"sourceMode":    "JOERN_CPGQL",
		})
	})
	mux.HandleFunc("POST /v1/analyze", func(writer http.ResponseWriter, request *http.Request) {
		if !authorized(request, config.ServiceToken) {
			writeError(writer, http.StatusUnauthorized, "SAST_UNAUTHORIZED", "服务令牌无效")
			return
		}
		if request.Header.Get("Content-Type") != "application/zip" {
			writeError(writer, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE",
				"仅支持 application/zip 源码包")
			return
		}
		workRoot, err := os.MkdirTemp("", "devsecurity-sast-*")
		if err != nil {
			writeError(writer, http.StatusInternalServerError, "SAST_WORKSPACE_ERROR", "无法创建扫描目录")
			return
		}
		defer os.RemoveAll(workRoot)

		archivePath := filepath.Join(workRoot, "source.zip")
		archive, err := os.Create(archivePath)
		if err != nil {
			writeError(writer, http.StatusInternalServerError, "SAST_WORKSPACE_ERROR", "无法创建源码文件")
			return
		}
		body := http.MaxBytesReader(writer, request.Body, config.MaxArchiveBytes+1)
		written, copyErr := io.Copy(archive, body)
		closeErr := archive.Close()
		if copyErr != nil || closeErr != nil {
			writeError(writer, http.StatusBadRequest, "INVALID_ARCHIVE", "无法读取源码归档")
			return
		}
		if written > config.MaxArchiveBytes {
			writeError(writer, http.StatusRequestEntityTooLarge, "ARCHIVE_TOO_LARGE", "源码归档超过大小限制")
			return
		}

		sourceRoot := filepath.Join(workRoot, "source")
		if err := os.MkdirAll(sourceRoot, 0o700); err != nil {
			writeError(writer, http.StatusInternalServerError, "SAST_WORKSPACE_ERROR", "无法准备扫描目录")
			return
		}
		hasJava, err := analyzer.ExtractZip(archivePath, sourceRoot, analyzer.ArchiveLimits{
			MaxEntries:      config.MaxEntries,
			MaxUncompressed: config.MaxUncompressed,
		})
		if err != nil {
			writeError(writer, http.StatusBadRequest, "INVALID_ARCHIVE", err.Error())
			return
		}
		if !hasJava {
			writeError(writer, http.StatusBadRequest, "JAVA_SOURCE_NOT_FOUND", "源码中未找到 .java 文件")
			return
		}

		result, err := config.Scanner.Analyze(request.Context(), sourceRoot, workRoot)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				writeError(writer, http.StatusGatewayTimeout, "SAST_TIMEOUT", "SAST 扫描执行超时")
				return
			}
			slog.Error("SAST analysis failed",
				"scan_id", request.Header.Get("X-Scan-Id"),
				"error", err)
			writeError(writer, http.StatusBadGateway, "SAST_FAILED", "Joern 扫描执行失败")
			return
		}
		writeJSON(writer, http.StatusOK, result)
	})
	return recoverer(requestLogger(mux))
}

func authorized(request *http.Request, expected string) bool {
	if expected == "" {
		return false
	}
	actual := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
	return len(actual) == len(expected) &&
		subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	response := apiError{}
	response.Error.Code = code
	response.Error.Message = message
	writeJSON(writer, status, response)
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic in SAST service", "panic", recovered)
				writeError(writer, http.StatusInternalServerError, "INTERNAL_ERROR", "服务内部错误")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		next.ServeHTTP(writer, request)
		slog.Info("request",
			"method", request.Method,
			"path", request.URL.Path,
			"scan_id", request.Header.Get("X-Scan-Id"),
			"duration_ms", time.Since(start).Milliseconds())
	})
}
