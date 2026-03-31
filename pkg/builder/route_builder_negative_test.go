package builder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// OperationID invalid panic is already covered in route_builder_test.go; avoid duplicate

func TestShouldPanicOnInvalidParamIn(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x")

	// Act & Assert
	require.Panics(t, func() { rb.WithParam("p", "invalid", "", 1, true) })
}

func TestShouldPanicOnBodyForGet(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x")

	// Act & Assert
	require.PanicsWithValue(t,
		"HTTP method GET does not support a request body",
		func() { rb.WithJsonBody(struct{ A int }{A: 1}) },
	)
}
