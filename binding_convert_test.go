package mux

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseValueBySchema_UUID(t *testing.T) {
	u := uuid.New()
	s := u.String()
	schema := &Schema{Type: "string", Format: "uuid"}

	v, err := parseValueBySchema([]string{s}, schema)
	assert.NoError(t, err)
	parsed, ok := v.(uuid.UUID)
	assert.True(t, ok)
	assert.Equal(t, u, parsed)

	// multiple values -> []uuid
	s2 := []string{u.String(), uuid.New().String()}
	v2, err := parseValueBySchema(s2, schema)
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
	param := &ParameterObject{Example: int(0)}
	v, ok := parseByExample("7", param)
	assert.True(t, ok)
	assert.Equal(t, 7, v.(int))
}

func TestParseSliceValues_SchemaIntegerItems(t *testing.T) {
	param := &ParameterObject{Schema: &Schema{Items: &Schema{Type: "integer"}}}
	v, ok := parseSliceValues([]string{"1", "2"}, param)
	assert.True(t, ok)
	// parseSliceValues returns []int64 for integer schema items
	arr, ok := v.([]int64)
	assert.True(t, ok)
	assert.EqualValues(t, []int64{1, 2}, arr)
}

func TestProcessParamAndSet_ConverterPrecedence(t *testing.T) {
	staging := make(map[string]any)
	param := &ParameterObject{Converter: func(vals []string) (any, error) {
		return 100, nil
	}}
	handled, err := processParamAndSet(staging, "k", []string{"x"}, "query", param)
	assert.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	assert.Equal(t, 100, v.(int))
}

func TestProcessParamAndSet_ArrayCSVSplit(t *testing.T) {
	staging := make(map[string]any)
	param := &ParameterObject{Schema: &Schema{Type: "array", Items: &Schema{Type: "string"}}}
	handled, err := processParamAndSet(staging, "k", []string{"a,b"}, "query", param)
	assert.NoError(t, err)
	assert.True(t, handled)
	v, ok := staging["k"]
	assert.True(t, ok)
	arr, ok := v.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arr)
}

func TestParseValueBySchema_InvalidUUID(t *testing.T) {
	schema := &Schema{Type: "string", Format: "uuid"}
	_, err := parseValueBySchema([]string{"not-a-uuid"}, schema)
	assert.Error(t, err)
}

func TestParseValueBySchema_IntegerOverflow(t *testing.T) {
	// very large integer should fail to parse into int64
	schema := &Schema{Type: "integer"}
	_, err := parseValueBySchema([]string{"9999999999999999999999999999"}, schema)
	assert.Error(t, err)
}

func TestProcessParamAndSet_ConverterErrorAndNil(t *testing.T) {
	// Converter returns error
	staging := make(map[string]any)
	paramErr := &ParameterObject{Converter: func(vals []string) (any, error) {
		return nil, fmt.Errorf("bad")
	}}
	handled, err := processParamAndSet(staging, "k", []string{"x"}, "query", paramErr)
	assert.Error(t, err)
	assert.False(t, handled)

	// Converter returns (nil, nil) -> should not set and return false,nil
	staging2 := make(map[string]any)
	paramNil := &ParameterObject{Converter: func(vals []string) (any, error) {
		return nil, nil
	}}
	handled2, err2 := processParamAndSet(staging2, "k", []string{"x"}, "query", paramNil)
	assert.NoError(t, err2)
	assert.False(t, handled2)
	_, ok := staging2["k"]
	assert.False(t, ok)
}

func TestParseSliceValues_NonIntegerFails(t *testing.T) {
	param := &ParameterObject{Schema: &Schema{Items: &Schema{Type: "integer"}}}
	v, ok := parseSliceValues([]string{"1", "bad"}, param)
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestSetNestedMap_DeepNesting(t *testing.T) {
	staging := make(map[string]any)
	setNestedMap(staging, "user", []string{"address", "city"}, "NYC")
	u, ok := staging["user"].(map[string]any)
	assert.True(t, ok)
	addr, ok := u["address"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "NYC", addr["city"])
}
