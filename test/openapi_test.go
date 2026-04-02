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
