package binder

import (
	"math"
	"reflect"
	"strconv"

	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/google/uuid"
)

// makeConverter builds a runtime converter for the given example type or schema.
// The returned function accepts the raw string values (multi-valued) and
// returns a typed value suitable for placing into the staging map.
func makeConverter(t reflect.Type, schema *openapi.Schema) func([]string) (any, error) {
	if t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t != nil {
		if c := scalarConverterForType(t); c != nil {
			return c
		}
		if conv := makeSliceElementConverter(t); conv != nil {
			return conv
		}
	}

	if schema != nil {
		return makeConverterFromSchema(schema)
	}

	return nil
}

// makeSliceElementConverter returns a converter that parses each element
// using the element's scalar converter, or nil if not applicable.
func makeSliceElementConverter(t reflect.Type) func([]string) (any, error) {
	if t == nil || t.Kind() != reflect.Slice {
		return nil
	}
	et := t.Elem()
	c := scalarConverterForType(et)
	if c == nil {
		return nil
	}
	return func(vals []string) (any, error) {
		out := make([]any, 0, len(vals))
		for _, s := range vals {
			parsed, err := c([]string{s})
			if err != nil {
				return nil, err
			}
			out = append(out, parsed)
		}
		return out, nil
	}
}

// makeConverterFromSchema creates a converter based on an OpenAPI schema.
func makeConverterFromSchema(schema *openapi.Schema) func([]string) (any, error) {
	if schema == nil {
		return nil
	}
	switch schema.Type {
	case "integer":
		return converterForInteger()
	case "number":
		return converterForNumber()
	case "boolean":
		return converterForBoolean()
	case "string":
		return converterForString(schema)
	default:
		return nil
	}
}

func converterForInteger() func([]string) (any, error) {
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			v, err := strconv.ParseInt(vals[0], 10, 64)
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		return nil, nil
	}
}

func converterForNumber() func([]string) (any, error) {
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			v, err := strconv.ParseFloat(vals[0], 64)
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		return nil, nil
	}
}

func converterForBoolean() func([]string) (any, error) {
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			v, err := strconv.ParseBool(vals[0])
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		return nil, nil
	}
}

func converterForString(schema *openapi.Schema) func([]string) (any, error) {
	if schema != nil && schema.Format == "uuid" {
		return func(vals []string) (any, error) {
			if len(vals) == 1 {
				u, err := uuid.Parse(vals[0])
				if err != nil {
					return nil, err
				}
				return u, nil
			}
			return nil, nil
		}
	}
	return makeStringConverter()
}

// parseValueBySchema attempts to coerce raw string values into a typed value
// guided by the provided Schema. Returns error on parse failure.
func parseValueBySchema(values []string, schema *openapi.Schema) (any, error) {
	if schema == nil {
		if len(values) == 1 {
			return values[0], nil
		}
		return values, nil
	}
	switch schema.Type {
	case "string":
		if schema.Format == "uuid" {
			return singleOrSliceUUID(values)
		}
		return singleOrSliceString(values)
	case "integer":
		return singleOrSliceInt64(values)
	case "number":
		return singleOrSliceFloat64(values)
	case "boolean":
		return singleOrSliceBool(values)
	case "array":
		if schema.Items != nil {
			switch schema.Items.Type {
			case "string":
				return values, nil
			case "integer":
				return parseInt64s(values)
			case "number":
				return parseFloat64s(values)
			}
		}
		return values, nil
	case "object":
		return singleOrSliceString(values)
	default:
		return singleOrSliceString(values)
	}
}

// scalarConverterForType returns a converter function for simple scalar types.
// It mirrors the previous scalarConv closure extracted from makeConverter.
func scalarConverterForType(typ reflect.Type) func([]string) (any, error) {
	if typ == nil {
		return nil
	}
	switch typ.Kind() {
	case reflect.String:
		return makeStringConverter()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return makeIntConverter(typ)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return makeUintConverter(typ)
	case reflect.Float32, reflect.Float64:
		return makeFloatConverter(typ)
	case reflect.Bool:
		return makeBoolConverter()
	case reflect.Slice, reflect.Struct:
		return nil
	default:
		return nil
	}
}

