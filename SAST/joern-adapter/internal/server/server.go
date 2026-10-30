package server

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func JoernAvailable() bool {
	_, err := exec.LookPath(binary())
	return err == nil
}

func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	mux.HandleFunc("POST /v1/analyze", analyze)
	return mux
}

func health(writer http.ResponseWriter, request *http.Request) {
	status := "READY"
	code := http.StatusOK
	if !JoernAvailable() {
		status = "JOERN_NOT_AVAILABLE"
		code = http.StatusServiceUnavailable
	}
	writeJSON(writer, code, map[string]any{
		"status": status,
		"engine": "joern",
		"mode":   "CPG_INGESTION_ONLY",
	})
}

func analyze(writer http.ResponseWriter, request *http.Request) {
	if !JoernAvailable() {
		writeJSON(writer, http.StatusServiceUnavailable, map[string]string{
			"code": "JOERN_NOT_AVAILABLE", "message": "当前镜像中没有 joern-parse",
		})
		return
	}
	if request.Header.Get("Content-Type") != "application/zip" {
		writeJSON(writer, http.StatusUnsupportedMediaType, map[string]string{"message": "仅支持 ZIP"})
		return
	}
	body := http.MaxBytesReader(writer, request.Body, 100*1024*1024)
	archive, err := os.CreateTemp("", "joern-source-*.zip")
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"message": "临时文件创建失败"})
		return
	}
	defer os.Remove(archive.Name())
	defer archive.Close()
	if _, err := io.Copy(archive, body); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"message": "源码读取失败"})
		return
	}
	root, err := os.MkdirTemp("", "joern-source-*")
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]string{"message": "工作目录创建失败"})
		return
	}
	defer os.RemoveAll(root)
	if err := extract(archive.Name(), root); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	timeout := 10 * time.Minute
	context, cancel := context.WithTimeout(request.Context(), timeout)
	defer cancel()
	command := exec.CommandContext(context, binary(), root)
	command.Dir = root
	output, err := command.CombinedOutput()
	if context.Err() != nil {
		writeJSON(writer, http.StatusGatewayTimeout, map[string]string{
			"code": "JOERN_TIMEOUT", "message": "Joern CPG 构建超时",
		})
		return
	}
	if err != nil {
		slog.Warn("Joern ingestion failed", "error", err)
		writeJSON(writer, http.StatusBadGateway, map[string]string{
			"code": "JOERN_FAILED", "message": "Joern CPG 构建失败",
		})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"engine":      "joern",
		"status":      "CPG_READY",
		"findings":    []any{},
		"diagnostics": []string{"当前适配器只完成 CPG 构建；规则查询和统一 Finding 映射尚未启用。"},
		"outputBytes": len(output),
		"completedAt": time.Now().UTC(),
	})
}

func binary() string {
	if value := os.Getenv("JOERN_PARSE_BINARY"); value != "" {
		return value
	}
	return "joern-parse"
}

func extract(path, target string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("ZIP 格式不合法")
	}
	defer reader.Close()
	for _, entry := range reader.File {
		name := filepath.Clean(strings.ReplaceAll(entry.Name, `\`, "/"))
		if name == "." || filepath.IsAbs(name) || name == ".." ||
			strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("源码包含不安全路径")
		}
		destination := filepath.Join(target, filepath.FromSlash(name))
		relative, err := filepath.Rel(target, destination)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("源码包含路径穿越")
		}
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(destination, 0o755); err != nil {
				return fmt.Errorf("解压失败")
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return fmt.Errorf("解压失败")
		}
		source, err := entry.Open()
		if err != nil {
			return fmt.Errorf("无法读取源码条目")
		}
		file, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			source.Close()
			return fmt.Errorf("解压失败")
		}
		_, copyErr := io.Copy(file, io.LimitReader(source, 500*1024*1024))
		source.Close()
		file.Close()
		if copyErr != nil {
			return fmt.Errorf("解压失败")
		}
	}
	return nil
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
