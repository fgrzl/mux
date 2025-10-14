package mux

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fgrzl/mux/pkg/router"
)

const errInvalidTLSFiles = "invalid TLS cert/key files"

// WebServerOption allows functional options to be passed to NewServer for configuring the WebServer.
type WebServerOption func(*WebServer)

// WebServer wraps an http.Server and a custom Router, providing methods for starting and stopping the server.
type WebServer struct {
	srv      *http.Server
	rtr      *router.Router
	certFile string
	keyFile  string
}

// WithTLS enables HTTPS for the WebServer using the provided certificate and key file paths.
func WithTLS(certFile, keyFile string) WebServerOption {
	return func(ws *WebServer) {
		ws.certFile = certFile
		ws.keyFile = keyFile
	}
}

// WithTLSDiscovery searches upward for a certs directory (up to 10 levels) and sets certFile/keyFile.
func WithTLSDiscovery(certsDir, certFile, keyFile string) WebServerOption {
	return func(ws *WebServer) {
		dir, err := os.Getwd()
		if err != nil {
			slog.Error("Could not get working directory for TLS discovery", "error", err)
			return
		}
		orig := dir
		for i := 0; i < 10; i++ {
			certsPath := filepath.Join(dir, certsDir)
			if stat, err := os.Stat(certsPath); err == nil && stat.IsDir() {
				ws.certFile = filepath.Join(certsPath, certFile)
				ws.keyFile = filepath.Join(certsPath, keyFile)
				return
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		slog.Error("Could not find certs directory for TLS discovery", "searched_from", orig)
	}
}

// NewServer creates a new WebServer with the given address, router, and optional configuration options.
func NewServer(addr string, rtr *router.Router, opts ...WebServerOption) *WebServer {
	srv := &http.Server{
		Addr:         addr,
		Handler:      rtr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	ws := &WebServer{srv: srv, rtr: rtr}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

// Start begins serving in the background and returns immediately.
// Shutdown will be triggered when ctx is canceled.
func (ws *WebServer) Start(ctx context.Context) error {
	if err := ws.validateAddr(); err != nil {
		return err
	}

	// If TLS is configured, validate files early so Start can fail fast.
	if ws.hasTLS() && !ws.validateTLSFiles() {
		return errors.New(errInvalidTLSFiles)
	}

	// Create and bind the listener before returning so callers know the server
	// is ready to accept connections. Use net.Listen to allocate the socket.
	ln, err := net.Listen("tcp", ws.srv.Addr)
	if err != nil {
		return err
	}

	// Run the server using the already-bound listener. run will take
	// ownership of the listener and close it during shutdown.
	go ws.run(ctx, nil, ln)
	return nil
}

// Listen blocks until ctx is canceled or the server exits unexpectedly.
func (ws *WebServer) Listen(ctx context.Context) error {
	if err := ws.validateAddr(); err != nil {
		return err
	}
	// Bind listener first so we can serve from it and still return errors
	// from the run goroutine.
	if ws.hasTLS() && !ws.validateTLSFiles() {
		return errors.New(errInvalidTLSFiles)
	}

	ln, err := net.Listen("tcp", ws.srv.Addr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	done := make(chan struct{})
	go func() { ws.run(ctx, errCh, ln); close(done) }()

	select {
	case err := <-errCh:
		<-done // ensure run finishes
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = ws.Stop(shutdownCtx)
		<-done
		return nil
	}
}

// Stop shuts down the HTTP server gracefully using the provided context.
func (ws *WebServer) Stop(ctx context.Context) error {
	return ws.srv.Shutdown(ctx)
}

// --- internals ---

func (ws *WebServer) validateAddr() error {
	_, err := net.ResolveTCPAddr("tcp", ws.srv.Addr)
	return err
}

func (ws *WebServer) hasTLS() bool {
	return ws.certFile != "" && ws.keyFile != ""
}

func (ws *WebServer) validateTLSFiles() bool {
	if _, err := os.Stat(ws.certFile); err != nil {
		slog.Error("TLS cert file not found", "path", ws.certFile, "error", err)
		return false
	}
	if _, err := os.Stat(ws.keyFile); err != nil {
		slog.Error("TLS key file not found", "path", ws.keyFile, "error", err)
		return false
	}
	return true
}

// getListenerFromContext attempts to retrieve a listener from the context (legacy support).
func (ws *WebServer) getListenerFromContext(ctx context.Context) net.Listener {
	if v := ctx.Value("listener"); v != nil {
		if l, ok := v.(net.Listener); ok {
			return l
		}
	}
	return nil
}

// startServerWithListener starts the server using a pre-bound listener.
func (ws *WebServer) startServerWithListener(ctx context.Context, ln net.Listener) error {
	if ws.hasTLS() {
		slog.InfoContext(ctx, "Starting HTTPS server (with listener)", "addr", ws.srv.Addr, "cert", ws.certFile, "key", ws.keyFile)
		return ws.srv.ServeTLS(ln, ws.certFile, ws.keyFile)
	}
	slog.InfoContext(ctx, "Starting HTTP server (with listener)", "addr", ws.srv.Addr)
	return ws.srv.Serve(ln)
}

// startServerWithoutListener starts the server by creating a new listener.
func (ws *WebServer) startServerWithoutListener(ctx context.Context) error {
	if ws.hasTLS() {
		if !ws.validateTLSFiles() {
			return errors.New(errInvalidTLSFiles)
		}
		slog.InfoContext(ctx, "Starting HTTPS server", "addr", ws.srv.Addr, "cert", ws.certFile, "key", ws.keyFile)
		return ws.srv.ListenAndServeTLS(ws.certFile, ws.keyFile)
	}
	slog.InfoContext(ctx, "Starting HTTP server", "addr", ws.srv.Addr)
	return ws.srv.ListenAndServe()
}

// handleServerShutdown gracefully shuts down the server when context is canceled.
func (ws *WebServer) handleServerShutdown(ctx context.Context, srvDone <-chan error) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ws.Stop(shutdownCtx); err != nil {
		slog.ErrorContext(ctx, "Error during server shutdown", "error", err)
	}
	// Ensure the server goroutine exits before returning
	<-srvDone
	slog.InfoContext(ctx, "Server shutdown complete")
}

// handleServerError processes server errors and sends them to errCh if available.
func (ws *WebServer) handleServerError(ctx context.Context, err error, errCh chan<- error) {
	if err != nil && err != http.ErrServerClosed {
		if errCh != nil {
			errCh <- err
		} else {
			slog.ErrorContext(ctx, "HTTP server error", "error", err)
		}
	}
}

// run manages the server lifecycle: start serving, log errors, shutdown on ctx cancel.
func (ws *WebServer) run(ctx context.Context, errCh chan<- error, ln net.Listener) {
	// Accept an optional listener so Start/Listen can create the bound
	// socket before calling run. If no listener was provided, try to
	// retrieve one from the context (legacy/private contract) as a
	// fallback.
	if ln == nil {
		ln = ws.getListenerFromContext(ctx)
	}

	srvDone := make(chan error, 1)

	// Start server in background
	go func() {
		var err error
		if ln != nil {
			err = ws.startServerWithListener(ctx, ln)
		} else {
			err = ws.startServerWithoutListener(ctx)
		}
		srvDone <- err
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		ws.handleServerShutdown(ctx, srvDone)
	case err := <-srvDone:
		ws.handleServerError(ctx, err, errCh)
	}
}
