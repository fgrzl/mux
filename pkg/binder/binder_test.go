package binder

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: These tests focus on edge branches not covered by existing tests to raise coverage.

func TestShouldConvertUnsignedIntegersSingleAndMulti(t *testing.T) {
	// Arrange
	types := []reflect.Type{reflect.TypeOf(uint(0)), reflect.TypeOf(uint8(0)), reflect.TypeOf(uint16(0)), reflect.TypeOf(uint32(0)), reflect.TypeOf(uint64(0))}
	values := []string{"1", "2", "3"}

	for _, tt := range types {
		t.Run(tt.String(), func(t *testing.T) {
			// Arrange
			conv := makeConverter(tt, nil)
			require.NotNil(t, conv)
			// Act (single)
			single, err := conv(values[:1])
			// Assert
			require.NoError(t, err)
			assert.NotNil(t, single)
			// Act (multi)
			multi, err := conv(values)
			// Assert
			require.NoError(t, err)
			assert.IsType(t, []uint64{}, multi)
		})
	}
}

func TestShouldReturnErrorOnUnsignedConvInvalidInput(t *testing.T) {
	// Arrange
	conv := makeConverter(reflect.TypeOf(uint(0)), nil)
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"abc"})
	// Assert
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldConvertSliceElementTypesUsingExampleSlice(t *testing.T) {
	// Arrange
	example := []int{0}
	conv := makeConverter(reflect.TypeOf(example), nil)
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"1", "2"})
	// Assert
	require.NoError(t, err)
	// Expect slice of interface{} with parsed ints
	arr := v.([]any)
	require.Len(t, arr, 2)
	assert.Equal(t, 1, arr[0])
	assert.Equal(t, 2, arr[1])
}

func TestShouldReturnNilFromSchemaScalarConverterForMultiValues(t *testing.T) {
	// Arrange: schema-backed integer converter should return nil for multi-values per implementation
	schema := &openapi.Schema{Type: "integer"}
	conv := makeConverter(nil, schema)
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"1", "2"})
	// Assert
	require.NoError(t, err)
	assert.Nil(t, v)
}

func TestShouldParseMultipleUUIDsBySchema(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "string", Format: "uuid"}
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	raw := []string{ids[0].String(), ids[1].String()}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	require.NoError(t, err)
	parsed := v.([]uuid.UUID)
	assert.Equal(t, ids, parsed)
}

func TestShouldPreferExampleOverSchemaTypeMismatch(t *testing.T) {
	// Arrange: Example is string, Schema is integer, should parse as string
	param := &openapi.ParameterObject{Example: "abc", Schema: &openapi.Schema{Type: "integer"}}
	// Act
	v, ok := parseByExample("abc", param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, "abc", v)
}

func TestShouldReturnFalseWhenParseByExampleParamNil(t *testing.T) {
	// Arrange / Act
	v, ok := parseByExample("x", nil)
	// Assert
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestShouldFailParseSliceValuesWhenExampleIntSliceElementInvalid(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Example: []int{0}}
	// Act
	v, ok := parseSliceValues([]string{"1", "bad"}, param)
	// Assert
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestShouldParseNumberSliceFromSchema(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "number"}}}
	// Act
	v, ok := parseSliceValues([]string{"1.5", "2.25"}, param)
	// Assert
	assert.True(t, ok)
	arr := v.([]float64)
	assert.InDeltaSlice(t, []float64{1.5, 2.25}, arr, 1e-9)
}

func TestShouldReturnRawValuesForArraySchemaWithUnsupportedItemType(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "object"}}
	raw := []string{"x", "y"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, raw, v)
}

func TestShouldParseObjectSchemaMultipleValuesAsStrings(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "object"}
	raw := []string{"a", "b"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, raw, v)
}

func TestShouldSplitAndTrimCSVValues(t *testing.T) {
	// Arrange
	input := "a, b ,c"
	// Act
	parts := splitAndTrim(input)
	// Assert
	assert.Equal(t, []string{"a", "b", "c"}, parts)
}

func TestShouldSplitQuotedCSVValues(t *testing.T) {
	// Arrange
	input := `"a,b", c , "d,e"`
	// Act
	parts := splitAndTrim(input)
	// Assert
	assert.Equal(t, []string{"a,b", "c", "d,e"}, parts)
}

func TestShouldFallBackToNaiveCSVSplitForMalformedQuotedInput(t *testing.T) {
	// Arrange
	input := `"a,b,c`
	// Act
	parts := splitAndTrim(input)
	// Assert
	assert.Equal(t, []string{`"a`, "b", "c"}, parts)
}

func TestShouldHandlePointerSliceExampleCSVSplit(t *testing.T) {
	// Arrange
	ex := &[]string{"sample"}
	param := &openapi.ParameterObject{Example: ex}
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"a, b , c"}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, []string{"a", "b", "c"}, staging["k"])
}

