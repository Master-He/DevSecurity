package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"devsecurity/sast-joern-adapter/internal/analyzer"
	"devsecurity/sast-joern-adapter/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	address := env("SAST_LISTEN_ADDR", "127.0.0.1:8087")
	token := strings.TrimSpace(os.Getenv("SAST_SERVICE_TOKEN"))
	if token == "" && !strings.HasPrefix(address, "127.0.0.1:") &&
		!strings.HasPrefix(address, "localhost:") {
		slog.Error("SAST_SERVICE_TOKEN is required beyond loopback")
		os.Exit(1)
	}
	if token == "" {
		token = "123"
	}
	joernHome, err := resolveJoernHome()
	if err != nil {
		slog.Error("Joern runtime is unavailable", "error", err)
		os.Exit(1)
	}
	timeoutSeconds, err := strconv.Atoi(env("SAST_TIMEOUT_SECONDS", "1800"))
	if err != nil || timeoutSeconds <= 0 {
		timeoutSeconds = 1800
	}
	handler := server.New(server.Config{
		Scanner: analyzer.Scanner{
			JoernHome: joernHome,
			Timeout:   time.Duration(timeoutSeconds) * time.Second,
		},
		ServiceToken: token,
	})
	httpServer := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      30 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info("Joern adapter listening",
		"address", httpServer.Addr,
		"joern_home", joernHome,
		"engine_version", analyzer.EngineVersion)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Joern adapter stopped", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownContext); err != nil {
		slog.Error("SAST service shutdown failed", "error", err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func resolveJoernHome() (string, error) {
	candidates := make([]string, 0, 5)
	if configured := strings.TrimSpace(os.Getenv("SAST_JOERN_HOME")); configured != "" {
		candidates = append(candidates, configured)
	}
	if executable, err := os.Executable(); err == nil {
		base := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(base, "..", ".runtime", "joern-cli"),
			filepath.Join(base, ".runtime", "joern-cli"))
	}
	if workingDirectory, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(workingDirectory, "..", ".runtime", "joern-cli"),
			filepath.Join(workingDirectory, ".runtime", "joern-cli"))
	}
	for _, candidate := range candidates {
		absolute, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		info, err := os.Stat(filepath.Join(absolute, "lib",
			"io.joern.joern-cli-"+analyzer.EngineVersion+".jar"))
		if err == nil && info.Mode().IsRegular() {
			return absolute, nil
		}
	}
	return "", errors.New("set SAST_JOERN_HOME to the extracted joern-cli directory")
}
