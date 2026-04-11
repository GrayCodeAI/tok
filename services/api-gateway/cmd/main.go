package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GrayCodeAI/tokman/services/api-gateway/internal/handler"
	"github.com/GrayCodeAI/tokman/services/api-gateway/internal/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Service endpoints from environment
	compressionAddr := getEnv("COMPRESSION_SERVICE_ADDR", "localhost:50051")
	analyticsAddr := getEnv("ANALYTICS_SERVICE_ADDR", "localhost:50052")
	securityAddr := getEnv("SECURITY_SERVICE_ADDR", "localhost:50053")
	configAddr := getEnv("CONFIG_SERVICE_ADDR", "localhost:50054")

	// Create gateway handler
	gateway, err := handler.NewGateway(compressionAddr, analyticsAddr, securityAddr, configAddr)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}
	defer gateway.Close()

	// Setup routes
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", gateway.HealthHandler())
	mux.HandleFunc("/health/services", gateway.ServicesHealthHandler())

	// API v1 routes
	mux.HandleFunc("/api/v1/compress", gateway.CompressHandler())
	mux.HandleFunc("/api/v1/compress/stream", gateway.CompressStreamHandler())
	mux.HandleFunc("/api/v1/scan", gateway.ScanHandler())
	mux.HandleFunc("/api/v1/redact", gateway.RedactHandler())
	mux.HandleFunc("/api/v1/stats", gateway.StatsHandler())
	mux.HandleFunc("/api/v1/commands", gateway.CommandsHandler())
	mux.HandleFunc("/api/v1/savings", gateway.SavingsHandler())
	mux.HandleFunc("/api/v1/filters", gateway.FiltersHandler())

	// Middleware chain
	handler := middleware.Recovery(
		middleware.Logging(
			middleware.RateLimit(
				middleware.CORS(mux),
			),
		),
	)

	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("API Gateway starting", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed", "error", err)
		}
	}()

	<-sigChan
	slog.Info("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Shutdown error", "error", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
