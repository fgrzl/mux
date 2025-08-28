package builder

import (
	"encoding/json"
	"reflect"
	"testing"

	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/stretchr/testify/assert"
)

func TestQuickSchemaShouldReturnByteSchemaForByteSlice(t *testing.T) {
	// Arrange
	tType := reflect.TypeOf([]byte{})

	// Act
	schema, err := QuickSchema(tType)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "string", schema.Type)
	assert.Equal(t, "byte", schema.Format)
}

func TestQuickSchemaShouldReturnAnySchemaForRawMessage(t *testing.T) {
	// Arrange
	tType := reflect.TypeOf(json.RawMessage{})

	// Act
	schema, err := QuickSchema(tType)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	// empty schema indicates "any JSON"
	assert.Equal(t, &openapi.Schema{}, schema)
}
