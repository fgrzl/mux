package binder

import (
	"reflect"
	"testing"

	"github.com/fgrzl/mux/internal/openapi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldParseIntVal(t *testing.T) {
	// Arrange & Act
	v, ok := ParseIntVal("42")
	// Assert
	assert.True(t, ok)
	assert.Equal(t, 42, v)
}

func TestShouldParseIntSlice(t *testing.T) {
	// Act
	sl, ok := ParseIntSlice([]string{"1", "2"})
	// Assert
	assert.True(t, ok)
	assert.Equal(t, []int{1, 2}, sl)
}

func TestShouldFailParseIntSliceOnInvalid(t *testing.T) {
	// Act
	_, ok := ParseIntSlice([]string{"1", "x"})
	// Assert
	assert.False(t, ok)
}

func TestShouldParseInt32Val(t *testing.T) {
	// Arrange & Act
	v32, ok := ParseInt32Val("2147483647")
	// Assert
	assert.True(t, ok)
	assert.Equal(t, int32(2147483647), v32)
}

func TestShouldParseInt64Val(t *testing.T) {
	// Act
	v64, ok := ParseInt64Val("9223372036854775807")
	// Assert
	assert.True(t, ok)
	assert.Equal(t, int64(9223372036854775807), v64)
}

func TestShouldParseInt32Slice(t *testing.T) {
	// Act
	s32, ok := ParseInt32Slice([]string{"1", "2"})
	// Assert
	assert.True(t, ok)
	assert.Equal(t, []int32{1, 2}, s32)
}

func TestShouldParseInt64Slice(t *testing.T) {
	// Act
	s64, ok := ParseInt64Slice([]string{"3", "4"})
	// Assert
	assert.True(t, ok)
	assert.Equal(t, []int64{3, 4}, s64)
}

func TestShouldFailParseInt32SliceOnInvalid(t *testing.T) {
	// Act
	_, ok := ParseInt32Slice([]string{"1", "bad"})
	// Assert
	assert.False(t, ok)
}

func TestShouldParseBoolVal(t *testing.T) {
	b, ok := ParseBoolVal("true")
	assert.True(t, ok)
	assert.True(t, b)
}

func TestShouldParseBoolSlice(t *testing.T) {
	bs, ok := ParseBoolSlice([]string{"true", "false"})
	assert.True(t, ok)
	assert.Equal(t, []bool{true, false}, bs)
}

func TestShouldParseFloat32Val(t *testing.T) {
	f32, ok := ParseFloat32Val("1.5")
	assert.True(t, ok)
	assert.InDelta(t, float32(1.5), f32, 1e-6)
}

func TestShouldParseFloat32Slice(t *testing.T) {
	f32s, ok := ParseFloat32Slice([]string{"1.25", "2.5"})
	assert.True(t, ok)
	assert.InDeltaSlice(t, []float32{1.25, 2.5}, f32s, 1e-6)
}

func TestShouldParseFloat64Val(t *testing.T) {
	f64, ok := ParseFloat64Val("2.75")
	assert.True(t, ok)
	assert.InDelta(t, 2.75, f64, 1e-9)
}

func TestShouldParseFloat64Slice(t *testing.T) {
	f64s, ok := ParseFloat64Slice([]string{"3.5", "4.75"})
	assert.True(t, ok)
	assert.InDeltaSlice(t, []float64{3.5, 4.75}, f64s, 1e-9)
}

func TestShouldParseInt16Slice(t *testing.T) {
	// Act
	s, ok := ParseInt16Slice([]string{"10", "20"})
	// Assert
	assert.True(t, ok)
	assert.Equal(t, []int16{10, 20}, s)
}

func TestShouldFailParseInt16SliceOnInvalid(t *testing.T) {
	// Act
	_, ok := ParseInt16Slice([]string{"10", "x"})
	// Assert
	assert.False(t, ok)
}

func TestShouldParseUUIDValue(t *testing.T) {
	id1 := uuid.New().String()
	u, ok := ParseUUIDVal(id1)
	assert.True(t, ok)
	assert.Equal(t, id1, u.String())
}

func TestShouldParseUUIDSlice(t *testing.T) {
	id1 := uuid.New().String()
	id2 := uuid.New().String()
	us, ok := ParseUUIDSlice([]string{id1, id2})
	assert.True(t, ok)
	require.Len(t, us, 2)
}