func TestShouldIncludeLocationAndKeyInConverterError(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Converter: func([]string) (any, error) { return nil, errors.New("boom") }}
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "age", []string{"x"}, "query", param)
	// Assert
	assert.Error(t, err)
	assert.False(t, handled)
	assert.Contains(t, err.Error(), "query \"age\": boom")
}

func TestShouldParseSliceValuesGivenEmptyExampleSlice(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Example: []string{}}
	// Act
	v, ok := parseSliceValues([]string{"a", "b"}, param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, v)
}

func TestShouldReturnFalseWhenParseSliceValuesUnsupportedExampleElemType(t *testing.T) {
	// Arrange
	type custom struct{ X int }
	param := &openapi.ParameterObject{Example: []custom{{X: 1}}}
	// Act
	v, ok := parseSliceValues([]string{"a"}, param)
	// Assert
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestShouldNotOverrideStagingWhenMultiValuesAndSliceParseFails(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "integer"}}}
	staging := map[string]any{"existing": 123}
	// Act: provide one good and one bad to force failure
	handled, err := ProcessParamAndSet(staging, "numbers", []string{"1", "bad"}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.False(t, handled)
	assert.Equal(t, 123, staging["existing"])
	_, exists := staging["numbers"]
	assert.False(t, exists)
}

func TestShouldReturnRawValueWhenSchemaNilMultiValue(t *testing.T) {
	// Arrange
	raw := []string{"a", "b"}
	// Act
	v, err := parseValueBySchema(raw, nil)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, raw, v)
}

func TestShouldReturnNilAndFalseWhenProcessParamAndSetParamNil(t *testing.T) {
	// Arrange
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"x"}, "query", nil)
	// Assert
	require.NoError(t, err)
	assert.False(t, handled)
	_, ok := staging["k"]
	assert.False(t, ok)
}

// parse helpers coverage (representative failures and successes)
func TestShouldParseInt16AndFailOnInvalid(t *testing.T) {
	// Arrange & Act
	v, ok := ParseInt16Val("123")
	// Assert
	assert.True(t, ok)
	assert.Equal(t, int16(123), v)

	// Act (fail)
	_, ok = ParseInt16Val("xyz")
	assert.False(t, ok)
}

func TestShouldParseFloat32SliceAndFailOnInvalid(t *testing.T) {
	// Arrange & Act
	v, ok := ParseFloat32Slice([]string{"1.25", "2.5"})
	// Assert
	assert.True(t, ok)
	assert.InDeltaSlice(t, []float32{1.25, 2.5}, v, 1e-6)

	// Act (fail)
	_, ok = ParseFloat32Slice([]string{"1.0", "bad"})
	assert.False(t, ok)
}

func TestShouldParseUUIDSliceAndHandleInvalid(t *testing.T) {
	// Arrange
	u1 := uuid.New().String()
	u2 := uuid.New().String()
	// Act
	v, ok := ParseUUIDSlice([]string{u1, u2})
	// Assert
	assert.True(t, ok)
	assert.Len(t, v, 2)

	// Act (fail)
	_, ok = ParseUUIDSlice([]string{u1, "not-a-uuid"})
	assert.False(t, ok)
}

func TestShouldReturnBoolSliceFromSchemaArray(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "boolean"}
	// Act
	v, err := parseValueBySchema([]string{"true", "false"}, schema)
	// Assert: since schema boolean w len>1 goes through branch generating []bool
	require.NoError(t, err)
	arr := v.([]bool)
	assert.Equal(t, []bool{true, false}, arr)
}

func TestShouldReturnNilFromSchemaBooleanConverterMultiValue(t *testing.T) {
	// Arrange: makeConverter schema boolean returns nil for multi value
	conv := makeConverter(nil, &openapi.Schema{Type: "boolean"})
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"true", "false"})
	// Assert
	require.NoError(t, err)
	assert.Nil(t, v)
}

func TestShouldReturnNilFromSchemaNumberConverterMultiValue(t *testing.T) {
	// Arrange
	conv := makeConverter(nil, &openapi.Schema{Type: "number"})
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"1", "2"})
	// Assert
	require.NoError(t, err)
	assert.Nil(t, v)
}

func TestShouldErrorMidUnsignedSliceParse(t *testing.T) {
	// Arrange
	conv := makeConverter(reflect.TypeOf(uint(0)), nil)
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{"1", "bad", "3"})
	// Assert
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldReturnStringSliceForSchemaStringMultiValues(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "string"}
	// Act
	v, err := parseValueBySchema([]string{"a", "b"}, schema)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, v)
}

func TestShouldHandleArraySchemaWithNilItemsAsRaw(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "array", Items: nil}
	raw := []string{"x", "y"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, raw, v)
}

func TestShouldParseByExampleWithPointerScalar(t *testing.T) {
	// Arrange
	val := 0
	param := &openapi.ParameterObject{Example: &val}
	// Act
	v, ok := parseByExample("12", param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, 12, v)
}

