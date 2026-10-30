package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"devsecurity/sast-joern-adapter/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	handler := server.New()
	httpServer := &http.Server{
		Addr:              ":" + env("PORT", "8087"),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      30 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}
	slog.Info("Joern adapter listening", "address", httpServer.Addr, "available", server.JoernAvailable())
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("Joern adapter stopped", "error", err)
		os.Exit(1)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
