package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"devsecurity/sca/internal/control"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	if err := run(); err != nil {
		slog.Error("SCA server failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	address := strings.TrimSpace(os.Getenv("SCA_LISTEN_ADDR"))
	if address == "" {
		address = "127.0.0.1:" + env("PORT", "8086")
	}
	authToken := strings.TrimSpace(os.Getenv("SCA_SERVICE_TOKEN"))
	if !isLoopbackAddress(address) && authToken == "" {
		return fmt.Errorf("SCA_SERVICE_TOKEN is required when listening beyond loopback")
	}

	workRoot, err := filepath.Abs(env("SCA_WORK_ROOT", ".data/sca-work"))
	if err != nil {
		return fmt.Errorf("resolve SCA_WORK_ROOT: %w", err)
	}
	if err := os.MkdirAll(workRoot, 0o750); err != nil {
		return fmt.Errorf("create SCA work root: %w", err)
	}
	cliPath, err := resolveCLI()
	if err != nil {
		return err
	}
	cliVersion, err := control.CLIVersion(cliPath)
	if err != nil {
		return err
	}
	dataPath, err := resolveDataPath(configuredDataPath())
	if err != nil {
		return err
	}
	concurrency := envInt("SCA_MAX_CONCURRENT_SCANS", 4)
	timeout := time.Duration(envInt("SCA_DEFAULT_TIMEOUT_SECONDS", 7200)) * time.Second
	manager := control.NewManager(control.NewCLIRunner(cliPath, dataPath), concurrency, timeout)
	handler := control.NewHTTPServer(manager, workRoot, authToken, cliVersion)
	server := &http.Server{
		Addr:              address,
		Handler:           handler.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		slog.Info("SCA server listening",
			"address", address,
			"work_root", workRoot,
			"cli_version", cliVersion,
			"auth_enabled", authToken != "")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("SCA HTTP server stopped unexpectedly", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return server.Shutdown(shutdownContext)
}

func configuredDataPath() string {
	if configured := strings.TrimSpace(os.Getenv("SCA_VULN_DB")); configured != "" {
		return configured
	}
	if executable, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(executable), "data", "vulnerabilities.json")
		if info, statErr := os.Stat(candidate); statErr == nil && info.Mode().IsRegular() {
			return candidate
		}
	}
	return filepath.Join("data", "vulnerabilities.json")
}

func resolveDataPath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve vulnerability database path: %w", err)
	}
	if info, err := os.Stat(absolute); err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("vulnerability database is unavailable at %s", absolute)
	}
	return absolute, nil
}

func resolveCLI() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("SCA_CLI_PATH")); configured != "" {
		absolute, err := filepath.Abs(configured)
		if err != nil {
			return "", fmt.Errorf("resolve SCA_CLI_PATH: %w", err)
		}
		if info, err := os.Stat(absolute); err != nil || !info.Mode().IsRegular() {
			return "", fmt.Errorf("SCA CLI is not available at %s", absolute)
		}
		return absolute, nil
	}
	executable, err := os.Executable()
	if err == nil {
		directory := filepath.Dir(executable)
		for _, name := range []string{"sca", "sca.exe"} {
			candidate := filepath.Join(directory, name)
			if info, statErr := os.Stat(candidate); statErr == nil && info.Mode().IsRegular() {
				return candidate, nil
			}
		}
	}
	for _, name := range []string{"sca", "sca.exe"} {
		if candidate, lookErr := exec.LookPath(name); lookErr == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("SCA CLI not found; set SCA_CLI_PATH")
}

func isLoopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
