package test

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/fgrzl/mux"
)

var once sync.Once
var service *TestService

func StartTestServer(t *testing.T) (context.Context, *TestClient) {

	ctx := context.Background()
	_, cancel := context.WithCancel(context.Background())

	once.Do(func() {
		service = NewFakeService()
	})

	t.Cleanup(func() {
		cancel()
		// teardown test server
	})

	r := mux.NewRouter("/api/v1")

	// Add middleware
	// r.UseLogging(&mux.LoggingOptions{})
	// r.UseCompression(&mux.CompressionOptions{})
	// r.UseAuthentication(&mux.AuthenticationOptions{})
	// r.UseAuthorization(&mux.AuthorizationOptions{})

	// break up your routes
	ConfigureRoutes(r, service)

	addr := getAddr()

	go http.ListenAndServe(addr, r)

	return ctx, NewTestClient("http://" + addr)
}

func getAddr() string {
	// Create a listener on an ephemeral port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("Could not find available port")
	}
	defer listener.Close() // Ensure the listener is closed

	// Retrieve the assigned address and port
	addr := listener.Addr().(*net.TCPAddr)
	return addr.AddrPort().String()
}
