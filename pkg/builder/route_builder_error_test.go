package builder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithOperationIDErrShouldReturnErrorForInvalidID(t *testing.T) {
	// Arrange
	rb := Route(http.MethodGet, "/x")

	// Act
	result, err := rb.WithOperationIDErr("invalid-id")

	// Assert
	require.Error(t, err)
	assert.Equal(t, rb, result)
	assert.Empty(t, rb.Options.OperationID)
	assert.Contains(t, err.Error(), "invalid OperationID")
}

func TestWithParamErrShouldReturnErrorForInvalidLocation(t *testing.T) {
	// Arrange
	rb := Route(http.MethodGet, "/x")

	// Act
	result, err := rb.WithParamErr("p", "invalid", "", 1, true)

	// Assert
	require.Error(t, err)
	assert.Equal(t, rb, result)
	assert.Empty(t, rb.Options.Parameters)
	assert.Contains(t, err.Error(), "invalid parameter 'in'")
}

func TestWithJsonBodyErrShouldReturnErrorForGet(t *testing.T) {
	// Arrange
	rb := Route(http.MethodGet, "/x")

	// Act
	result, err := rb.WithJsonBodyErr(struct{ A int }{A: 1})

	// Assert
	require.Error(t, err)
	assert.Equal(t, rb, result)
	assert.Nil(t, rb.Options.RequestBody)
	assert.Contains(t, err.Error(), "HTTP method GET does not support a request body")
}
