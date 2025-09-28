package binder

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShouldParseUUIDSchemaAndMultipleUUIDs(t *testing.T) {
	// Arrange
	u := uuid.New()
	s := u.String()
	schema := &openapi.Schema{Type: "string", Format: "uuid"}

	// Act
	v, err := ParseValueBySchema([]string{s}, schema)
	assert.NoError(t, err)
	parsed, ok := v.(uuid.UUID)
	assert.True(t, ok)
	assert.Equal(t, u, parsed)

	// multiple values -> []uuid
	s2 := []string{u.String(), uuid.New().String()}
	// Act (multi)
	v2, err := ParseValueBySchema(s2, schema)
	assert.NoError(t, err)
	arr, ok := v2.([]uuid.UUID)
	assert.True(t, ok)
	assert.Len(t, arr, 2)
}

func TestShouldConvertSingleIntUsingMakeConverter(t *testing.T) {
	// Arrange
	conv := makeConverter(reflect.TypeOf(int(0)), nil)
	assert.NotNil(t, conv)
	// Act
	v, err := conv([]string{"42"})
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 42, v.(int))
}

func TestShouldConvertBoolAndFloatsUsingMakeConverter(t *testing.T) {
	// Arrange & Act
	cb := makeConverter(reflect.TypeOf(true), nil)
	v1, err1 := cb([]string{"true"})
	// Assert (bool)
	assert.NoError(t, err1)
	assert.Equal(t, true, v1.(bool))

	cf32 := makeConverter(reflect.TypeOf(float32(0)), nil)
	v2, err2 := cf32([]string{"1.5"})
	assert.NoError(t, err2)
	assert.InDelta(t, float32(1.5), v2.(float32), 1e-6)

	cf64 := makeConverter(reflect.TypeOf(float64(0)), nil)
	v3, err3 := cf64([]string{"2.75"})
	assert.NoError(t, err3)
	assert.InDelta(t, 2.75, v3.(float64), 1e-9)
}

func TestShouldPreferExampleOverSchemaWhenParsing(t *testing.T) {
	// Arrange: Example provided as int should drive parsing
	param := &openapi.ParameterObject{Example: int(0)}
	v, ok := ParseByExample("7", param)
	assert.True(t, ok)
	assert.Equal(t, 7, v.(int))
}

func TestShouldParseSliceValuesAsInt64GivenIntegerItemsSchema(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "integer"}}}
	// Act
	v, ok := ParseSliceValues([]string{"1", "2"}, param)
	// Assert
	assert.True(t, ok)
	arr, ok := v.([]int64)
	assert.True(t, ok)
	assert.EqualValues(t, []int64{1, 2}, arr)
}

func TestShouldApplyConverterBeforeOtherParsing(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Converter: func(vals []string) (any, error) { return 100, nil }}
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"x"}, "query", param)
	// Assert
	assert.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	assert.Equal(t, 100, v.(int))
}

func TestShouldSplitCSVIntoArrayForArrayParam(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}}}
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"a,b"}, "query", param)
	// Assert
	assert.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	arr, ok := v.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arr)
}

func TestShouldReturnErrorForInvalidUUIDSchemaValue(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "string", Format: "uuid"}
	// Act
	_, err := ParseValueBySchema([]string{"not-a-uuid"}, schema)
	// Assert
	assert.Error(t, err)
}

func TestShouldFailParsingWhenIntegerOverflowsInt64(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "integer"}
	// Act
	_, err := ParseValueBySchema([]string{"9999999999999999999999999999"}, schema)
	// Assert
	assert.Error(t, err)
}

func TestShouldHandleConverterErrorAndNilResult(t *testing.T) {
	// Arrange (error case)
	staging := make(map[string]any)
	paramErr := &openapi.ParameterObject{Converter: func(vals []string) (any, error) { return nil, fmt.Errorf("bad") }}
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"x"}, "query", paramErr)
	// Assert
	assert.Error(t, err)
	assert.False(t, handled)

	// Arrange (nil result case)
	staging2 := make(map[string]any)
	paramNil := &openapi.ParameterObject{Converter: func(vals []string) (any, error) { return nil, nil }}
	// Act
	handled2, err2 := ProcessParamAndSet(staging2, "k", []string{"x"}, "query", paramNil)
	// Assert
	assert.NoError(t, err2)
	assert.False(t, handled2)
	_, ok := staging2["k"]
	assert.False(t, ok)
}

func TestShouldReturnFalseParsingSliceWhenElementInvalid(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "integer"}}}
	// Act
	v, ok := ParseSliceValues([]string{"1", "bad"}, param)
	// Assert
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestShouldParseArraysNumbersAndBoolsBySchema(t *testing.T) {
	// Arrange & Act / Assert grouped per subtype
	// array of numbers
	numSchema := &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "number"}}
	v, err := ParseValueBySchema([]string{"1.5", "2.25"}, numSchema)
	assert.NoError(t, err)
	arr, ok := v.([]float64)
	assert.True(t, ok)
	assert.Len(t, arr, 2)

	// array of strings (pass-through)
	strSchema := &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}}
	v2, err2 := ParseValueBySchema([]string{"a", "b"}, strSchema)
	assert.NoError(t, err2)
	arr2, ok := v2.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arr2)

	// boolean
	boolSchema := &openapi.Schema{Type: "boolean"}
	v3, err3 := ParseValueBySchema([]string{"true"}, boolSchema)
	assert.NoError(t, err3)
	assert.Equal(t, true, v3.(bool))
}

func TestShouldNotHandleParamWhenNoConverterAndNoSchemaOrExample(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	param := &openapi.ParameterObject{}
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"raw"}, "query", param)
	// Assert
	assert.NoError(t, err)
	assert.False(t, handled)
}
