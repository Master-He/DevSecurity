package control

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"devsecurity/sca/internal/workspace"
)

var (
	ErrScanExists   = errors.New("scan already exists")
	ErrScanNotFound = errors.New("scan not found")
)

type Runner interface {
	Run(ctx context.Context, layout workspace.Layout, sourceKind string) error
}

type Manager struct {
	mu             sync.RWMutex
	jobs           map[string]*job
	slots          chan struct{}
	runner         Runner
	defaultTimeout time.Duration
}

type job struct {
	request     ScanRequest
	layout      workspace.Layout
	fingerprint string
	ctx         context.Context
	cancel      context.CancelFunc
	view        JobView
}

func NewManager(runner Runner, maxConcurrent int, defaultTimeout time.Duration) *Manager {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	if defaultTimeout <= 0 {
		defaultTimeout = 2 * time.Hour
	}
	return &Manager{
		jobs:           make(map[string]*job),
		slots:          make(chan struct{}, maxConcurrent),
		runner:         runner,
		defaultTimeout: defaultTimeout,
	}
}

func (m *Manager) Submit(request ScanRequest, layout workspace.Layout) (JobView, bool, error) {
	timeoutSeconds := request.Limits.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = int(m.defaultTimeout / time.Second)
	}
	request.Limits.TimeoutSeconds = timeoutSeconds
	fingerprint := requestFingerprint(request)
	m.mu.Lock()
	if existing, ok := m.jobs[request.ScanID]; ok {
		if existing.fingerprint != fingerprint {
			m.mu.Unlock()
			return JobView{}, false, ErrScanExists
		}
		view := existing.view
		m.mu.Unlock()
		return view, false, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	accepted := time.Now().UTC()
	current := &job{
		request:     request,
		layout:      layout,
		fingerprint: fingerprint,
		ctx:         ctx,
		cancel:      cancel,
		view: JobView{
			ScanID:      request.ScanID,
			State:       StateQueued,
			SourceKind:  request.SourceKind,
			Profile:     request.Profile,
			AcceptedAt:  accepted,
			TimeoutSecs: timeoutSeconds,
		},
	}
	m.jobs[request.ScanID] = current
	view := current.view
	m.mu.Unlock()

	go m.execute(current)
	return view, true, nil
}

func (m *Manager) Get(scanID string) (JobView, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	current, ok := m.jobs[scanID]
	if !ok {
		return JobView{}, ErrScanNotFound
	}
	return current.view, nil
}

func (m *Manager) Cancel(scanID string) (JobView, error) {
	m.mu.Lock()
	current, ok := m.jobs[scanID]
	if !ok {
		m.mu.Unlock()
		return JobView{}, ErrScanNotFound
	}
	if terminal(current.view.State) {
		view := current.view
		m.mu.Unlock()
		return view, nil
	}
	current.cancel()
	view := current.view
	m.mu.Unlock()
	return view, nil
}

func (m *Manager) execute(current *job) {
	defer current.cancel()
	select {
	case m.slots <- struct{}{}:
		defer func() { <-m.slots }()
	case <-current.ctx.Done():
		m.finish(current, StateCanceled, 0, "CANCELED", "scan canceled", "")
		return
	}

	started := time.Now().UTC()
	m.mu.Lock()
	if current.ctx.Err() != nil {
		current.view.State = StateCanceled
		current.view.FinishedAt = &started
		current.view.ErrorCode = "CANCELED"
		current.view.Message = "scan canceled"
		m.mu.Unlock()
		return
	}
	current.view.State = StateRunning
	current.view.StartedAt = &started
	m.mu.Unlock()

	if err := current.layout.EnsureOutputDirectories(); err != nil {
		m.finish(current, StateFailed, 1, "WORKSPACE_ERROR", err.Error(), "")
		return
	}
	timeout := time.Duration(current.view.TimeoutSecs) * time.Second
	ctx, cancel := context.WithTimeout(current.ctx, timeout)
	defer cancel()
	err := m.runner.Run(ctx, current.layout, current.request.SourceKind)
	switch {
	case errors.Is(current.ctx.Err(), context.Canceled):
		m.finish(current, StateCanceled, 1, "CANCELED", "scan canceled", "")
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		m.finish(current, StateFailed, 1, "TIMEOUT", "scan timed out", "")
	case err != nil:
		m.finish(current, StateFailed, 1, "CLI_FAILED", err.Error(), "")
	default:
		m.finish(current, StateSucceeded, 0, "", "", current.layout.ResultReference())
	}
}

func (m *Manager) finish(current *job, state string, exitCode int, code, message, resultPath string) {
	finished := time.Now().UTC()
	m.mu.Lock()
	defer m.mu.Unlock()
	current.view.State = state
	current.view.FinishedAt = &finished
	current.view.ExitCode = &exitCode
	current.view.ErrorCode = code
	current.view.Message = message
	current.view.ResultPath = resultPath
}

func terminal(state string) bool {
	return state == StateSucceeded || state == StateFailed || state == StateCanceled
}

func requestFingerprint(request ScanRequest) string {
	value := fmt.Sprintf("%s|%s|%s|%d|%d|%d|%d",
		request.SourceKind, request.Profile, request.ContractVersion,
		request.Limits.TimeoutSeconds, request.Limits.MaxCPU,
		request.Limits.MaxMemoryMB, request.Limits.MaxDiskBytes)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(value)))
}
