package mux

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// WebServer wraps an http.Server and your custom Router.
type WebServer struct {
	srv    *http.Server
	Router *Router
}

// NewServer sets up the HTTP server with sane timeouts and a mux Router.
func NewServer(addr string, router *Router) *WebServer {
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return &WebServer{srv: srv, Router: router}
}

// Start runs the HTTP server in a goroutine and handles graceful shutdown.
func (ws *WebServer) Start(ctx context.Context) error {
	// Validate address (optional)
	if _, err := net.ResolveTCPAddr("tcp", ws.srv.Addr); err != nil {
		return err
	}

	go func() {
		if err := ws.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = ws.Stop(context.Background())
	}()

	return nil
}

// Stop shuts down the HTTP server gracefully.
func (ws *WebServer) Stop(ctx context.Context) error {
	return ws.srv.Shutdown(ctx)
}
