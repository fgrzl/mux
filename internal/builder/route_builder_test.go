package builder

import (
	"net"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fgrzl/mux/internal/common"
	openapi "github.com/fgrzl/mux/internal/openapi"
)

const (
	paramAccept = "Accept"
)

func TestShouldCreateRouteBuilder(t *testing.T) {
	// Arrange & Act
	builder := Route(http.MethodGet, "/users")

	// Assert
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.Options)
	assert.Equal(t, http.MethodGet, builder.Options.Method)
	assert.Equal(t, "/users", builder.Options.Pattern)
	assert.NotNil(t, builder.Options.Responses)
}

func TestShouldSetAllowAnonymous(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/public")

	// Act
	result := builder.AllowAnonymous()

	// Assert
	assert.True(t, builder.Options.AllowAnonymous)
	assert.Equal(t, builder, result) // Should return self for chaining
}

func TestShouldSetRequiredPermissions(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/secure")
	perms := []string{"read", "write"}

	// Act
	result := builder.RequirePermission(perms...)

	// Assert
	assert.Equal(t, perms, builder.Options.Permissions)
	assert.Equal(t, builder, result)
}

func TestShouldSetRequiredRoles(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/admin")
	roles := []string{"admin", "moderator"}

	// Act
	result := builder.RequireRoles(roles...)

	// Assert
	assert.Equal(t, roles, builder.Options.Roles)
	assert.Equal(t, builder, result)
}

func TestShouldSetRequiredScopes(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/api")
	scopes := []string{"api:read", "api:write"}

	// Act
	result := builder.RequireScopes(scopes...)

	// Assert
	assert.Equal(t, scopes, builder.Options.Scopes)
	assert.Equal(t, builder, result)
}

func TestShouldSetRateLimit(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/limited")
	limit := 100
	interval := time.Minute

	// Act
	result := builder.WithRateLimit(limit, interval)

	// Assert
	assert.Equal(t, limit, builder.Options.RateLimit)
	assert.Equal(t, interval, builder.Options.RateInterval)
	assert.Equal(t, builder, result)
}

func TestShouldSetOperationID(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	opID := "getUsers"

	// Act
	result := builder.WithOperationID(opID)

	// Assert
	assert.Equal(t, opID, builder.Options.OperationID)
	assert.Equal(t, builder, result)
}

func TestShouldPanicOnInvalidOperationID(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")

	// Act & Assert
	assert.Panics(t, func() {
		builder.WithOperationID("invalid-id")
	})
}

func TestShouldAddPathParameter(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users/{id}")
	example := "123"

	// Act
	result := builder.WithPathParam("id", example)

	// Assert
	assert.Equal(t, builder, result)
	require.Len(t, builder.Options.Parameters, 1)
	param := builder.Options.Parameters[0]
	assert.Equal(t, "id", param.Name)
	assert.Equal(t, "path", param.In)
	assert.True(t, param.Required)
	assert.NotNil(t, param.Schema)
}

func TestShouldAddQueryParameter(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	example := "10"

	// Act
	result := builder.WithQueryParam("limit", example)

	// Assert
	assert.Equal(t, builder, result)
	require.Len(t, builder.Options.Parameters, 1)
	param := builder.Options.Parameters[0]
	assert.Equal(t, "limit", param.Name)
	assert.Equal(t, "query", param.In)
	assert.False(t, param.Required)
}

func TestShouldAddRequiredQueryParameter(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	example := "active"

	// Act
	result := builder.WithRequiredQueryParam("status", example)

	// Assert
	assert.Equal(t, builder, result)
	require.Len(t, builder.Options.Parameters, 1)
	param := builder.Options.Parameters[0]
	assert.Equal(t, "status", param.Name)
	assert.Equal(t, "query", param.In)
	assert.True(t, param.Required)
}

func TestShouldAddHeaderParameter(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	example := common.MimeJSON

	// Act
	result := builder.WithHeaderParam(paramAccept, example, true)

	// Assert
	assert.Equal(t, builder, result)
	require.Len(t, builder.Options.Parameters, 1)
	param := builder.Options.Parameters[0]
	assert.Equal(t, "Accept", param.Name)
	assert.Equal(t, "header", param.In)
	assert.True(t, param.Required)
}

func TestShouldAddCookieParameter(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	example := "session123"

	// Act
	result := builder.WithCookieParam("sessionId", example, false)

	// Assert
	assert.Equal(t, builder, result)
	require.Len(t, builder.Options.Parameters, 1)
	param := builder.Options.Parameters[0]
	assert.Equal(t, "sessionId", param.Name)
	assert.Equal(t, "cookie", param.In)
	assert.False(t, param.Required)
}

