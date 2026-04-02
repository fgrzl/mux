package builder

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedirectResponsesInRouteBuilder verifies that redirect response builder methods
// properly add redirect responses to the route options
func TestRedirectResponsesInRouteBuilder(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/redirect-test").
		WithSummary("Test endpoint with all redirect types").
		WithDescription("This endpoint demonstrates all HTTP redirect status codes").
		With301Response().
		With302Response().
		With303Response().
		With307Response().
		With308Response().
		WithOKResponse(map[string]string{"message": "success"})

	// Act
	opts := builder.Options

	// Assert - Verify all redirect responses are present in the options
	assert.NotNil(t, opts.Responses["301"], "301 Moved Permanently should be documented")
	assert.NotNil(t, opts.Responses["302"], "302 Found should be documented")
	assert.NotNil(t, opts.Responses["303"], "303 See Other should be documented")
	assert.NotNil(t, opts.Responses["307"], "307 Temporary Redirect should be documented")
	assert.NotNil(t, opts.Responses["308"], "308 Permanent Redirect should be documented")
	assert.NotNil(t, opts.Responses["200"], "200 OK should be documented")

	// Verify redirect responses have no content (as expected for redirects)
	assert.Nil(t, opts.Responses["301"].Content, "301 should have no content")
	assert.Nil(t, opts.Responses["302"].Content, "302 should have no content")
	assert.Nil(t, opts.Responses["303"].Content, "303 should have no content")
	assert.Nil(t, opts.Responses["307"].Content, "307 should have no content")
	assert.Nil(t, opts.Responses["308"].Content, "308 should have no content")

	// Verify 200 response has content
	assert.NotNil(t, opts.Responses["200"].Content, "200 should have content")
}

// TestIndividualRedirectResponseBuilders tests each redirect builder method individually
func TestIndividualRedirectResponseBuilders(t *testing.T) {
	tests := []struct {
		name       string
		statusCode string
		method     func(*RouteBuilder) *RouteBuilder
	}{
		{"301 Moved Permanently", "301", func(rb *RouteBuilder) *RouteBuilder { return rb.With301Response() }},
		{"302 Found", "302", func(rb *RouteBuilder) *RouteBuilder { return rb.With302Response() }},
		{"303 See Other", "303", func(rb *RouteBuilder) *RouteBuilder { return rb.With303Response() }},
		{"307 Temporary Redirect", "307", func(rb *RouteBuilder) *RouteBuilder { return rb.With307Response() }},
		{"308 Permanent Redirect", "308", func(rb *RouteBuilder) *RouteBuilder { return rb.With308Response() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			builder := Route(http.MethodGet, "/test")
			builder = tt.method(builder)

			// Assert
			opts := builder.Options
			response := opts.Responses[tt.statusCode]
			assert.NotNil(t, response, "Response %s should be documented", tt.statusCode)
			assert.Nil(t, response.Content, "Redirect response %s should have no content", tt.statusCode)
		})
	}
}

// Test302ResponseForLoginEndpoint tests a realistic login redirect scenario
func Test302ResponseForLoginEndpoint(t *testing.T) {
	// Arrange & Act - A realistic login redirect scenario
	builder := Route(http.MethodGet, "/login").
		WithSummary("Login page").
		WithDescription("Redirects to authentication provider").
		With302Response().
		WithOperationID("login").
		WithTags("Authentication")

	opts := builder.Options

	// Assert
	assert.Equal(t, "login", opts.OperationID)
	assert.Equal(t, "Login page", opts.Summary)
	assert.Equal(t, "Redirects to authentication provider", opts.Description)
	assert.Contains(t, opts.Tags, "Authentication")

	response302 := opts.Responses["302"]
	assert.NotNil(t, response302, "302 response should be documented")
	assert.Nil(t, response302.Content, "302 should have no content body")
}

