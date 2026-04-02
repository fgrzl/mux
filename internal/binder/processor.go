package binder

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"

	"github.com/fgrzl/mux/internal/openapi"
)

// ProcessParamAndSet centralizes parameter parsing/conversion logic for query/header/path params.
// It returns (true, nil) if it set a typed value on staging, (false, nil) if caller should fall back
// to storing raw values, or (false, err) if a conversion error occurred.
func ProcessParamAndSet(staging map[string]any, key string, values []string, location string, param *openapi.ParameterObject) (bool, error) {
	if param == nil {
		return false, nil
	}
	// normalize values according to Schema/Example (split CSV into slice when needed)
	values = normalizeValuesForParam(values, param)
	// converter has highest precedence
	if param.Converter != nil {
		if typed, err := param.Converter(values); err != nil {
			return false, fmt.Errorf("%s %q: %w", location, key, err)
		} else if typed != nil {
			staging[key] = typed
			return true, nil
		}
	}
	if len(values) == 1 {
		if parsed, ok := ParseByExample(values[0], param); ok {
			staging[key] = parsed
			return true, nil
		}
	} else {
		if parsedSlice, ok := ParseSliceValues(values, param); ok {
			staging[key] = parsedSlice
			return true, nil
		}
	}
	return false, nil
}

func splitAndTrim(s string) []string {
	reader := csv.NewReader(strings.NewReader(s))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	parts, err := reader.Read()
	if err != nil {
		return splitAndTrimFallback(s)
	}

	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func splitAndTrimFallback(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

// isParamArray returns true when the ParameterObject expects an array
// either via Schema type or Example value being a slice.
func isParamArray(param *openapi.ParameterObject) bool {
	if param == nil {
		return false
	}
	if param.Schema != nil && param.Schema.Type == "array" {
		return true
	}
	if param.Example != nil {
		exT := reflect.TypeOf(param.Example)
		if exT.Kind() == reflect.Ptr {
			exT = exT.Elem()
		}
		return exT.Kind() == reflect.Slice
	}
	return false
}

// normalizeValuesForParam will split a single CSV value into elements when
// the parameter is considered an array.
func normalizeValuesForParam(values []string, param *openapi.ParameterObject) []string {
	if isParamArray(param) && len(values) == 1 && strings.Contains(values[0], ",") {
		return splitAndTrim(values[0])
	}
	return values
}