func makeStringConverter() func([]string) (any, error) {
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			return vals[0], nil
		}
		return vals, nil
	}
}

func makeIntConverter(typ reflect.Type) func([]string) (any, error) {
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			if typ.Kind() == reflect.Int {
				iv, err := parsePlatformInt(vals[0])
				if err != nil {
					return nil, err
				}
				return iv, nil
			}

			v, err := strconv.ParseInt(vals[0], 10, typ.Bits())
			if err != nil {
				return nil, err
			}
			switch typ.Kind() {
			case reflect.Int8:
				return int8(v), nil
			case reflect.Int16:
				return int16(v), nil
			case reflect.Int32:
				return int32(v), nil
			default:
				return v, nil
			}
		}
		// Multi-valued: parse into int64 slice (safer and consistent across callers).
		return parseInt64s(vals)
	}
}

func makeUintConverter(typ reflect.Type) func([]string) (any, error) {
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			if typ.Kind() == reflect.Uint {
				uv, err := parsePlatformUint(vals[0])
				if err != nil {
					return nil, err
				}
				return uv, nil
			}

			v, err := strconv.ParseUint(vals[0], 10, typ.Bits())
			if err != nil {
				return nil, err
			}
			switch typ.Kind() {
			case reflect.Uint8:
				return uint8(v), nil
			case reflect.Uint16:
				return uint16(v), nil
			case reflect.Uint32:
				return uint32(v), nil
			default:
				return v, nil
			}
		}
		// Multi-valued: parse into uint64 slice (safer and consistent across callers).
		return parseUint64s(vals)
	}
}

// parsePlatformInt parses s as a 64-bit integer and ensures it fits into the
// host platform's int size before returning it as int.
func parsePlatformInt(s string) (int, error) {
	v64, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if v64 < int64(math.MinInt) || v64 > int64(math.MaxInt) {
		return 0, strconv.ErrRange
	}
	return int(v64), nil
}

// parsePlatformUint parses s as a 64-bit unsigned integer and ensures it fits
// into the host platform's uint size before returning it as uint.
func parsePlatformUint(s string) (uint, error) {
	v64, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	maxUint := uint64(^uint(0))
	if v64 > maxUint {
		return 0, strconv.ErrRange
	}
	return uint(v64), nil
}

// helper: parse a slice of uint64 values
func parseUint64s(values []string) (any, error) {
	out := make([]uint64, 0, len(values))
	for _, s := range values {
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func makeFloatConverter(typ reflect.Type) func([]string) (any, error) {
	bits := 64
	if typ.Kind() == reflect.Float32 {
		bits = 32
	}
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			v, err := strconv.ParseFloat(vals[0], bits)
			if err != nil {
				return nil, err
			}
			if bits == 32 {
				return float32(v), nil
			}
			return v, nil
		}
		return parseFloat64s(vals)
	}
}

