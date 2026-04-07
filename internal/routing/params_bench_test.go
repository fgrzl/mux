package routing

import (
	"testing"
)

// BenchmarkParamsGet benchmarks parameter retrieval from slice-based Params
func BenchmarkParamsGet(b *testing.B) {
	params := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
		{Key: "category", Value: "books"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = params.Get("category")
	}
}

// BenchmarkParamsSet benchmarks adding/updating parameters in slice-based Params
func BenchmarkParamsSet(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := make(Params, 0, 4)
		params.Set("id", "123")
		params.Set("name", "test")
		params.Set("category", "books")
	}
}

// BenchmarkParamsPool benchmarks acquiring and releasing from the Params pool
func BenchmarkParamsPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := AcquireParams()
		params.Set("id", "123")
		params.Set("name", "test")
		_ = params.Get("name")
		ReleaseParams(params)
	}
}

// BenchmarkParamsIteration benchmarks iterating over slice-based Params
func BenchmarkParamsIteration(b *testing.B) {
	params := Params{
		{Key: "id", Value: "123"},
		{Key: "name", Value: "test"},
		{Key: "category", Value: "books"},
		{Key: "author", Value: "john"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range params {
			_ = params[j].Key
			_ = params[j].Value
		}
	}
}
