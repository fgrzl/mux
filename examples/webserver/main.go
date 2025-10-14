package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/fgrzl/mux"
)

func main() {
	// Create router
	router := mux.NewRouter()

	// Add health probes
	router.Healthz()
	router.Livez()
	router.Readyz()

	// Add routes
	router.GET("/", func(c mux.RouteContext) {
		c.OK(map[string]string{
			"message": "Hello from WebServer!",
			"version": "1.0.0",
		})
	})

	api := router.NewRouteGroup("/api/v1")
	api.GET("/status", func(c mux.RouteContext) {
		c.OK(map[string]string{
			"status": "running",
		})
	})

	// Create WebServer with production defaults
	// - ReadTimeout: 10s
	// - WriteTimeout: 10s
	// - IdleTimeout: 120s
	server := mux.NewServer(":8080", router)

	// Setup graceful shutdown on SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	slog.Info("Starting server", "addr", ":8080")
	slog.Info("Press Ctrl+C to shutdown gracefully")

	// Listen blocks until shutdown
	if err := server.Listen(ctx); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}

	slog.Info("Server shutdown complete")
}
