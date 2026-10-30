package control

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPServerRequiresTokenAndRunsScan(t *testing.T) {
	runner := &fakeRunner{}
	manager := NewManager(runner, 1, time.Minute)
	layout := testLayout(t)
	handler := NewHTTPServer(manager, layout.Root, "test-token", "0.2.0").Handler()
	body, err := json.Marshal(testRequest())
	if err != nil {
		t.Fatal(err)
	}

	unauthorized := httptest.NewRequest(http.MethodPost, "/v1/scans", bytes.NewReader(body))
	unauthorized.Header.Set("Content-Type", "application/json")
	unauthorizedResponse := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedResponse, unauthorized)
	if unauthorizedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", unauthorizedResponse.Code)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/scans", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer test-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", response.Code, response.Body.String())
	}
	waitForState(t, manager, StateSucceeded)

	statusRequest := httptest.NewRequest(http.MethodGet, "/v1/scans/"+testScanID, nil)
	statusRequest.Header.Set("Authorization", "Bearer test-token")
	statusResponse := httptest.NewRecorder()
	handler.ServeHTTP(statusResponse, statusRequest)
	if statusResponse.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", statusResponse.Code)
	}
	if bytes.Contains(statusResponse.Body.Bytes(), []byte("<project")) {
		t.Fatal("status response must not contain source content")
	}
}

func TestHTTPServerRejectsInvalidScanID(t *testing.T) {
	manager := NewManager(&fakeRunner{}, 1, time.Minute)
	root := t.TempDir()
	handler := NewHTTPServer(manager, root, "", "0.2.0").Handler()
	request := httptest.NewRequest(http.MethodPost, "/v1/scans",
		bytes.NewBufferString(`{"contract_version":"1.0","scan_id":"../../bad","source_kind":"directory"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
}
