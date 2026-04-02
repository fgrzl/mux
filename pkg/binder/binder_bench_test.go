package binder

import (
	"io"
	"log/slog"
	"reflect"
	"testing"

	openapi "github.com/fgrzl/mux/pkg/openapi"
)

func init() {
	// silence slog during benchmarks to avoid noisy output
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
}

func BenchmarkMakeConverterString(b *testing.B) {
	conv := makeConverter(reflect.TypeOf(""), nil)
	vals := []string{"hello"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = conv(vals)
	}
}

func BenchmarkMakeConverterIntSlice(b *testing.B) {
	conv := makeConverter(reflect.SliceOf(reflect.TypeOf(int64(0))), nil)
	vals := []string{"1", "2", "3", "4", "5"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = conv(vals)
	}
}

func BenchmarkParseValueBySchemaInteger(b *testing.B) {
	schema := &openapi.Schema{Type: "integer"}
	vals := []string{"12345"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseValueBySchema(vals, schema)
	}
}

func BenchmarkParseValueBySchemaStringUUID(b *testing.B) {
	schema := &openapi.Schema{Type: "string", Format: "uuid"}
	vals := []string{"6ba7b814-9dad-11d1-80b4-00c04fd430c8"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseValueBySchema(vals, schema)
	}
}

func BenchmarkParseByExampleWithExample(b *testing.B) {
	param := &openapi.ParameterObject{Example: 123}
	val := "123"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseByExample(val, param)
	}
}

func BenchmarkParseSliceValuesSchemaInteger(b *testing.B) {
	param := &openapi.ParameterObject{Schema: &openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "integer"}}}
	vals := []string{"1", "2", "3", "4", "5"}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseSliceValues(vals, param)
	}
}