func TestShouldPanicOnInvalidParameterIn(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")

	// Act & Assert
	assert.Panics(t, func() {
		builder.WithParam("test", "invalid", "example", false)
	})
}

func TestShouldPanicOnEmptyParameterName(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")

	// Act & Assert
	assert.Panics(t, func() {
		builder.WithParam("", "query", "example", false)
	})
}

func TestShouldAddResponse(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	example := struct {
		Name string `json:"name"`
	}{Name: "John"}

	// Act
	result := builder.WithResponse(200, example)

	// Assert
	assert.Equal(t, builder, result)
	response, exists := builder.Options.Responses["200"]
	assert.True(t, exists)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Content)
	mediaType := response.Content[common.MimeJSON]
	assert.NotNil(t, mediaType)
	assert.Equal(t, example, mediaType.Example)
}

func TestShouldAddStandardResponses(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	example := []string{"user1", "user2"}

	// Act
	builder.WithOKResponse(example).
		WithCreatedResponse(example).
		WithAcceptedResponse(example).
		WithNoContentResponse().
		WithNotFoundResponse().
		WithConflictResponse().
		WithBadRequestResponse().
		WithStandardErrors()

	// Assert
	assert.NotNil(t, builder.Options.Responses["200"])
	assert.NotNil(t, builder.Options.Responses["201"])
	assert.NotNil(t, builder.Options.Responses["202"])
	assert.NotNil(t, builder.Options.Responses["204"])
	assert.NotNil(t, builder.Options.Responses["404"])
	assert.NotNil(t, builder.Options.Responses["409"])
	assert.NotNil(t, builder.Options.Responses["400"])
}

func TestShouldAddJsonBody(t *testing.T) {
	// Arrange
	builder := Route(http.MethodPost, "/users")
	example := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{Name: "John", Email: "john@example.com"}

	// Act
	result := builder.WithJsonBody(example)

	// Assert
	assert.Equal(t, builder, result)
	assert.NotNil(t, builder.Options.RequestBody)
	assert.True(t, builder.Options.RequestBody.Required)
	mediaType := builder.Options.RequestBody.Content[common.MimeJSON]
	assert.NotNil(t, mediaType)
	assert.Equal(t, example, mediaType.Example)
}

func TestShouldAddFormBody(t *testing.T) {
	// Arrange
	builder := Route(http.MethodPost, "/users")
	example := struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}{Name: "John", Email: "john@example.com"}

	// Act
	result := builder.WithFormBody(example)

	// Assert
	assert.Equal(t, builder, result)
	assert.NotNil(t, builder.Options.RequestBody)
	mediaType := builder.Options.RequestBody.Content[common.MimeFormURLEncoded]
	assert.NotNil(t, mediaType)
	assert.Equal(t, example, mediaType.Example)
}

func TestShouldAddMultipartBody(t *testing.T) {
	// Arrange
	builder := Route(http.MethodPost, "/upload")
	example := struct {
		File string `json:"file"`
	}{File: "test.txt"}

	// Act
	result := builder.WithMultipartBody(example)

	// Assert
	assert.Equal(t, builder, result)
	assert.NotNil(t, builder.Options.RequestBody)
	mediaType := builder.Options.RequestBody.Content[common.MimeMultipartFormData]
	assert.NotNil(t, mediaType)
	assert.Equal(t, example, mediaType.Example)
}

func TestShouldSetSummary(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	summary := "Get all users"

	// Act
	result := builder.WithSummary(summary)

	// Assert
	assert.Equal(t, summary, builder.Options.Summary)
	assert.Equal(t, builder, result)
}

func TestShouldSetDescription(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	description := "Retrieves a list of all users in the system"

	// Act
	result := builder.WithDescription(description)

	// Assert
	assert.Equal(t, description, builder.Options.Description)
	assert.Equal(t, builder, result)
}

func TestShouldAddTags(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	tags := []string{"users", "admin"}

	// Act
	result := builder.WithTags(tags...)

	// Assert
	assert.Equal(t, tags, builder.Options.Tags)
	assert.Equal(t, builder, result)
}

func TestShouldSetExternalDocs(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	url := "https://api.example.com/docs"
	desc := "User API Documentation"

	// Act
	result := builder.WithExternalDocs(url, desc)

	// Assert
	assert.NotNil(t, builder.Options.ExternalDocs)
	assert.Equal(t, url, builder.Options.ExternalDocs.URL)
	assert.Equal(t, desc, builder.Options.ExternalDocs.Description)
	assert.Equal(t, builder, result)
}

func TestShouldAddSecurity(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/users")
	security := &openapi.SecurityRequirement{}

	// Act
	result := builder.WithSecurity(security)

	// Assert
	require.Len(t, builder.Options.Security, 1)
	assert.Equal(t, security, builder.Options.Security[0])
	assert.Equal(t, builder, result)
}