func TestShouldProcessParamAndSetWithArrayNumberItemsSuccessAndFailure(t *testing.T) {
	// success
	staging := make(map[string]any)
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "number"}}}
	handled, err := ProcessParamAndSet(staging, "nums", []string{"1.5", "2.25"}, "query", param)
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, []float64{1.5, 2.25}, staging["nums"])

	// failure
	staging2 := make(map[string]any)
	handled2, err2 := ProcessParamAndSet(staging2, "nums", []string{"1.0", "bad"}, "query", param)
	require.NoError(t, err2)
	assert.False(t, handled2)
	_, ok := staging2["nums"]
	assert.False(t, ok)
}

func TestShouldDefaultToValuesWhenUnknownSchemaTypeSingleValue(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "weird"}
	// Act
	v, err := parseValueBySchema([]string{"solo"}, schema)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, "solo", v)
}

func TestShouldReturnNilFromSchemaStringUUIDConverterMultiValue(t *testing.T) {
	// Arrange
	conv := makeConverter(nil, &openapi.Schema{Type: "string", Format: "uuid"})
	require.NotNil(t, conv)
	// Act
	v, err := conv([]string{uuid.New().String(), uuid.New().String()})
	// Assert
	require.NoError(t, err)
	assert.Nil(t, v)
}

func TestShouldNotHandleWhenMultipleValuesAndParseSliceValuesFails(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "integer"}}}
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "ints", []string{"1", "bad"}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.False(t, handled)
	_, present := staging["ints"]
	assert.False(t, present)
}

func TestShouldReturnTrueForParseSliceValuesSchemaString(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Items: &openapi.Schema{Type: "string"}}}
	// Act
	v, ok := parseSliceValues([]string{"a", "b"}, param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, v)
}

func TestShouldReturnTrueForParseSliceValuesExampleStringSlice(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Example: []string{"x"}}
	// Act
	v, ok := parseSliceValues([]string{"a", "b"}, param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, []string{"a", "b"}, v)
}

func TestShouldParseByExampleWhenSchemaNilButExamplePresent(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Example: 5}
	// Act
	v, ok := parseByExample("9", param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, 9, v)
}

func TestShouldReturnFalseParseByExampleWhenExampleTypeUnsupported(t *testing.T) {
	// Arrange
	type custom struct{ A int }
	param := &openapi.ParameterObject{Example: custom{A: 1}}
	// Act
	v, ok := parseByExample("anything", param)
	// Assert
	assert.False(t, ok)
	assert.Nil(t, v)
}

func TestShouldNotHandleWhenConverterReturnsNil(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Converter: func([]string) (any, error) { return nil, nil }}
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"1"}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.False(t, handled)
}

func TestShouldHandleConverterReturningSlice(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Converter: func(vals []string) (any, error) { return []int{1, 2}, nil }}
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "k", []string{"ignored"}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, []int{1, 2}, staging["k"])
}

func TestShouldReturnUUIDSliceFromSchemaStringFormatUUIDMultipleValues(t *testing.T) {
	// Arrange
	ids := []string{uuid.New().String(), uuid.New().String()}
	schema := &openapi.Schema{Type: "string", Format: "uuid"}
	// Act
	v, err := parseValueBySchema(ids, schema)
	// Assert
	require.NoError(t, err)
	arr := v.([]uuid.UUID)
	assert.Len(t, arr, 2)
}

func TestShouldReturnRawValuesForBooleanArrayItemsSchema(t *testing.T) {
	// Arrange: boolean item type is not explicitly handled in array branch, so raw values returned
	schema := &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "boolean"}}
	raw := []string{"true", "false", "true"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, raw, v)
}

func TestShouldReturnRawValuesForBooleanArrayItemsSchemaWithInvalidValue(t *testing.T) {
	// Arrange: still returns raw slice, no validation performed
	schema := &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "boolean"}}
	raw := []string{"true", "nope"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, raw, v)
}

func TestShouldErrorParsingInvalidUuidInSchemaArray(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "string", Format: "uuid"}
	raw := []string{"not-a-uuid"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldErrorParsingInvalidIntegerInSchemaArray(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "integer"}
	raw := []string{"bad"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldErrorParsingInvalidNumberInSchemaArray(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "number"}
	raw := []string{"bad"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldErrorParsingInvalidBooleanInSchema(t *testing.T) {
	// Arrange
	schema := &openapi.Schema{Type: "boolean"}
	raw := []string{"notabool"}
	// Act
	v, err := parseValueBySchema(raw, schema)
	// Assert
	assert.Error(t, err)
	assert.Nil(t, v)
}

func TestShouldErrorFromConverterSurfaceLocationAndKey(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Converter: func([]string) (any, error) { return nil, fmt.Errorf("broken") }}
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "q", []string{"1"}, "header", param)
	// Assert
	assert.Error(t, err)
	assert.False(t, handled)
	assert.Contains(t, err.Error(), "header \"q\": broken")
}
