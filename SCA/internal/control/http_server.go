package control

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"devsecurity/sca/internal/workspace"
)

type HTTPServer struct {
	manager    *Manager
	workRoot   string
	authToken  string
	cliVersion string
}

func NewHTTPServer(manager *Manager, workRoot, authToken, cliVersion string) *HTTPServer {
	return &HTTPServer{
		manager:    manager,
		workRoot:   workRoot,
		authToken:  authToken,
		cliVersion: cliVersion,
	}
}

func (s *HTTPServer) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.Handle("POST /v1/scans", s.authorize(http.HandlerFunc(s.submit)))
	mux.Handle("GET /v1/scans/{scan_id}", s.authorize(http.HandlerFunc(s.status)))
	mux.Handle("POST /v1/scans/{scan_id}/cancel", s.authorize(http.HandlerFunc(s.cancel)))
	return requestLogging(recoverer(mux))
}

func (s *HTTPServer) health(writer http.ResponseWriter, _ *http.Request) {
	info, err := os.Stat(s.workRoot)
	accessible := err == nil && info.IsDir()
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":               "UP",
		"engine":               "sca-maven",
		"cli_version":          s.cliVersion,
		"work_root_accessible": accessible,
	})
}

func (s *HTTPServer) submit(writer http.ResponseWriter, request *http.Request) {
	if !hasJSONContentType(request) {
		writeError(writer, http.StatusUnsupportedMediaType, "UNSUPPORTED_MEDIA_TYPE",
			"Content-Type must be application/json")
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, 64*1024)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var payload ScanRequest
	if err := decoder.Decode(&payload); err != nil {
		writeError(writer, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON request")
		return
	}
	if err := requireJSONEOF(decoder); err != nil {
		writeError(writer, http.StatusBadRequest, "INVALID_REQUEST", "request must contain one JSON object")
		return
	}
	if payload.ContractVersion != "1.0" {
		writeError(writer, http.StatusBadRequest, "UNSUPPORTED_CONTRACT_VERSION",
			"contract_version must be 1.0")
		return
	}
	if payload.SourceKind != "directory" && payload.SourceKind != "zip" {
		writeError(writer, http.StatusBadRequest, "UNSUPPORTED_SOURCE_KIND",
			"source_kind must be directory or zip")
		return
	}
	if payload.Profile != "" && payload.Profile != "java-maven" {
		writeError(writer, http.StatusBadRequest, "UNSUPPORTED_PROFILE",
			"profile must be empty or java-maven")
		return
	}
	if payload.Limits.TimeoutSeconds < 0 || payload.Limits.TimeoutSeconds > 7*24*60*60 ||
		payload.Limits.MaxCPU < 0 || payload.Limits.MaxMemoryMB < 0 ||
		payload.Limits.MaxDiskBytes < 0 {
		writeError(writer, http.StatusBadRequest, "INVALID_LIMITS",
			"scan limits must be non-negative and timeout_seconds must not exceed 604800")
		return
	}
	layout, err := workspace.New(s.workRoot, payload.ScanID)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "INVALID_SCAN_ID", err.Error())
		return
	}
	if _, err := layout.ValidateInput(payload.SourceKind); err != nil {
		writeError(writer, http.StatusConflict, "WORKSPACE_NOT_READY", err.Error())
		return
	}
	view, _, err := s.manager.Submit(payload, layout)
	if errors.Is(err, ErrScanExists) {
		writeError(writer, http.StatusConflict, "SCAN_CONFLICT",
			"scan_id already exists with different request parameters")
		return
	}
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "INTERNAL_ERROR", "cannot submit scan")
		return
	}
	writer.Header().Set("Location", "/v1/scans/"+view.ScanID)
	writeJSON(writer, http.StatusAccepted, view)
}

func (s *HTTPServer) status(writer http.ResponseWriter, request *http.Request) {
	view, err := s.manager.Get(request.PathValue("scan_id"))
	if errors.Is(err, ErrScanNotFound) {
		writeError(writer, http.StatusNotFound, "SCAN_NOT_FOUND", "scan not found")
		return
	}
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "INTERNAL_ERROR", "cannot read scan")
		return
	}
	writeJSON(writer, http.StatusOK, view)
}

func (s *HTTPServer) cancel(writer http.ResponseWriter, request *http.Request) {
	view, err := s.manager.Cancel(request.PathValue("scan_id"))
	if errors.Is(err, ErrScanNotFound) {
		writeError(writer, http.StatusNotFound, "SCAN_NOT_FOUND", "scan not found")
		return
	}
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "INTERNAL_ERROR", "cannot cancel scan")
		return
	}
	writeJSON(writer, http.StatusAccepted, view)
}

func (s *HTTPServer) authorize(next http.Handler) http.Handler {
	if s.authToken == "" {
		return next
	}
	expected := []byte("Bearer " + s.authToken)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		actual := []byte(request.Header.Get("Authorization"))
		if len(actual) != len(expected) || subtle.ConstantTimeCompare(actual, expected) != 1 {
			writer.Header().Set("WWW-Authenticate", "Bearer")
			writeError(writer, http.StatusUnauthorized, "UNAUTHORIZED", "invalid service token")
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("multiple JSON values")
	}
	return err
}

func hasJSONContentType(request *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	return err == nil && strings.EqualFold(mediaType, "application/json")
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func requestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(writer, request)
		slog.Info("request",
			"method", request.Method,
			"path", request.URL.Path,
			"scan_id", request.PathValue("scan_id"))
	})
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic while serving request", "panic", recovered,
					"stack", string(debug.Stack()))
				writeError(writer, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}
