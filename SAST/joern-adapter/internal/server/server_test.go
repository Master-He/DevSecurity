package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"devsecurity/sast-joern-adapter/internal/analyzer"
)

func TestAnalyzeRequiresServiceToken(t *testing.T) {
	handler := New(Config{
		Scanner:      analyzer.Scanner{JoernHome: t.TempDir()},
		ServiceToken: "expected",
	})
	request := httptest.NewRequest(http.MethodPost, "/v1/analyze", nil)
	request.Header.Set("Content-Type", "application/zip")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestHealthReportsMissingJoern(t *testing.T) {
	handler := New(Config{
		Scanner:      analyzer.Scanner{JoernHome: t.TempDir()},
		ServiceToken: "expected",
	})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
}

func TestAuthorizedUsesExactBearerToken(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/v1/analyze", nil)
	request.Header.Set("Authorization", "Bearer expected")
	if !authorized(request, "expected") {
		t.Fatal("expected valid token")
	}
	request.Header.Set("Authorization", "Bearer expected-extra")
	if authorized(request, "expected") {
		t.Fatal("expected prefix token to be rejected")
	}
}
