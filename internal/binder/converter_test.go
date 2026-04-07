package binder

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fgrzl/mux/internal/openapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldParseUUIDSchemaSingle(t *testing.T) {
	// Arrange
	u := uuid.New()
	s := u.String()
	schema := &openapi.Schema{Type: "string", Format: "uuid"}

	// Act
	v, err := ParseValueBySchema([]string{s}, schema)
	// Assert
	assert.NoError(t, err)
	parsed, ok := v.(uuid.UUID)
	assert.True(t, ok)
	assert.Equal(t, u, parsed)
}

func TestShouldParseUUIDSchemaMultiple(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "string", Format: "uuid"}
	s2 := []string{uuid.New().String(), uuid.New().String()}
	// Act
	v2, err := ParseValueBySchema(s2, schema)
	// Assert
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
	require.NoError(t, err)
	assert.Equal(t, 42, v.(int))
}

func TestShouldConvertPointerTextUnmarshalerUsingMakeConverter(t *testing.T) {
	// Arrange
	conv := makeConverter(reflect.TypeOf(new(uuid.UUID)), nil)
	require.NotNil(t, conv)
	want := uuid.New()

	// Act
	v, err := conv([]string{want.String()})

	// Assert
	require.NoError(t, err)
	parsed, ok := v.(*uuid.UUID)
	require.True(t, ok, "expected *uuid.UUID, got %T", v)
	require.NotNil(t, parsed)
	assert.Equal(t, want, *parsed)
}

func TestShouldConvertBoolUsingMakeConverter(t *testing.T) {
	// Arrange
	cb := makeConverter(reflect.TypeOf(true), nil)
	require.NotNil(t, cb)
	// Act
	v1, err := cb([]string{"true"})
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, true, v1.(bool))
}

func TestShouldConvertFloat32UsingMakeConverter(t *testing.T) {
	// Arrange
	cf32 := makeConverter(reflect.TypeOf(float32(0)), nil)
	require.NotNil(t, cf32)
	// Act
	v, err := cf32([]string{"1.5"})
	// Assert
	assert.NoError(t, err)
	assert.InDelta(t, float32(1.5), v.(float32), 1e-6)
}

func TestShouldConvertFloat64UsingMakeConverter(t *testing.T) {
	// Arrange
	cf64 := makeConverter(reflect.TypeOf(float64(0)), nil)
	require.NotNil(t, cf64)
	// Act
	v, err := cf64([]string{"2.75"})
	// Assert
	assert.NoError(t, err)
	assert.InDelta(t, 2.75, v.(float64), 1e-9)
}

func TestShouldPreferExampleOverSchemaWhenParsing(t *testing.T) {
	// Arrange: Example provided as int should drive parsing
	param := &openapi.ParameterObject{Example: int(0)}
	v, ok := ParseByExample("7", param)
	assert.True(t, ok)
	assert.Equal(t, 7, v.(int))
}

func TestShouldParseUUIDByExample(t *testing.T) {
	// Arrange
	want := uuid.New()
	param := &openapi.ParameterObject{Example: uuid.UUID{}}

	// Act
	v, ok := ParseByExample(want.String(), param)

	// Assert
	assert.True(t, ok)
	assert.Equal(t, want, v)
}

func TestShouldParsePointerUUIDByExample(t *testing.T) {
	// Arrange
	want := uuid.New()
	param := &openapi.ParameterObject{Example: new(uuid.UUID)}

	// Act
	v, ok := ParseByExample(want.String(), param)

	// Assert
	require.True(t, ok)
	parsed, ok := v.(*uuid.UUID)
	require.True(t, ok, "expected *uuid.UUID, got %T", v)
	require.NotNil(t, parsed)
	assert.Equal(t, want, *parsed)
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

func TestShouldParsePointerTextUnmarshalerSlicesFromExample(t *testing.T) {
	// Arrange
	want := []uuid.UUID{uuid.New(), uuid.New()}
	param := &openapi.ParameterObject{Example: []*uuid.UUID{}}

	// Act
	v, ok := ParseSliceValues([]string{want[0].String(), want[1].String()}, param)

	// Assert
	require.True(t, ok)
	parsed, ok := v.([]*uuid.UUID)
	require.True(t, ok, "expected []*uuid.UUID, got %T", v)
	require.Len(t, parsed, len(want))
	for i := range want {
		require.NotNil(t, parsed[i])
		assert.Equal(t, want[i], *parsed[i])
	}
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

func TestShouldParseArrayNumbersBySchema(t *testing.T) {
	// Arrange
	numSchema := &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "number"}}
	// Act
	v, err := ParseValueBySchema([]string{"1.5", "2.25"}, numSchema)
	// Assert
	assert.NoError(t, err)
	arr, ok := v.([]float64)
	assert.True(t, ok)
	assert.Len(t, arr, 2)
}

