package test

import (
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
	// Arrange
	router := mux.NewRouter(mux.WithTitle("test title"), mux.WithDescription("test description"), mux.WithVersion("1.0.0"))
	testsupport.ConfigureRoutes(router)
	generator := mux.NewGenerator()

	// Act
	spec, err := mux.GenerateSpecWithGenerator(generator, router)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, spec)

	expected := loadExpected(t)
	assert.Equal(t, expected.Normalize(), spec.Normalize())
}

func TestShouldNotMutateRouterMetadataWhenGeneratingOpenApiSpec(t *testing.T) {
	// Arrange
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

	routesBefore, err := router.Routes()
	require.NoError(t, err)
	require.Len(t, routesBefore, 1)
	bodyMediaBefore := routesBefore[0].Options.RequestBody.Content[mux.MimeJSON]
	responseMediaBefore := routesBefore[0].Options.Responses["200"].Content[mux.MimeJSON]
	require.NotNil(t, bodyMediaBefore)
	require.NotNil(t, responseMediaBefore)
	requestRefBefore := bodyMediaBefore.Schema.Ref
	responseRefBefore := responseMediaBefore.Schema.Ref

	// Act
	defaultSpec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
	require.NoError(t, err)
	withExamplesSpec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(mux.WithExamples()), router)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, defaultSpec)
	require.NotNil(t, withExamplesSpec)

	routesAfter, err := router.Routes()
	require.NoError(t, err)
	require.Len(t, routesAfter, 1)
	bodyMediaAfter := routesAfter[0].Options.RequestBody.Content[mux.MimeJSON]
	responseAfter := routesAfter[0].Options.Responses["200"]
	responseMediaAfter := responseAfter.Content[mux.MimeJSON]
	require.NotNil(t, bodyMediaAfter)
	require.NotNil(t, responseMediaAfter)

	assert.Equal(t, requestRefBefore, bodyMediaAfter.Schema.Ref)
	assert.Equal(t, responseRefBefore, responseMediaAfter.Schema.Ref)
	assert.Equal(t, page, bodyMediaAfter.Example)
	assert.Equal(t, page, responseMediaAfter.Example)
	assert.Nil(t, bodyMediaAfter.Schema.Example)
	assert.Nil(t, responseMediaAfter.Schema.Example)
	assert.Equal(t, "", responseAfter.Description)

	generatedRequestMedia := defaultSpec.Paths["/users"].Post.RequestBody.Content[mux.MimeJSON]
	require.NotNil(t, generatedRequestMedia)
	assert.Nil(t, generatedRequestMedia.Example)
	assert.NotEmpty(t, withExamplesSpec.Paths["/users"].Post.RequestBody.Content[mux.MimeJSON].Schema.Ref)
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