// TestPostRedirectGet303Pattern tests the POST-Redirect-GET pattern with 303
func TestPostRedirectGet303Pattern(t *testing.T) {
	// Arrange & Act - POST with 303 redirect (POST-Redirect-GET pattern)
	builder := Route(http.MethodPost, "/submit-form").
		WithSummary("Submit form data").
		WithDescription("Processes form submission and redirects to result page").
		WithJsonBody(map[string]string{"data": "example"}).
		With303Response(). // POST -> GET redirect
		WithBadRequestResponse()

	opts := builder.Options

	// Assert
	assert.NotNil(t, opts.RequestBody, "Should have request body")

	response303 := opts.Responses["303"]
	assert.NotNil(t, response303, "303 response should be documented for POST-Redirect-GET")
	assert.Nil(t, response303.Content, "303 should have no content")

	response400 := opts.Responses["400"]
	assert.NotNil(t, response400, "400 response should be documented")
}

// TestAPIVersionRedirect307And308 tests API versioning with method-preserving redirects
func TestAPIVersionRedirect307And308(t *testing.T) {
	// Test 307 - Temporary redirect, preserves method
	builderV1 := Route(http.MethodPost, "/api/v1/users").
		WithSummary("Create user (v1 - deprecated)").
		WithDescription("Redirects to v2 endpoint, preserving POST method").
		WithJsonBody(map[string]string{"name": "example"}).
		With307Response()

	optsV1 := builderV1.Options
	response307 := optsV1.Responses["307"]
	assert.NotNil(t, response307, "307 should be documented for temporary API redirect")
	assert.Nil(t, response307.Content, "307 should have no content")

	// Test 308 - Permanent redirect, preserves method
	builderV0 := Route(http.MethodPost, "/api/v0/users").
		WithSummary("Create user (v0 - obsolete)").
		WithDescription("Permanently redirects to v2 endpoint").
		WithJsonBody(map[string]string{"name": "example"}).
		With308Response()

	optsV0 := builderV0.Options
	response308 := optsV0.Responses["308"]
	assert.NotNil(t, response308, "308 should be documented for permanent API redirect")
	assert.Nil(t, response308.Content, "308 should have no content")
}

// TestRedirectResponseChaining verifies that redirect responses can be chained with other builder methods
func TestRedirectResponseChaining(t *testing.T) {
	// Test that all methods return *RouteBuilder for chaining
	builder := Route(http.MethodGet, "/chain-test").
		WithSummary("Chain test").
		With301Response().
		With302Response().
		With303Response().
		WithOKResponse(nil).
		WithCreatedResponse(nil).
		With307Response().
		With308Response().
		WithBadRequestResponse()

	// If we got here, chaining works
	assert.NotNil(t, builder)
	// We should have 8 responses: 301, 302, 303, 200, 201, 307, 308, 400
	assert.Equal(t, 8, len(builder.Options.Responses), "Should have 8 responses (301, 302, 303, 200, 201, 307, 308, 400)")
}

// TestNamedRedirectResponseBuilders tests the named versions of redirect builder methods
func TestNamedRedirectResponseBuilders(t *testing.T) {
	tests := []struct {
		name       string
		statusCode string
		method     func(*RouteBuilder) *RouteBuilder
	}{
		{"MovedPermanently", "301", func(rb *RouteBuilder) *RouteBuilder { return rb.WithMovedPermanentlyResponse() }},
		{"Found", "302", func(rb *RouteBuilder) *RouteBuilder { return rb.WithFoundResponse() }},
		{"SeeOther", "303", func(rb *RouteBuilder) *RouteBuilder { return rb.WithSeeOtherResponse() }},
		{"TemporaryRedirect", "307", func(rb *RouteBuilder) *RouteBuilder { return rb.WithTemporaryRedirectResponse() }},
		{"PermanentRedirect", "308", func(rb *RouteBuilder) *RouteBuilder { return rb.WithPermanentRedirectResponse() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			builder := Route(http.MethodGet, "/test")
			builder = tt.method(builder)

			// Assert
			opts := builder.Options
			response := opts.Responses[tt.statusCode]
			assert.NotNil(t, response, "Response %s should be documented using %s", tt.statusCode, tt.name)
			assert.Nil(t, response.Content, "Redirect response %s should have no content", tt.statusCode)
		})
	}
}