func TestShouldParseArrayStringsBySchema(t *testing.T) {
	// Arrange
	strSchema := &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}}
	// Act
	v2, err2 := ParseValueBySchema([]string{"a", "b"}, strSchema)
	// Assert
	assert.NoError(t, err2)
	arr2, ok := v2.([]string)
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, arr2)
}

func TestShouldParseBooleanBySchema(t *testing.T) {
	// Arrange
	boolSchema := &openapi.Schema{Type: "boolean"}
	// Act
	v3, err3 := ParseValueBySchema([]string{"true"}, boolSchema)
	// Assert
	assert.NoError(t, err3)
	assert.Equal(t, true, v3.(bool))
}

func TestShouldUseMakeConverterWrapper(t *testing.T) {
	// Arrange
	conv := MakeConverter(reflect.TypeOf(int(0)), nil)
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"5"})
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 5, v.(int))
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

func TestShouldConvertMultiBoolUsingMakeConverter(t *testing.T) {
	// Arrange
	cb := makeConverter(reflect.TypeOf(true), nil)
	require.NotNil(t, cb)
	// Act
	v, err := cb([]string{"true", "false"})
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, []bool{true, false}, v)
}

func TestShouldConvertMultiFloat32UsingMakeConverter(t *testing.T) {
	cf32 := makeConverter(reflect.TypeOf(float32(0)), nil)
	require.NotNil(t, cf32)
	v, err := cf32([]string{"1.25", "2.5"})
	assert.NoError(t, err)
	assert.Equal(t, []float64{1.25, 2.5}, v)
}

func TestShouldConvertMultiFloat64UsingMakeConverter(t *testing.T) {
	cf64 := makeConverter(reflect.TypeOf(float64(0)), nil)
	require.NotNil(t, cf64)
	v, err := cf64([]string{"1.25", "2.5"})
	assert.NoError(t, err)
	assert.Equal(t, []float64{1.25, 2.5}, v)
}

func TestShouldReturnStringOrSliceForSchemaString(t *testing.T) {
	// single
	conv := makeConverter(nil, &openapi.Schema{Type: "string"})
	require.NotNil(t, conv)
	v, err := conv([]string{"x"})
	assert.NoError(t, err)
	assert.Equal(t, "x", v)
	// multi
	vm, err := conv([]string{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, vm)
}

func TestShouldReturnNilConverterForUnsupportedType(t *testing.T) {
	type custom struct{ X int }
	conv := makeConverter(reflect.TypeOf(custom{}), nil)
	assert.Nil(t, conv)
}

func TestShouldParseByExampleUsingSchemaOnly(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Type: "integer"}}
	// Act
	v, ok := ParseByExample("42", param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, int64(42), v)
}

func TestShouldReturnFalseParseSliceValuesWhenParamNil(t *testing.T) {
	v, ok := ParseSliceValues([]string{"1"}, nil)
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestShouldConvertVariousSignedIntsSingle(t *testing.T) {
	cases := []struct {
		name string
		t    reflect.Type
		want any
	}{
		{"int8", reflect.TypeOf(int8(0)), int8(12)},
		{"int16", reflect.TypeOf(int16(0)), int16(12)},
		{"int32", reflect.TypeOf(int32(0)), int32(12)},
		{"int64", reflect.TypeOf(int64(0)), int64(12)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conv := makeConverter(tc.t, nil)
			require.NotNil(t, conv)
			v, err := conv([]string{"12"})
			require.NoError(t, err)
			assert.Equal(t, tc.want, v)
		})
	}
}

