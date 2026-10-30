package control

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"devsecurity/sca/internal/workspace"
)

type CLIRunner struct {
	Executable string
	DataPath   string
}

func NewCLIRunner(executable, dataPath string) *CLIRunner {
	return &CLIRunner{Executable: executable, DataPath: dataPath}
}

func (r *CLIRunner) Run(ctx context.Context, layout workspace.Layout, sourceKind string) error {
	command := exec.CommandContext(ctx, r.Executable,
		"analyze",
		"--input", r.InputPath(layout, sourceKind),
		"--output", layout.ResultPath(),
		"--data", r.DataPath,
	)
	command.Dir = layout.ScanRoot
	command.Env = scannerEnvironment()
	configureProcess(command)
	command.Cancel = func() error {
		terminateProcessTree(command)
		return nil
	}
	command.WaitDelay = 5 * time.Second

	stdout, err := os.OpenFile(filepath.Join(layout.Logs, "sca.stdout.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open CLI stdout log: %w", err)
	}
	defer stdout.Close()
	stderr, err := os.OpenFile(filepath.Join(layout.Logs, "sca.stderr.log"),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open CLI stderr log: %w", err)
	}
	defer stderr.Close()
	command.Stdout = stdout
	command.Stderr = stderr

	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		detail := readLogTail(filepath.Join(layout.Logs, "sca.stderr.log"), layout, 1200)
		if detail == "" {
			return fmt.Errorf("SCA CLI failed: %w", err)
		}
		return fmt.Errorf("SCA CLI failed: %w: %s", err, detail)
	}
	result, err := os.Stat(layout.ResultPath())
	if err != nil {
		return fmt.Errorf("SCA CLI did not publish result.json: %w", err)
	}
	if !result.Mode().IsRegular() || result.Size() == 0 {
		return fmt.Errorf("SCA CLI result.json is empty or invalid")
	}
	return nil
}

func (r *CLIRunner) InputPath(layout workspace.Layout, sourceKind string) string {
	if sourceKind == "zip" {
		return filepath.Join(layout.InputRoot, "source.zip")
	}
	return layout.InputRoot
}

func CLIVersion(executable string) (string, error) {
	command := exec.Command(executable, "version")
	command.Env = scannerEnvironment()
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("run SCA CLI version: %w", err)
	}
	return strings.TrimSpace(output.String()), nil
}

func scannerEnvironment() []string {
	allowed := map[string]bool{
		"PATH": true, "JAVA_HOME": true, "MAVEN_HOME": true,
		"SYSTEMROOT": true, "WINDIR": true, "TEMP": true, "TMP": true,
		"SCA_USE_MAVEN": true,
	}
	values := make([]string, 0)
	for _, value := range os.Environ() {
		name := strings.ToUpper(strings.SplitN(value, "=", 2)[0])
		if allowed[name] || strings.HasPrefix(name, "LC_") {
			values = append(values, value)
		}
	}
	return values
}

func readLogTail(path string, layout workspace.Layout, maximum int64) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.Size() == 0 {
		return ""
	}
	offset := info.Size() - maximum
	if offset < 0 {
		offset = 0
	}
	if _, err := file.Seek(offset, 0); err != nil {
		return ""
	}
	content, err := io.ReadAll(io.LimitReader(file, maximum))
	if err != nil {
		return ""
	}
	detail := strings.TrimSpace(string(content))
	detail = strings.ReplaceAll(detail, "\r", " ")
	detail = strings.ReplaceAll(detail, "\n", " | ")
	detail = strings.ReplaceAll(detail, layout.ScanRoot, "<scan-workspace>")
	detail = strings.ReplaceAll(detail, layout.Root, "<work-root>")
	if len(detail) > 800 {
		detail = detail[len(detail)-800:]
	}
	return detail
}
