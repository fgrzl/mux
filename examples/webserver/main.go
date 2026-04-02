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
	router := mux.NewRouter()

	if err := router.Configure(func(router *mux.Router) {
		router.Healthz()
		router.Livez()
		router.Readyz()

		router.GET("/", func(c mux.RouteContext) {
			c.OK(map[string]string{
				"message": "Hello from WebServer!",
				"version": "1.0.0",
			})
		})

		api := router.Group("/api/v1")
		api.GET("/status", func(c mux.RouteContext) {
			c.OK(map[string]string{
				"status": "running",
			})
		})
	}); err != nil {
		slog.Error("Invalid router configuration", "error", err)
		os.Exit(1)
	}

	server := mux.NewServer(":8080", router)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	slog.Info("Starting server", "addr", ":8080")
	slog.Info("Press Ctrl+C to shutdown gracefully")

	if err := server.Listen(ctx); err != nil {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}

	slog.Info("Server shutdown complete")
}