func makeBoolConverter() func([]string) (any, error) {
	return func(vals []string) (any, error) {
		if len(vals) == 1 {
			v, err := strconv.ParseBool(vals[0])
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		out := make([]bool, 0, len(vals))
		for _, s := range vals {
			v, err := strconv.ParseBool(s)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	}
}

// helper: return single string or slice
func singleOrSliceString(values []string) (any, error) {
	if len(values) == 1 {
		return values[0], nil
	}
	return values, nil
}

// helper: parse a slice of int64 values
func parseInt64s(values []string) (any, error) {
	out := make([]int64, 0, len(values))
	for _, s := range values {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// helper: parse a slice of float64 values
func parseFloat64s(values []string) (any, error) {
	out := make([]float64, 0, len(values))
	for _, s := range values {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// helper: return a single uuid.UUID or slice of them
func singleOrSliceUUID(values []string) (any, error) {
	if len(values) == 1 {
		u, err := uuid.Parse(values[0])
		if err != nil {
			return nil, err
		}
		return u, nil
	}
	out := make([]uuid.UUID, 0, len(values))
	for _, s := range values {
		u, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

// helper: return single int64 or slice
func singleOrSliceInt64(values []string) (any, error) {
	if len(values) == 1 {
		v, err := strconv.ParseInt(values[0], 10, 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return parseInt64s(values)
}

// helper: return single float64 or slice
func singleOrSliceFloat64(values []string) (any, error) {
	if len(values) == 1 {
		v, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	return parseFloat64s(values)
}

// helper: return single bool or slice
func singleOrSliceBool(values []string) (any, error) {
	if len(values) == 1 {
		v, err := strconv.ParseBool(values[0])
		if err != nil {
			return nil, err
		}
		return v, nil
	}
	out := make([]bool, 0, len(values))
	for _, s := range values {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// parseByExample attempts to parse a single string value into the type suggested by the ParameterObject's Example or Schema.
func parseByExample(val string, param *openapi.ParameterObject) (any, bool) {
	if param == nil {
		return nil, false
	}
	// prefer Example if present
	if param.Example != nil {
		if conv := makeConverter(reflect.TypeOf(param.Example), param.Schema); conv != nil {
			if v, err := conv([]string{val}); err == nil && v != nil {
				return v, true
			}
		}
	}
	if param.Schema != nil {
		if v, err := parseValueBySchema([]string{val}, param.Schema); err == nil {
			return v, true
		}
	}
	return nil, false
}

func parseSlice[T any](vals []string, parse func(string) (T, error)) ([]T, bool) {
	result := make([]T, 0, len(vals))
	for _, val := range vals {
		parsed, err := parse(val)
		if err != nil {
			return nil, false
		}
		result = append(result, parsed)
	}
	return result, true
}

func parseSliceValues(values []string, param *openapi.ParameterObject) (any, bool) {
	if param == nil {
		return nil, false
	}
	if parsed, ok := parseSliceFromExample(values, param.Example); ok {
		return parsed, true
	}
	if parsed, ok := parseSliceFromSchema(values, param.Schema); ok {
		return parsed, true
	}
	return nil, false
}

// parseSliceFromExample inspects an example value to infer slice element types.
func parseSliceFromExample(values []string, example any) (any, bool) {
	if example == nil {
		return nil, false
	}
	t := reflect.TypeOf(example)
	if t.Kind() != reflect.Slice {
		return nil, false
	}
	exVal := reflect.ValueOf(example)
	if exVal.Len() == 0 {
		return nil, false
	}
	elem := exVal.Index(0).Interface()
	switch elem.(type) {
	case int:
		if parsed, ok := parseSlice(values, func(s string) (int, error) {
			v64, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return 0, err
			}
			if v64 < int64(math.MinInt) || v64 > int64(math.MaxInt) {
				return 0, strconv.ErrRange
			}
			return int(v64), nil
		}); ok {
			return parsed, true
		}
	case string:
		return values, true
	}
	return nil, false
}

// parseSliceFromSchema parses slice values based on Schema->Items type.
func parseSliceFromSchema(values []string, schema *openapi.Schema) (any, bool) {
	if schema == nil || schema.Items == nil {
		return nil, false
	}
	switch schema.Items.Type {
	case "integer":
		if parsed, ok := parseSlice(values, func(s string) (int64, error) {
			return strconv.ParseInt(s, 10, 64)
		}); ok {
			return parsed, true
		}
	case "number":
		if parsed, ok := parseSlice(values, func(s string) (float64, error) {
			return strconv.ParseFloat(s, 64)
		}); ok {
			return parsed, true
		}
	case "string":
		return values, true
	}
	return nil, false
}
