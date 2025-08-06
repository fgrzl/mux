package mux

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// WebServerOption allows functional options for WebServer.
type WebServerOption func(*WebServer)

// WebServer wraps an http.Server and your custom Router.
type WebServer struct {
	srv      *http.Server
	router   *Router
	certFile string
	keyFile  string
}

func WithTLS(certFile, keyFile string) WebServerOption {
	return func(ws *WebServer) {
		ws.certFile = certFile
		ws.keyFile = keyFile
	}
}

// WithTLSDiscovery searches for a certs directory by walking up the directory tree (up to 10 levels),
// and sets the certFile and keyFile fields on the WebServer to the discovered paths.
// This allows for flexible certificate management in local development or deployment environments.
//
// certsDir: the directory name to search for (e.g., ".certs")
// certFile: the certificate file name (e.g., "localhost.crt")
// keyFile:  the key file name (e.g., "localhost.key")
//
// If the certs directory is not found, an error is logged and TLS will not be enabled.
func WithTLSDiscovery(certsDir, certFile, keyFile string) WebServerOption {
	return func(ws *WebServer) {
		dir, err := os.Getwd()
		if err != nil {
			slog.Error("Could not get working directory for TLS discovery", "error", err)
			return
		}
		found := false
		for i := 0; i < 10; i++ {
			certsPath := filepath.Join(dir, certsDir)
			if stat, err := os.Stat(certsPath); err == nil && stat.IsDir() {
				ws.certFile = filepath.Join(certsPath, certFile)
				ws.keyFile = filepath.Join(certsPath, keyFile)
				found = true
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
		if !found {
			slog.Error("Could not find certs directory for TLS discovery", "searched_from", dir)
		}
	}

}

// NewServer sets up the HTTP server with sane timeouts and a mux Router.
func NewServer(addr string, router *Router, opts ...WebServerOption) *WebServer {
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	ws := &WebServer{srv: srv, router: router}
	for _, opt := range opts {
		opt(ws)
	}
	return ws
}

// Start runs the HTTP or HTTPS server in a goroutine and handles graceful shutdown.
func (ws *WebServer) Start(ctx context.Context) error {
	// Validate address (optional)
	if _, err := net.ResolveTCPAddr("tcp", ws.srv.Addr); err != nil {
		return err
	}

	go func() {
		var err error
		if ws.certFile != "" && ws.keyFile != "" {
			// Check cert and key file existence before starting TLS
			if _, errCert := os.Stat(ws.certFile); errCert != nil {
				slog.Error("TLS cert file not found", "path", ws.certFile, "error", errCert)
				return
			}
			if _, errKey := os.Stat(ws.keyFile); errKey != nil {
				slog.Error("TLS key file not found", "path", ws.keyFile, "error", errKey)
				return
			}
			slog.Info("Starting HTTPS server", "addr", ws.srv.Addr, "cert", ws.certFile, "key", ws.keyFile)
			err = ws.srv.ListenAndServeTLS(ws.certFile, ws.keyFile)
		} else {
			slog.Info("Starting HTTP server", "addr", ws.srv.Addr)
			err = ws.srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := ws.Stop(shutdownCtx); err != nil {
			slog.Error("Error during server shutdown", "error", err)
		} else {
			slog.Info("Server shutdown complete")
		}
	}()

	return nil
}

// Stop shuts down the HTTP server gracefully.
func (ws *WebServer) Stop(ctx context.Context) error {
	return ws.srv.Shutdown(ctx)
}