// TestNumericAndNamedRedirectMethodsEquivalent verifies both naming styles produce the same result
func TestNumericAndNamedRedirectMethodsEquivalent(t *testing.T) {
	tests := []struct {
		statusCode    string
		numericMethod func(*RouteBuilder) *RouteBuilder
		namedMethod   func(*RouteBuilder) *RouteBuilder
	}{
		{
			"301",
			func(rb *RouteBuilder) *RouteBuilder { return rb.With301Response() },
			func(rb *RouteBuilder) *RouteBuilder { return rb.WithMovedPermanentlyResponse() },
		},
		{
			"302",
			func(rb *RouteBuilder) *RouteBuilder { return rb.With302Response() },
			func(rb *RouteBuilder) *RouteBuilder { return rb.WithFoundResponse() },
		},
		{
			"303",
			func(rb *RouteBuilder) *RouteBuilder { return rb.With303Response() },
			func(rb *RouteBuilder) *RouteBuilder { return rb.WithSeeOtherResponse() },
		},
		{
			"307",
			func(rb *RouteBuilder) *RouteBuilder { return rb.With307Response() },
			func(rb *RouteBuilder) *RouteBuilder { return rb.WithTemporaryRedirectResponse() },
		},
		{
			"308",
			func(rb *RouteBuilder) *RouteBuilder { return rb.With308Response() },
			func(rb *RouteBuilder) *RouteBuilder { return rb.WithPermanentRedirectResponse() },
		},
	}

	for _, tt := range tests {
		t.Run("Status_"+tt.statusCode, func(t *testing.T) {
			// Build with numeric method
			numericBuilder := Route(http.MethodGet, "/test1")
			numericBuilder = tt.numericMethod(numericBuilder)

			// Build with named method
			namedBuilder := Route(http.MethodGet, "/test2")
			namedBuilder = tt.namedMethod(namedBuilder)

			// Both should have the same response
			assert.NotNil(t, numericBuilder.Options.Responses[tt.statusCode])
			assert.NotNil(t, namedBuilder.Options.Responses[tt.statusCode])
			assert.Equal(t,
				numericBuilder.Options.Responses[tt.statusCode],
				namedBuilder.Options.Responses[tt.statusCode],
				"Numeric and named methods should produce identical responses for status %s",
				tt.statusCode,
			)
		})
	}
}

// TestNamedRedirectMethodsInRealisticScenario tests using named methods in realistic scenarios
func TestNamedRedirectMethodsInRealisticScenario(t *testing.T) {
	// Use WithFoundResponse for a login redirect (most common case)
	loginBuilder := Route(http.MethodGet, "/login").
		WithSummary("Login redirect").
		WithFoundResponse()

	assert.NotNil(t, loginBuilder.Options.Responses["302"])

	// Use WithSeeOtherResponse for POST-Redirect-GET pattern
	submitBuilder := Route(http.MethodPost, "/form").
		WithSummary("Form submission").
		WithJsonBody(map[string]string{"data": "example"}).
		WithSeeOtherResponse().
		WithBadRequestResponse()

	assert.NotNil(t, submitBuilder.Options.Responses["303"])
	assert.NotNil(t, submitBuilder.Options.Responses["400"])

	// Use WithMovedPermanentlyResponse for SEO-friendly permanent redirects
	oldPageBuilder := Route(http.MethodGet, "/old-page").
		WithSummary("Old page (moved)").
		WithMovedPermanentlyResponse()

	assert.NotNil(t, oldPageBuilder.Options.Responses["301"])

	// Use WithTemporaryRedirectResponse for API versioning
	apiV1Builder := Route(http.MethodPost, "/api/v1/resource").
		WithSummary("API v1 (deprecated)").
		WithJsonBody(map[string]string{"id": "123"}).
		WithTemporaryRedirectResponse()

	assert.NotNil(t, apiV1Builder.Options.Responses["307"])

	// Use WithPermanentRedirectResponse for permanently moved API endpoints
	apiV0Builder := Route(http.MethodPost, "/api/v0/resource").
		WithSummary("API v0 (obsolete)").
		WithJsonBody(map[string]string{"id": "123"}).
		WithPermanentRedirectResponse()

	assert.NotNil(t, apiV0Builder.Options.Responses["308"])
}
