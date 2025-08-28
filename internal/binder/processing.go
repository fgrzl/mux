package binder

import (
	"fmt"
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
	// if parameter expects array and a single CSV value was given, split it
	if param.Schema != nil && param.Schema.Type == "array" && len(values) == 1 && strings.Contains(values[0], ",") {
		values = splitAndTrim(values[0])
	}
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
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}