func TestShouldConvertVariousUnsignedIntsSingle(t *testing.T) {
	cases := []struct {
		name string
		t    reflect.Type
		want any
	}{
		{"uint8", reflect.TypeOf(uint8(0)), uint8(9)},
		{"uint16", reflect.TypeOf(uint16(0)), uint16(9)},
		{"uint32", reflect.TypeOf(uint32(0)), uint32(9)},
		{"uint64", reflect.TypeOf(uint64(0)), uint64(9)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conv := makeConverter(tc.t, nil)
			require.NotNil(t, conv)
			v, err := conv([]string{"9"})
			require.NoError(t, err)
			assert.Equal(t, tc.want, v)
		})
	}
}

func TestShouldConvertSignedIntMultiUsingMakeConverter(t *testing.T) {
	// Arrange
	conv := makeConverter(reflect.TypeOf(int(0)), nil)
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"1", "2", "3"})
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, v)
}

func TestShouldErrorConvertingSliceElementsWhenOneInvalid(t *testing.T) {
	// Arrange
	conv := makeConverter(reflect.TypeOf([]int{}), nil)
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"1", "bad"})
	// Assert
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldErrorOnBoolSliceWithInvalidElement(t *testing.T) {
	conv := makeConverter(reflect.TypeOf(true), nil)
	require.NotNil(t, conv)
	v, err := conv([]string{"true", "bad"})
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldErrorOnFloatSliceWithInvalidElement(t *testing.T) {
	conv := makeConverter(reflect.TypeOf(float64(0)), nil)
	require.NotNil(t, conv)
	v, err := conv([]string{"1.0", "oops"})
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldReturnNilConverterForUnsupportedSliceElemType(t *testing.T) {
	type custom struct{ X int }
	conv := makeConverter(reflect.TypeOf([]custom{}), nil)
	assert.Nil(t, conv)
}

func TestShouldSupportPointerToTypeForMakeConverter(t *testing.T) {
	var p *int
	conv := makeConverter(reflect.TypeOf(p), nil)
	require.NotNil(t, conv)
	v, err := conv([]string{"7"})
	assert.NoError(t, err)
	assert.Equal(t, 7, v)
}

func TestShouldMakeConverterSchemaIntegerSingleAndError(t *testing.T) {
	conv := makeConverter(nil, &openapi.Schema{Type: "integer"})
	require.NotNil(t, conv)
	// success
	v, err := conv([]string{"42"})
	assert.NoError(t, err)
	assert.Equal(t, int64(42), v)
	// error
	v2, err2 := conv([]string{"x"})
	assert.Error(t, err2)
	assert.Nil(t, v2)
}

func TestShouldMakeConverterSchemaNumberSingleAndError(t *testing.T) {
	conv := makeConverter(nil, &openapi.Schema{Type: "number"})
	require.NotNil(t, conv)
	v, err := conv([]string{"1.25"})
	assert.NoError(t, err)
	assert.Equal(t, 1.25, v)
	v2, err2 := conv([]string{"nope"})
	assert.Error(t, err2)
	assert.Nil(t, v2)
}

func TestShouldMakeConverterSchemaBooleanSingleAndError(t *testing.T) {
	conv := makeConverter(nil, &openapi.Schema{Type: "boolean"})
	require.NotNil(t, conv)
	v, err := conv([]string{"true"})
	assert.NoError(t, err)
	assert.Equal(t, true, v)
	v2, err2 := conv([]string{"not-bool"})
	assert.Error(t, err2)
	assert.Nil(t, v2)
}

func TestShouldMakeConverterSchemaUUIDSingle(t *testing.T) {
	conv := makeConverter(nil, &openapi.Schema{Type: "string", Format: "uuid"})
	require.NotNil(t, conv)
	id := uuid.New().String()
	v, err := conv([]string{id})
	assert.NoError(t, err)
	parsed := v.(uuid.UUID)
	assert.Equal(t, id, parsed.String())
}

func TestShouldParseByExampleFallbackToSchemaWhenExampleFails(t *testing.T) {
	// Example expects int but value invalid; schema is string and will parse fine
	param := &openapi.ParameterObject{Example: int(0), Schema: &openapi.Schema{Type: "string"}}
	v, ok := ParseByExample("abc", param)
	assert.True(t, ok)
	assert.Equal(t, "abc", v)
}
