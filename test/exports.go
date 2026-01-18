package test

import (
	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/test/testsupport"
)

// MockServerHandler exports the mock server handler for use in benchmarks
func MockServerHandler() *mux.Router {
	return mockServerHandler()
}

func mockServerHandler() *mux.Router {
	r := mux.NewRouter()

	// Add middleware
	// r.UseLogging(&mux.LoggingOptions{})
	// r.UseCompression(&mux.CompressionOptions{})
	// r.UseAuthentication(&mux.AuthenticationOptions{})
	// r.UseAuthorization(&mux.AuthorizationOptions{})

	// break up your routes

	testsupport.ConfigureRoutes(r)

	return r
}
