package control

import "time"

const (
	StateQueued    = "queued"
	StateRunning   = "running"
	StateSucceeded = "succeeded"
	StateFailed    = "failed"
	StateCanceled  = "canceled"
)

type Limits struct {
	TimeoutSeconds int   `json:"timeout_seconds"`
	MaxCPU         int   `json:"max_cpu"`
	MaxMemoryMB    int   `json:"max_memory_mb"`
	MaxDiskBytes   int64 `json:"max_disk_bytes"`
}

type ScanRequest struct {
	ContractVersion string `json:"contract_version"`
	ScanID          string `json:"scan_id"`
	SourceKind      string `json:"source_kind"`
	Profile         string `json:"profile"`
	Limits          Limits `json:"limits"`
}

type JobView struct {
	ScanID      string     `json:"scan_id"`
	State       string     `json:"state"`
	SourceKind  string     `json:"source_kind"`
	Profile     string     `json:"profile,omitempty"`
	AcceptedAt  time.Time  `json:"accepted_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	ExitCode    *int       `json:"exit_code,omitempty"`
	ErrorCode   string     `json:"error_code,omitempty"`
	Message     string     `json:"message,omitempty"`
	ResultPath  string     `json:"result_path,omitempty"`
	TimeoutSecs int        `json:"timeout_seconds"`
}
