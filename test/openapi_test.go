package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/test/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type publicOpenAPIUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type publicOpenAPIPage[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

func TestShouldGenerateOpenApiSpec(t *testing.T) {
	router := mux.NewRouter(mux.WithTitle("test title"), mux.WithDescription("test description"), mux.WithVersion("1.0.0"))
	testsupport.ConfigureRoutes(router)

	spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)

	require.NoError(t, err)
	require.NotNil(t, spec)

	expected := loadExpected(t)
	assert.Equal(t, expected.Normalize(), spec.Normalize())
}

func TestShouldNotMutateRouterMetadataWhenGeneratingOpenApiSpec(t *testing.T) {
	router := mux.NewRouter(mux.WithTitle("test title"), mux.WithDescription("test description"), mux.WithVersion("1.0.0"))
	page := publicOpenAPIPage[publicOpenAPIUser]{
		Items: []publicOpenAPIUser{{ID: 1, Name: "Alice"}},
		Total: 1,
	}

	router.POST("/users", func(c mux.RouteContext) {
		c.OK(page)
	}).
		WithOperationID("createUsers").
		WithJsonBody(page).
		WithOKResponse(page)

	defaultSpec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
	require.NoError(t, err)
	withExamplesSpec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(mux.WithOpenAPIExamples()), router)
	require.NoError(t, err)
	defaultSpecAgain, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
	require.NoError(t, err)

	defaultJSON := string(specJSONBytes(t, defaultSpec))
	withExamplesJSON := string(specJSONBytes(t, withExamplesSpec))
	defaultAgainJSON := string(specJSONBytes(t, defaultSpecAgain))

	assert.JSONEq(t, defaultJSON, defaultAgainJSON)
	assert.NotEqual(t, defaultJSON, withExamplesJSON)
	assert.NotContains(t, defaultJSON, `"Alice"`)
	assert.Contains(t, withExamplesJSON, `"Alice"`)
}

func TestGenerateSpecShouldRejectNilGeneratorOrRouter(t *testing.T) {
	router := mux.NewRouter(mux.WithTitle("Users API"), mux.WithVersion("1.0.0"))
	router.GET("/users", func(c mux.RouteContext) { c.NoContent() }).WithNoContentResponse()

	var generator *mux.Generator
	spec, err := mux.GenerateSpecWithGenerator(generator, router)
	require.Error(t, err)
	assert.Nil(t, spec)
	assert.ErrorContains(t, err, "generator is nil")

	spec, err = mux.GenerateSpecWithGenerator(mux.NewGenerator(), nil)
	require.Error(t, err)
	assert.Nil(t, spec)
	assert.ErrorContains(t, err, "router is nil")
}

func TestShouldPreserveTypedSecurityRequirements(t *testing.T) {
	router := mux.NewRouter(mux.WithTitle("Users API"), mux.WithVersion("1.0.0"))
	security := mux.SecurityRequirement{"oauth2": []string{"read", "write"}}

	router.GET("/users", func(c mux.RouteContext) {
		c.OK(map[string]string{"status": "ok"})
	}).
		WithOperationID("listUsers").
		WithSecurity(security).
		WithOKResponse(map[string]string{"status": "ok"})

	spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
	require.NoError(t, err)

	specMap := specJSONMap(t, spec)
	paths := requireMap(t, specMap["paths"])
	usersPath := requireMap(t, paths["/users"])
	getOp := requireMap(t, usersPath["get"])
	securityList := requireSlice(t, getOp["security"])
	require.Len(t, securityList, 1)

	firstRequirement := requireMap(t, securityList[0])
	scopes := requireSlice(t, firstRequirement["oauth2"])
	require.Equal(t, []any{"read", "write"}, scopes)
}

func TestShouldPreserveEmptySecurityScopes(t *testing.T) {
	router := mux.NewRouter(mux.WithTitle("Users API"), mux.WithVersion("1.0.0"))
	security := mux.SecurityRequirement{"oauth2": []string{}}

	router.GET("/users", func(c mux.RouteContext) {
		c.OK(map[string]string{"status": "ok"})
	}).
		WithOperationID("listUsers").
		WithSecurity(security).
		WithOKResponse(map[string]string{"status": "ok"})

	spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
	require.NoError(t, err)

	specMap := specJSONMap(t, spec)
	paths := requireMap(t, specMap["paths"])
	usersPath := requireMap(t, paths["/users"])
	getOp := requireMap(t, usersPath["get"])
	securityList := requireSlice(t, getOp["security"])
	require.Len(t, securityList, 1)

	firstRequirement := requireMap(t, securityList[0])
	scopes := requireSlice(t, firstRequirement["oauth2"])
	assert.Empty(t, scopes)
}

func TestShouldDocumentNamedResponseHelpers(t *testing.T) {
	router := mux.NewRouter(mux.WithTitle("Users API"), mux.WithVersion("1.0.0"))

	router.GET("/users", func(c mux.RouteContext) {
		c.OK(map[string]string{"status": "ok"})
	}).
		WithOperationID("listUsers").
		WithOKResponse(map[string]string{"status": "ok"}).
		WithBadRequestResponse().
		WithUnauthorizedResponse().
		WithForbiddenResponse().
		WithNotFoundResponse().
		WithConflictResponse().
		WithMovedPermanentlyResponse().
		WithFoundResponse().
		WithSeeOtherResponse().
		WithTemporaryRedirectResponse().
		WithPermanentRedirectResponse()

	spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
	require.NoError(t, err)

	specMap := specJSONMap(t, spec)
	paths := requireMap(t, specMap["paths"])
	usersPath := requireMap(t, paths["/users"])
	getOp := requireMap(t, usersPath["get"])
	responses := requireMap(t, getOp["responses"])

	for _, code := range []string{"200", "301", "302", "303", "307", "308", "400", "401", "403", "404", "409"} {
		_, ok := responses[code]
		require.Truef(t, ok, "expected response %s to be documented", code)
	}
}

func specJSONBytes(t *testing.T, spec *mux.OpenAPISpec) []byte {
	t.Helper()

	data, err := json.Marshal(spec)
	require.NoError(t, err)
	return data
}

func specJSONMap(t *testing.T, spec *mux.OpenAPISpec) map[string]any {
	t.Helper()

	var out map[string]any
	require.NoError(t, json.Unmarshal(specJSONBytes(t, spec), &out))
	return out
}

func requireMap(t *testing.T, value any) map[string]any {
	t.Helper()

	out, ok := value.(map[string]any)
	require.True(t, ok, "expected map[string]any, got %T", value)
	return out
}

func requireSlice(t *testing.T, value any) []any {
	t.Helper()

	out, ok := value.([]any)
	require.True(t, ok, "expected []any, got %T", value)
	return out
}

func loadExpected(t *testing.T) mux.OpenAPISpec {
	t.Helper()
	expectedPath := filepath.Join(".", "openapi.yaml")
	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	var expected mux.OpenAPISpec
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)
	return expected
}
