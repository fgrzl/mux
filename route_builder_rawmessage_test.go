package mux

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuickSchemaShouldReturnByteSchemaForByteSlice(t *testing.T) {
	// Arrange
	tType := reflect.TypeOf([]byte{})

	// Act
	schema, err := quickSchema(tType)

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
	schema, err := quickSchema(tType)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	// empty schema indicates "any JSON"
	assert.Equal(t, &Schema{}, schema)
}
