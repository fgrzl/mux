package binder

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseValueBySchema_UUID(t *testing.T) {
	u := uuid.New()
	s := u.String()
	schema := &openapi.Schema{Type: "string", Format: "uuid"}

	v, err := ParseValueBySchema([]string{s}, schema)
	assert.NoError(t, err)
	parsed, ok := v.(uuid.UUID)
	assert.True(t, ok)
	assert.Equal(t, u, parsed)

	// multiple values -> []uuid
	s2 := []string{u.String(), uuid.New().String()}
	v2, err := ParseValueBySchema(s2, schema)
	assert.NoError(t, err)
	arr, ok := v2.([]uuid.UUID)
	assert.True(t, ok)
	assert.Len(t, arr, 2)
}

func TestMakeConverter_Int(t *testing.T) {
	conv := makeConverter(reflect.TypeOf(int(0)), nil)
	assert.NotNil(t, conv)
	v, err := conv([]string{"42"})
	assert.NoError(t, err)
	// conv returns an int for single int value
	assert.Equal(t, 42, v.(int))
}

func TestParseByExample_PrefersExample(t *testing.T) {
	// Example provided as int should drive parsing
	param := &openapi.ParameterObject{Example: int(0)}
	v, ok := ParseByExample("7", param)
	assert.True(t, ok)
	assert.Equal(t, 7, v.(int))
}

func TestParseSliceValues_SchemaIntegerItems(t *testing.T) {
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "integer"}}}
	v, ok := ParseSliceValues([]string{"1", "2"}, param)
	assert.True(t, ok)
	// parseSliceValues returns []int64 for integer schema items
	arr, ok := v.([]int64)
	assert.True(t, ok)
	assert.EqualValues(t, []int64{1, 2}, arr)
}

func TestProcessParamAndSet_ConverterPrecedence(t *testing.T) {
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Converter: func(vals []string) (any, error) {
		return 100, nil
	}}
	handled, err := ProcessParamAndSet(staging, "k", []string{"x"}, "query", param)
	assert.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	assert.Equal(t, 100, v.(int))
}

func TestProcessParamAndSet_ArrayCSVSplit(t *testing.T) {
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}}}
	handled, err := ProcessParamAndSet(staging, "k", []string{"a,b"}, "query", param)
	assert.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	arr, ok := v.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arr)
}

func TestParseValueBySchema_InvalidUUID(t *testing.T) {
	schema := &openapi.Schema{Type: "string", Format: "uuid"}
	_, err := ParseValueBySchema([]string{"not-a-uuid"}, schema)
	assert.Error(t, err)
}

func TestParseValueBySchema_IntegerOverflow(t *testing.T) {
	// very large integer should fail to parse into int64
	schema := &openapi.Schema{Type: "integer"}
	_, err := ParseValueBySchema([]string{"9999999999999999999999999999"}, schema)
	assert.Error(t, err)
}

func TestProcessParamAndSet_ConverterErrorAndNil(t *testing.T) {
	// Converter returns error
	staging := make(map[string]any)
	paramErr := &openapi.ParameterObject{Converter: func(vals []string) (any, error) {
		return nil, fmt.Errorf("bad")
	}}
	handled, err := ProcessParamAndSet(staging, "k", []string{"x"}, "query", paramErr)
	assert.Error(t, err)
	assert.False(t, handled)

	// Converter returns (nil, nil) -> should not set and return false,nil
	staging2 := make(map[string]any)
	paramNil := &openapi.ParameterObject{Converter: func(vals []string) (any, error) {
		return nil, nil
	}}
	handled2, err2 := ProcessParamAndSet(staging2, "k", []string{"x"}, "query", paramNil)
	assert.NoError(t, err2)
	assert.False(t, handled2)
	_, ok := staging2["k"]
	assert.False(t, ok)
}

func TestParseSliceValues_NonIntegerFails(t *testing.T) {
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "integer"}}}
	v, ok := ParseSliceValues([]string{"1", "bad"}, param)
	assert.False(t, ok)
	assert.Nil(t, v)
}
