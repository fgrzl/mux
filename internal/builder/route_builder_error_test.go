package builder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithOperationIDErrShouldReturnErrorForInvalidID(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x")

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
	rb := DetachedRoute(http.MethodGet, "/x")

	// Act
	result, err := rb.WithParamErr("p", "invalid", "", 1, true)

	// Assert
	require.Error(t, err)
	assert.Equal(t, rb, result)
	assert.Empty(t, rb.Options.Parameters)
	assert.Contains(t, err.Error(), "invalid parameter 'in'")
}

func TestWithPathParamErrShouldMarkRouteParameterRequired(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x/{id}")

	// Act
	result, err := rb.WithPathParamErr("id", "route identifier", "123")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, rb, result)
	require.Len(t, rb.Options.Parameters, 1)
	assert.Equal(t, "path", rb.Options.Parameters[0].In)
	assert.True(t, rb.Options.Parameters[0].Required)
	assert.Equal(t, "123", rb.Options.Parameters[0].Example)
	assert.NotNil(t, rb.Options.Parameters[0].Converter)
}

func TestWithQueryParamErrShouldAddOptionalRouteParameter(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x")

	// Act
	result, err := rb.WithQueryParamErr("limit", "page size", 10)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, rb, result)
	require.Len(t, rb.Options.Parameters, 1)
	assert.Equal(t, "query", rb.Options.Parameters[0].In)
	assert.False(t, rb.Options.Parameters[0].Required)
	assert.Equal(t, 10, rb.Options.Parameters[0].Example)
	assert.NotNil(t, rb.Options.ParamIndex)
	assert.NotNil(t, rb.Options.ParamIndex["query:limit"])
}

func TestWithCreatedResponseErrShouldAddCreatedResponse(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodPost, "/x")

	// Act
	result, err := rb.WithCreatedResponseErr(struct{ ID string }{ID: "123"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, rb, result)
	require.Contains(t, rb.Options.Responses, "201")
	assert.NotNil(t, rb.Options.Responses["201"].Content)
	assert.Contains(t, rb.Options.Responses["201"].Content, "application/json")
}

func TestWithStandardErrorsErrShouldAddCommonErrorResponses(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x")

	// Act
	result, err := rb.WithStandardErrorsErr()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, rb, result)
	assert.Contains(t, rb.Options.Responses, "400")
	assert.Contains(t, rb.Options.Responses, "404")
	assert.NotNil(t, rb.Options.Responses["400"].Content)
	assert.Nil(t, rb.Options.Responses["404"].Content)
}

func TestWithAuthResponsesErrShouldAddAuthResponses(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x")

	// Act
	result, err := rb.WithUnauthorizedResponseErr()
	require.NoError(t, err)
	assert.Equal(t, rb, result)

	result, err = rb.WithForbiddenResponseErr()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, rb, result)
	assert.Contains(t, rb.Options.Responses, "401")
	assert.Contains(t, rb.Options.Responses, "403")
	assert.Nil(t, rb.Options.Responses["401"].Content)
	assert.Nil(t, rb.Options.Responses["403"].Content)
}

func TestWithJsonBodyErrShouldReturnErrorForGet(t *testing.T) {
	// Arrange
	rb := DetachedRoute(http.MethodGet, "/x")

	// Act
	result, err := rb.WithJsonBodyErr(struct{ A int }{A: 1})

	// Assert
	require.Error(t, err)
	assert.Equal(t, rb, result)
	assert.Nil(t, rb.Options.RequestBody)
	assert.Contains(t, err.Error(), "HTTP method GET does not support a request body")
}
