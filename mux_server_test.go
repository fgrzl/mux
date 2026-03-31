package mux

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fgrzl/mux/pkg/router"
	routing "github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAddrHTTP       = ":8080"
	testAddrHTTPS      = ":8443"
	testAddrLocal      = "127.0.0.1:0"
	testInvalidAddr    = "invalid:address:format"
	testCertFile       = "/path/to/cert.pem"
	testKeyFile        = "/path/to/key.pem"
	testCertName       = "localhost.crt"
	testKeyName        = "localhost.key"
	testCertsDir       = ".certs"
	testNonexistentDir = ".nonexistent"
	testReadTimeout    = 10 * time.Second
	testWriteTimeout   = 10 * time.Second
	testIdleTimeout    = 120 * time.Second
)

func TestShouldCreateNewServerWithDefaults(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()

	// Act
	server := NewServer(testAddrHTTP, rtr)

	// Assert
	assert.NotNil(t, server)
	assert.NotNil(t, server.srv)
	assert.Equal(t, testAddrHTTP, server.srv.Addr)
	assert.Equal(t, rtr, server.srv.Handler)
	assert.Equal(t, testReadTimeout, server.srv.ReadTimeout)
	assert.Equal(t, testWriteTimeout, server.srv.WriteTimeout)
	assert.Equal(t, testIdleTimeout, server.srv.IdleTimeout)
	assert.Equal(t, "", server.certFile)
	assert.Equal(t, "", server.keyFile)
}

func TestShouldConfigureServerWithTLS(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	// Act
	server := NewServer(testAddrHTTPS, rtr, WithTLS(testCertFile, testKeyFile))

	// Assert
	assert.NotNil(t, server)
	assert.Equal(t, testCertFile, server.certFile)
	assert.Equal(t, testKeyFile, server.keyFile)
}

func TestShouldConfigureServerWithTLSDiscovery(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	tempDir := t.TempDir()
	certsDir := filepath.Join(tempDir, testCertsDir)
	require.NoError(t, os.MkdirAll(certsDir, 0755))

	// Create dummy cert files
	certFile := filepath.Join(certsDir, testCertName)
	keyFile := filepath.Join(certsDir, testKeyName)
	require.NoError(t, os.WriteFile(certFile, []byte("dummy cert"), 0644))
	require.NoError(t, os.WriteFile(keyFile, []byte("dummy key"), 0644))

	// Change to temp directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// Act
	server := NewServer(testAddrHTTPS, rtr, WithTLSDiscovery(testCertsDir, testCertName, testKeyName))

	// Assert
	assert.NotNil(t, server)
	resolvedCertFile, _ := filepath.EvalSymlinks(certFile)
	resolvedServerCertFile, _ := filepath.EvalSymlinks(server.certFile)
	assert.Equal(t, resolvedCertFile, resolvedServerCertFile)

	resolvedKeyFile, _ := filepath.EvalSymlinks(keyFile)
	resolvedServerKeyFile, _ := filepath.EvalSymlinks(server.keyFile)
	assert.Equal(t, resolvedKeyFile, resolvedServerKeyFile)
}

func TestShouldHandleTLSDiscoveryWhenCertsDirNotFound(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	tempDir := t.TempDir()

	// Change to temp directory and ensure we change back
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() { require.NoError(t, os.Chdir(origDir)) }()

	// Act
	server := NewServer(testAddrHTTPS, rtr, WithTLSDiscovery(testNonexistentDir, testCertName, testKeyName))

	// Assert
	assert.NotNil(t, server)
	assert.Equal(t, "", server.certFile)
	assert.Equal(t, "", server.keyFile)
}

func TestShouldStartHTTPServerSuccessfully(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	rtr.GET("/health", func(c routing.RouteContext) {
		c.OK("OK")
	})

	server := NewServer(testAddrLocal, rtr) // Use 0 for random port
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	err := server.Start(ctx)

	// Assert
	require.NoError(t, err)

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cleanup
	cancel()
	time.Sleep(100 * time.Millisecond) // Give server time to shutdown
}

func TestShouldReturnErrorForInvalidAddress(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	server := NewServer(testInvalidAddr, rtr)
	ctx := context.Background()

	// Act
	err := server.Start(ctx)

	// Assert
	assert.Error(t, err)
}

func TestShouldStopServerGracefully(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	server := NewServer(testAddrLocal, rtr)
	ctx, cancel := context.WithCancel(context.Background())

	// Start server
	require.NoError(t, server.Start(ctx))
	time.Sleep(50 * time.Millisecond) // Give server time to start

	// Act
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	err := server.Stop(shutdownCtx)

	// Assert
	assert.NoError(t, err)

	// Cleanup
	cancel()
}

func TestShouldReturnFromListenWhenContextIsCancelled(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	server := NewServer(testAddrLocal, rtr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	// Act
	go func() {
		errCh <- server.Listen(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	// Assert
	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Listen did not return after context cancellation")
	}
}

func TestShouldShutdownServerOnlyOnceWhenListenContextIsCancelled(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	server := NewServer(testAddrLocal, rtr)
	var shutdownCalls atomic.Int32
	server.srv.RegisterOnShutdown(func() {
		shutdownCalls.Add(1)
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	// Act
	go func() {
		errCh <- server.Listen(ctx)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	// Assert
	select {
	case err := <-errCh:
		assert.NoError(t, err)
		assert.Eventually(t, func() bool {
			return shutdownCalls.Load() == 1
		}, time.Second, 10*time.Millisecond)
	case <-time.After(2 * time.Second):
		t.Fatal("Listen did not return after context cancellation")
	}
}

func TestShouldReturnErrorFromListenWhenPortIsAlreadyBound(t *testing.T) {
	// Arrange
	occupied, err := net.Listen("tcp", testAddrLocal)
	require.NoError(t, err)
	defer func() { require.NoError(t, occupied.Close()) }()

	rtr := router.NewRouter()
	server := NewServer(occupied.Addr().String(), rtr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Act
	err = server.Listen(ctx)

	// Assert
	assert.Error(t, err)
}

func TestShouldHandleMultipleOptions(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	// Act
	server := NewServer(testAddrHTTPS, rtr,
		WithTLS(testCertFile, testKeyFile),
		WithTLSDiscovery(testCertsDir, "other.crt", "other.key"), // This should be applied after WithTLS
	)

	// Assert - TLSDiscovery applies after WithTLS, so the original values should remain
	// since .certs directory won't exist, and it won't override existing values
	assert.NotNil(t, server)
	assert.Equal(t, testCertFile, server.certFile) // Should keep original value
	assert.Equal(t, testKeyFile, server.keyFile)   // Should keep original value
}