func TestShouldSetDeprecated(t *testing.T) {
	// Arrange
	builder := Route(http.MethodGet, "/old-endpoint")

	// Act
	result := builder.WithDeprecated()

	// Assert
	assert.True(t, builder.Options.Deprecated)
	assert.Equal(t, builder, result)
}

func TestQuickSchemaShouldHandleBasicTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		expectedType string
		expectedFmt  string
	}{
		{"string", "test", "string", ""},
		{"int", 42, "integer", ""},
		{"bool", true, "boolean", ""},
		{"float", 3.14, "number", ""},
		{"uuid", uuid.New(), "string", "uuid"},
		{"time", time.Now(), "string", "date-time"},
		{"bytes", []byte("test"), "string", "byte"},
		{"ip", net.IPv4(127, 0, 0, 1), "string", "ipv4"},
		{"url", url.URL{Scheme: "https", Host: "example.com"}, "string", "uri"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			schema, err := QuickSchema(reflect.TypeOf(tt.input))

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedType, schema.Type)
			if tt.expectedFmt != "" {
				assert.Equal(t, tt.expectedFmt, schema.Format)
			}
		})
	}
}

func TestQuickSchemaShouldHandleSlices(t *testing.T) {
	// Arrange
	input := []string{"a", "b", "c"}

	// Act
	schema, err := QuickSchema(reflect.TypeOf(input))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "array", schema.Type)
	assert.NotNil(t, schema.Items)
	assert.Equal(t, "string", schema.Items.Type)
}

func TestQuickSchemaShouldHandleNamedStruct(t *testing.T) {
	// Arrange
	type User struct {
		Name string
		Age  int
	}
	input := User{}

	// Act
	schema, err := QuickSchema(reflect.TypeOf(input))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "#/components/schemas/User", schema.Ref)
}

func TestQuickSchemaShouldHandleAnonymousStruct(t *testing.T) {
	// Arrange
	input := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{}

	// Act
	schema, err := QuickSchema(reflect.TypeOf(input))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "object", schema.Type)
	assert.NotNil(t, schema.Properties)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "age")
	assert.Equal(t, "string", schema.Properties["name"].Type)
	assert.Equal(t, "integer", schema.Properties["age"].Type)
}

func TestQuickSchemaShouldHandlePointer(t *testing.T) {
	// Arrange
	input := "test"
	ptr := &input

	// Act
	schema, err := QuickSchema(reflect.TypeOf(ptr))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "string", schema.Type)
}

func TestQuickSchemaShouldReturnErrorForNilType(t *testing.T) {
	// Act
	schema, err := QuickSchema(nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, schema)
	assert.Contains(t, err.Error(), "nil type")
}

func TestRegisterSchemaShouldAddCustomSchema(t *testing.T) {
	// Arrange
	type CustomType struct{}
	customSchema := &openapi.Schema{Type: "string", Format: "custom"}

	typ := reflect.TypeOf(CustomType{})
	RegisterSchema(typ, customSchema)
	t.Cleanup(func() {
		// Remove the custom schema from the registry to maintain test isolation
		removeSchema(typ)
	})

	// Assert
	schema, err := QuickSchema(typ)
	require.NoError(t, err)
	assert.Equal(t, customSchema, schema)
}

func TestShouldChainFluentMethods(t *testing.T) {
	// Arrange & Act
	builder := Route(http.MethodGet, "/users/{id}").
		AllowAnonymous().
		RequirePermission("read").
		RequireRoles("user").
		RequireScopes("api:read").
		WithRateLimit(100, time.Minute).
		WithOperationID("getUser").
		WithPathParam("id", "123").
		WithQueryParam("include", "profile").
		WithOKResponse(struct{ Name string }{Name: "John"}).
		WithSummary("Get user by ID").
		WithDescription("Retrieves a single user by their unique identifier").
		WithTags("users").
		WithDeprecated()

	// Assert
	assert.True(t, builder.Options.AllowAnonymous)
	assert.Contains(t, builder.Options.Permissions, "read")
	assert.Contains(t, builder.Options.Roles, "user")
	assert.Contains(t, builder.Options.Scopes, "api:read")
	assert.Equal(t, 100, builder.Options.RateLimit)
	assert.Equal(t, time.Minute, builder.Options.RateInterval)
	assert.Equal(t, "getUser", builder.Options.OperationID)
	assert.Len(t, builder.Options.Parameters, 2)
	assert.Contains(t, builder.Options.Responses, "200")
	assert.Equal(t, "Get user by ID", builder.Options.Summary)
	assert.Equal(t, "Retrieves a single user by their unique identifier", builder.Options.Description)
	assert.Contains(t, builder.Options.Tags, "users")
	assert.True(t, builder.Options.Deprecated)
}
