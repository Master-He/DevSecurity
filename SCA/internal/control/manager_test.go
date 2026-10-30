package control

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"devsecurity/sca/internal/model"
	"devsecurity/sca/internal/output"
	"devsecurity/sca/internal/workspace"
)

const testScanID = "123e4567-e89b-42d3-a456-426614174000"

type fakeRunner struct {
	mu          sync.Mutex
	calls       int
	sourceKinds []string
}

func (r *fakeRunner) Run(_ context.Context, layout workspace.Layout, sourceKind string) error {
	r.mu.Lock()
	r.calls++
	r.sourceKinds = append(r.sourceKinds, sourceKind)
	r.mu.Unlock()
	return output.WriteJSON(layout.ResultPath(), model.Result{Engine: "test"})
}

func TestManagerRunsCLIAndIsIdempotent(t *testing.T) {
	runner := &fakeRunner{}
	manager := NewManager(runner, 1, time.Minute)
	layout := testLayout(t)
	request := testRequest()

	first, created, err := manager.Submit(request, layout)
	if err != nil || !created {
		t.Fatalf("submit: created=%v err=%v", created, err)
	}
	second, created, err := manager.Submit(request, layout)
	if err != nil || created || second.ScanID != first.ScanID {
		t.Fatalf("idempotent submit: created=%v err=%v", created, err)
	}
	view := waitForState(t, manager, StateSucceeded)
	if view.ResultPath != layout.ResultReference() {
		t.Fatalf("unexpected result path: %s", view.ResultPath)
	}
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.calls != 1 || len(runner.sourceKinds) != 1 || runner.sourceKinds[0] != "directory" {
		t.Fatalf("unexpected runner calls: %#v", runner.sourceKinds)
	}
}

func TestManagerRejectsConflictingRequest(t *testing.T) {
	manager := NewManager(&fakeRunner{}, 1, time.Minute)
	layout := testLayout(t)
	request := testRequest()
	if _, _, err := manager.Submit(request, layout); err != nil {
		t.Fatal(err)
	}
	request.SourceKind = "zip"
	if _, _, err := manager.Submit(request, layout); !errors.Is(err, ErrScanExists) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestManagerCancelsQueuedJob(t *testing.T) {
	manager := NewManager(&fakeRunner{}, 1, time.Minute)
	layout := testLayout(t)
	request := testRequest()
	if _, _, err := manager.Submit(request, layout); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Cancel(request.ScanID); err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, StateCanceled)
}

func TestManagerCancelsRunningJob(t *testing.T) {
	runner := &blockingRunner{started: make(chan struct{})}
	manager := NewManager(runner, 1, time.Minute)
	layout := testLayout(t)
	request := testRequest()
	if _, _, err := manager.Submit(request, layout); err != nil {
		t.Fatal(err)
	}
	select {
	case <-runner.started:
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not start")
	}
	if _, err := manager.Cancel(request.ScanID); err != nil {
		t.Fatal(err)
	}
	waitForState(t, manager, StateCanceled)
}

type blockingRunner struct {
	started chan struct{}
}

func (r *blockingRunner) Run(ctx context.Context, _ workspace.Layout, _ string) error {
	close(r.started)
	<-ctx.Done()
	return ctx.Err()
}

func testRequest() ScanRequest {
	return ScanRequest{
		ContractVersion: "1.0",
		ScanID:          testScanID,
		SourceKind:      "directory",
		Profile:         "java-maven",
	}
}

func testLayout(t *testing.T) workspace.Layout {
	t.Helper()
	root := t.TempDir()
	layout, err := workspace.New(root, testScanID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(layout.InputRoot, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout.InputRoot, "pom.xml"),
		[]byte("<project/>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.Ready, []byte("ready"), 0o600); err != nil {
		t.Fatal(err)
	}
	return layout
}

func waitForState(t *testing.T, manager *Manager, expected string) JobView {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		view, err := manager.Get(testScanID)
		if err != nil {
			t.Fatal(err)
		}
		if view.State == expected {
			return view
		}
		time.Sleep(10 * time.Millisecond)
	}
	view, _ := manager.Get(testScanID)
	t.Fatalf("scan did not reach %s: %#v", expected, view)
	return JobView{}
}
