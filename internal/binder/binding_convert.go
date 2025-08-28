package binder

import (
	"reflect"
	"strconv"

	"github.com/fgrzl/mux/internal/openapi"
	"github.com/google/uuid"
)

// makeConverter builds a runtime converter for the given example type or schema.
// The returned function accepts the raw string values (multi-valued) and
// returns a typed value suitable for placing into the staging map.
func makeConverter(t reflect.Type, schema *openapi.Schema) func([]string) (any, error) {
	if t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// scalar converters
	scalarConv := func(typ reflect.Type) func([]string) (any, error) {
		if typ == nil {
			return nil
		}
		switch typ.Kind() {
		case reflect.String:
			return func(vals []string) (any, error) {
				if len(vals) == 1 {
					return vals[0], nil
				}
				return vals, nil
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return func(vals []string) (any, error) {
				if len(vals) == 1 {
					v, err := strconv.ParseInt(vals[0], 10, typ.Bits())
					if err != nil {
						return nil, err
					}
					switch typ.Kind() {
					case reflect.Int:
						return int(v), nil
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
				// parse slice of ints as []int64 by default
				out := make([]int64, 0, len(vals))
				for _, s := range vals {
					v, err := strconv.ParseInt(s, 10, 64)
					if err != nil {
						return nil, err
					}
					out = append(out, v)
				}
				return out, nil
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return func(vals []string) (any, error) {
				if len(vals) == 1 {
					v, err := strconv.ParseUint(vals[0], 10, typ.Bits())
					if err != nil {
						return nil, err
					}
					switch typ.Kind() {
					case reflect.Uint:
						return uint(v), nil
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
				out := make([]uint64, 0, len(vals))
				for _, s := range vals {
					v, err := strconv.ParseUint(s, 10, 64)
					if err != nil {
						return nil, err
					}
					out = append(out, v)
				}
				return out, nil
			}
		case reflect.Float32, reflect.Float64:
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
				out := make([]float64, 0, len(vals))
				for _, s := range vals {
					v, err := strconv.ParseFloat(s, 64)
					if err != nil {
						return nil, err
					}
					out = append(out, v)
				}
				return out, nil
			}
		case reflect.Bool:
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
		case reflect.Slice:
			return nil
		case reflect.Struct:
			return nil
		default:
			return nil
		}
	}

	if t != nil {
		if c := scalarConv(t); c != nil {
			return c
		}
		if t.Kind() == reflect.Slice {
			et := t.Elem()
			if c := scalarConv(et); c != nil {
				return func(vals []string) (any, error) {
					out := make([]any, 0, len(vals))
					for _, s := range vals {
						parsed, err := scalarConv(et)([]string{s})
						if err != nil {
							return nil, err
						}
						out = append(out, parsed)
					}
					return out, nil
				}
			}
		}
	}

	if schema != nil {
		switch schema.Type {
		case "integer":
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
		case "number":
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
		case "boolean":
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
		case "string":
			if schema.Format == "uuid" {
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
			return func(vals []string) (any, error) {
				if len(vals) == 1 {
					return vals[0], nil
				}
				return vals, nil
			}
		}
	}

	return nil
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
		if len(values) == 1 {
			return values[0], nil
		}
		return values, nil
	case "integer":
		if len(values) == 1 {
			v, err := strconv.ParseInt(values[0], 10, 64)
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		out := make([]int64, 0, len(values))
		for _, s := range values {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case "number":
		if len(values) == 1 {
			v, err := strconv.ParseFloat(values[0], 64)
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		out := make([]float64, 0, len(values))
		for _, s := range values {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case "boolean":
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
	case "array":
		if schema.Items != nil {
			switch schema.Items.Type {
			case "string":
				return values, nil
			case "integer":
				out := make([]int64, 0, len(values))
				for _, s := range values {
					v, err := strconv.ParseInt(s, 10, 64)
					if err != nil {
						return nil, err
					}
					out = append(out, v)
				}
				return out, nil
			case "number":
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
		}
		return values, nil
	case "object":
		if len(values) == 1 {
			return values[0], nil
		}
		return values, nil
	default:
		if len(values) == 1 {
			return values[0], nil
		}
		return values, nil
	}
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

func parseSliceValues(values []string, param *openapi.ParameterObject) (any, bool) {
	if param == nil {
		return nil, false
	}
	if param.Example != nil {
		if reflect.TypeOf(param.Example).Kind() == reflect.Slice {
			exVal := reflect.ValueOf(param.Example)
			if exVal.Len() > 0 {
				elem := exVal.Index(0).Interface()
				switch elem.(type) {
				case int:
					if parsed, ok := parseSlice[int](values, func(s string) (int, error) {
						v, err := strconv.Atoi(s)
						return v, err
					}); ok {
						return parsed, true
					}
				case string:
					return values, true
				}
			}
		}
	}
	if param.Schema != nil && param.Schema.Items != nil {
		switch param.Schema.Items.Type {
		case "integer":
			if parsed, ok := parseSlice[int64](values, func(s string) (int64, error) {
				return strconv.ParseInt(s, 10, 64)
			}); ok {
				return parsed, true
			}
		case "number":
			if parsed, ok := parseSlice[float64](values, func(s string) (float64, error) {
				return strconv.ParseFloat(s, 64)
			}); ok {
				return parsed, true
			}
		case "string":
			return values, true
		}
	}
	return nil, false
}
