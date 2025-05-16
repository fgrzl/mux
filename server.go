package mux

import (
	"context"
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

	return &WebServer{
		srv:    srv,
		Router: router,
	}
}

// Start runs the HTTP server in a goroutine and handles graceful shutdown.
func (ws *WebServer) Start(ctx context.Context) {
	go func() {
		if err := ws.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Replace with slog or your logger if needed
			panic("HTTP server error: " + err.Error())
		}
	}()

	// Optional: hook for shutdown with context
	go func() {
		<-ctx.Done()
		_ = ws.Stop(context.Background())
	}()
}

// Stop shuts down the HTTP server gracefully.
func (ws *WebServer) Stop(ctx context.Context) error {
	return ws.srv.Shutdown(ctx)
}
