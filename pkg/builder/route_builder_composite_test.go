package builder

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/common"
)

type Dog struct {
	Type string `json:"type"`
	Bark string `json:"bark"`
}

type Cat struct {
	Type string `json:"type"`
	Meow string `json:"meow"`
}

type Bird struct {
	Type string `json:"type"`
	Sing string `json:"sing"`
}

func TestWithOneOfJsonBodyShouldCreateOneOfSchema(t *testing.T) {
	rb := Route(http.MethodPost, "/pets").
		WithOneOfJsonBody(
			Dog{Type: "dog", Bark: "woof"},
			Cat{Type: "cat", Meow: "meow"},
		).
		WithOperationID("createPet")

	if rb.Options.RequestBody == nil {
		t.Fatal("expected request body to be set")
	}

	mediaType := rb.Options.RequestBody.Content[common.MimeJSON]
	if mediaType == nil {
		t.Fatal("expected JSON media type to be set")
	}

	schema := mediaType.Schema
	if schema == nil {
		t.Fatal("expected schema to be set")
	}

	if len(schema.OneOf) != 2 {
		t.Fatalf("expected 2 oneOf schemas, got %d", len(schema.OneOf))
	}

	// Verify the schemas are references to the component schemas
	if schema.OneOf[0].Ref != "#/components/schemas/Dog" {
		t.Errorf("expected first schema to reference Dog, got %s", schema.OneOf[0].Ref)
	}
	if schema.OneOf[1].Ref != "#/components/schemas/Cat" {
		t.Errorf("expected second schema to reference Cat, got %s", schema.OneOf[1].Ref)
	}
}

func TestWithAnyOfJsonBodyShouldCreateAnyOfSchema(t *testing.T) {
	rb := Route(http.MethodPost, "/animals").
		WithAnyOfJsonBody(
			Dog{Type: "dog", Bark: "woof"},
			Cat{Type: "cat", Meow: "meow"},
			Bird{Type: "bird", Sing: "tweet"},
		).
		WithOperationID("createAnimal")

	if rb.Options.RequestBody == nil {
		t.Fatal("expected request body to be set")
	}

	mediaType := rb.Options.RequestBody.Content[common.MimeJSON]
	if mediaType == nil {
		t.Fatal("expected JSON media type to be set")
	}

	schema := mediaType.Schema
	if schema == nil {
		t.Fatal("expected schema to be set")
	}

	if len(schema.AnyOf) != 3 {
		t.Fatalf("expected 3 anyOf schemas, got %d", len(schema.AnyOf))
	}

	if schema.AnyOf[0].Ref != "#/components/schemas/Dog" {
		t.Errorf("expected first schema to reference Dog, got %s", schema.AnyOf[0].Ref)
	}
	if schema.AnyOf[1].Ref != "#/components/schemas/Cat" {
		t.Errorf("expected second schema to reference Cat, got %s", schema.AnyOf[1].Ref)
	}
	if schema.AnyOf[2].Ref != "#/components/schemas/Bird" {
		t.Errorf("expected third schema to reference Bird, got %s", schema.AnyOf[2].Ref)
	}
}

func TestWithAllOfJsonBodyShouldCreateAllOfSchema(t *testing.T) {
	type BaseEntity struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	type Metadata struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	rb := Route(http.MethodPost, "/entities").
		WithAllOfJsonBody(
			BaseEntity{ID: "123", Name: "Test"},
			Metadata{CreatedAt: "2024-01-01", UpdatedAt: "2024-01-02"},
		).
		WithOperationID("createEntity")

	if rb.Options.RequestBody == nil {
		t.Fatal("expected request body to be set")
	}

	mediaType := rb.Options.RequestBody.Content[common.MimeJSON]
	if mediaType == nil {
		t.Fatal("expected JSON media type to be set")
	}

	schema := mediaType.Schema
	if schema == nil {
		t.Fatal("expected schema to be set")
	}

	if len(schema.AllOf) != 2 {
		t.Fatalf("expected 2 allOf schemas, got %d", len(schema.AllOf))
	}

	if schema.AllOf[0].Ref != "#/components/schemas/BaseEntity" {
		t.Errorf("expected first schema to reference BaseEntity, got %s", schema.AllOf[0].Ref)
	}
	if schema.AllOf[1].Ref != "#/components/schemas/Metadata" {
		t.Errorf("expected second schema to reference Metadata, got %s", schema.AllOf[1].Ref)
	}
}

func TestWithOneOfJsonBodyShouldHandleEmptySlice(t *testing.T) {
	rb := Route(http.MethodPost, "/pets").
		WithOneOfJsonBody().
		WithOperationID("createPet")

	if rb.Options.RequestBody != nil {
		t.Error("expected request body to be nil for empty examples")
	}
}

func TestWithOneOfJsonBodyShouldSkipNilExamples(t *testing.T) {
	rb := Route(http.MethodPost, "/pets").
		WithOneOfJsonBody(
			Dog{Type: "dog", Bark: "woof"},
			nil,
			Cat{Type: "cat", Meow: "meow"},
		).
		WithOperationID("createPet")

	if rb.Options.RequestBody == nil {
		t.Fatal("expected request body to be set")
	}

	schema := rb.Options.RequestBody.Content[common.MimeJSON].Schema
	if len(schema.OneOf) != 2 {
		t.Fatalf("expected 2 oneOf schemas (nil should be skipped), got %d", len(schema.OneOf))
	}
}

func TestWithOneOfJsonBodyShouldPanicForInvalidMethods(t *testing.T) {
	tests := []struct {
		method string
	}{
		{http.MethodGet},
		{http.MethodHead},
		{http.MethodDelete},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for method %s", tt.method)
				}
			}()
			Route(tt.method, "/test").WithOneOfJsonBody(Dog{Type: "dog", Bark: "woof"})
		})
	}
}

func TestWithAnyOfJsonBodyShouldSetExample(t *testing.T) {
	dog := Dog{Type: "dog", Bark: "woof"}
	cat := Cat{Type: "cat", Meow: "meow"}

	rb := Route(http.MethodPost, "/animals").
		WithAnyOfJsonBody(dog, cat).
		WithOperationID("createAnimal")

	mediaType := rb.Options.RequestBody.Content[common.MimeJSON]
	if mediaType.Example == nil {
		t.Fatal("expected example to be set")
	}

	// Should use the first non-nil example
	if mediaType.Example != dog {
		t.Error("expected example to be the first example in the slice")
	}
}

func TestWithAllOfJsonBodyShouldRequireBody(t *testing.T) {
	type Base struct {
		ID string `json:"id"`
	}

	rb := Route(http.MethodPost, "/test").
		WithAllOfJsonBody(Base{ID: "123"}).
		WithOperationID("test")

	if !rb.Options.RequestBody.Required {
		t.Error("expected request body to be required")
	}
}
