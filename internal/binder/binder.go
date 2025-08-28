package binder

import (
	"reflect"

	"github.com/fgrzl/mux/pkg/openapi"
)

// MakeConverter is a thin wrapper that delegates to the package-local makeConverter
// implementation to avoid duplicating conversion logic across files.
func MakeConverter(t reflect.Type, schema *openapi.Schema) func([]string) (any, error) {
	return makeConverter(t, schema)
}

// ParseValueBySchema coerces raw string values into typed value guided by the provided Schema.
// This exported wrapper delegates to the package-local parseValueBySchema implementation.
func ParseValueBySchema(values []string, schema *openapi.Schema) (any, error) {
	return parseValueBySchema(values, schema)
}

// ParseByExample attempts to parse a single string value into the type suggested by the
// ParameterObject's Example or Schema. It delegates to parseByExample.
func ParseByExample(val string, param *openapi.ParameterObject) (any, bool) {
	return parseByExample(val, param)
}

// ParseSliceValues delegates to parseSliceValues to avoid duplicating parsing logic.
func ParseSliceValues(values []string, param *openapi.ParameterObject) (any, bool) {
	return parseSliceValues(values, param)
}