func TestShouldConvertSignedIntsAndFloatsAndBoolsWithMakeConverter(t *testing.T) {
	// Signed ints
	kinds := []reflect.Type{reflect.TypeOf(int(0)), reflect.TypeOf(int8(0)), reflect.TypeOf(int16(0)), reflect.TypeOf(int32(0)), reflect.TypeOf(int64(0))}
	for _, typ := range kinds {
		// Arrange
		conv := makeConverter(typ, nil)
		require.NotNil(t, conv)
		// Act (single)
		v, err := conv([]string{"7"})
		// Assert
		require.NoError(t, err)
		assert.NotNil(t, v)
		// Act (multi)
		mv, err := conv([]string{"1", "2"})
		// Assert
		require.NoError(t, err)
		assert.IsType(t, []int64{}, mv)
	}

	// Floats
	f32 := makeConverter(reflect.TypeOf(float32(0)), nil)
	require.NotNil(t, f32)
	mv, err := f32([]string{"1.5", "2.25"})
	require.NoError(t, err)
	assert.IsType(t, []float64{}, mv)

	f64 := makeConverter(reflect.TypeOf(float64(0)), nil)
	require.NotNil(t, f64)
	mv2, err := f64([]string{"1.5", "2.25"})
	require.NoError(t, err)
	assert.IsType(t, []float64{}, mv2)

	// Bool
	bconv := makeConverter(reflect.TypeOf(true), nil)
	require.NotNil(t, bconv)
	bm, err := bconv([]string{"true", "false"})
	require.NoError(t, err)
	assert.Equal(t, []bool{true, false}, bm)
}

func TestShouldConvertStringMultiValue(t *testing.T) {
	sconv := makeConverter(reflect.TypeOf(""), nil)
	require.NotNil(t, sconv)
	mv, err := sconv([]string{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, mv)
}

func TestShouldConvertExampleBoolSlice(t *testing.T) {
	ex := []bool{false}
	conv := makeConverter(reflect.TypeOf(ex), nil)
	require.NotNil(t, conv)
	out, err := conv([]string{"true", "false"})
	assert.NoError(t, err)
	arr := out.([]bool)
	require.Len(t, arr, 2)
	assert.Equal(t, true, arr[0])
	assert.Equal(t, false, arr[1])
}

func TestShouldUseSchemaConvertersAndHandleSingleAndMultiValues(t *testing.T) {
	// integer schema single
	iconv := makeConverter(nil, &openapi.Schema{Type: "integer"})
	require.NotNil(t, iconv)
	v, err := iconv([]string{"10"})
	require.NoError(t, err)
	assert.Equal(t, int64(10), v)

	// multi => []int64
	mv, err := iconv([]string{"1", "2"})
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2}, mv)

	// number multi => []float64
	nconv := makeConverter(nil, &openapi.Schema{Type: "number"})
	mvn, err := nconv([]string{"1.0", "2.0"})
	require.NoError(t, err)
	assert.Equal(t, []float64{1, 2}, mvn)

	// boolean multi => []bool
	bconv := makeConverter(nil, &openapi.Schema{Type: "boolean"})
	mvb, err := bconv([]string{"true", "false"})
	require.NoError(t, err)
	assert.Equal(t, []bool{true, false}, mvb)

	// string uuid multi => []uuid.UUID
	uconv := makeConverter(nil, &openapi.Schema{Type: "string", Format: "uuid"})
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	mvu, err := uconv([]string{ids[0].String(), ids[1].String()})
	require.NoError(t, err)
	assert.Equal(t, ids, mvu)
}

func TestShouldParseValueBySchemaStringVariantsAndDefaults(t *testing.T) {
	// string (non-uuid) single & multi
	v, err := parseValueBySchema([]string{"x"}, &openapi.Schema{Type: "string"})
	require.NoError(t, err)
	assert.Equal(t, "x", v)

	mv, err := parseValueBySchema([]string{"a", "b"}, &openapi.Schema{Type: "string"})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, mv)

	// unknown type -> default behavior
	dv, err := parseValueBySchema([]string{"d1", "d2"}, &openapi.Schema{Type: "date"})
	require.NoError(t, err)
	assert.Equal(t, []string{"d1", "d2"}, dv)

	// nil schema multi => raw
	rv, err := parseValueBySchema([]string{"r1", "r2"}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"r1", "r2"}, rv)
}

func TestShouldProcessParamWithArraySchemaAndCSV(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}}}
	staging := make(map[string]any)
	// Act
	handled, err := ProcessParamAndSet(staging, "tags", []string{"a, b, c"}, "query", param)
	// Assert
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, []string{"a", "b", "c"}, staging["tags"])
}

func TestShouldParseBySchemaWhenExampleMissing(t *testing.T) {
	// Arrange
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Type: "integer"}}
	// Act
	v, ok := parseByExample("7", param)
	// Assert
	assert.True(t, ok)
	assert.Equal(t, int64(7), v)
}
