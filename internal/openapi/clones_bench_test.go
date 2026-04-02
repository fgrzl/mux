package openapi

import "testing"

var (
	benchmarkClonedInfo      *InfoObject
	benchmarkClonedOperation *Operation
)

type benchmarkCloneNested struct {
	Value string `json:"value"`
}

type benchmarkClonePayload struct {
	Name   *string               `json:"name"`
	Nested *benchmarkCloneNested `json:"nested"`
	Tags   []string              `json:"tags"`
}

func BenchmarkCloneInfoObjectRichMetadata(b *testing.B) {
	info := benchmarkInfoObject()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkClonedInfo = CloneInfoObject(info)
	}
}

func BenchmarkCloneOperationRichMetadata(b *testing.B) {
	op := benchmarkOperation()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchmarkClonedOperation = CloneOperation(op)
	}
}

func benchmarkInfoObject() *InfoObject {
	return &InfoObject{
		Title:          "API",
		Summary:        "Summary",
		Description:    "Description",
		TermsOfService: "https://example.com/terms",
		Contact: &ContactObject{
			Name:  "Support",
			URL:   "https://example.com/support",
			Email: "support@example.com",
			Extensions: map[string]any{
				"x-contact": map[string]any{"tier": "gold", "region": "us"},
			},
		},
		License: &LicenseObject{
			Name:       "MIT",
			Identifier: "MIT",
			URL:        "https://example.com/license",
			Extensions: map[string]any{
				"x-license": []string{"approved", "oss"},
			},
		},
		Version: "1.0.0",
		Extensions: map[string]any{
			"x-info": map[string]any{"env": "prod", "tier": "public"},
		},
	}
}

func benchmarkOperation() *Operation {
	name := "stable"
	payload := &benchmarkClonePayload{
		Name:   &name,
		Nested: &benchmarkCloneNested{Value: "root"},
		Tags:   []string{"one", "two"},
	}
	explode := true
	return &Operation{
		Tags:         []string{"users", "admin"},
		Summary:      "Get users",
		Description:  "Fetch users with rich OpenAPI metadata",
		ExternalDocs: &ExternalDocumentation{URL: "https://example.com/docs", Description: "User docs", Extensions: map[string]any{"x-doc": map[string]any{"audience": "public"}}},
		OperationID:  "getUsers",
		Parameters: []*ParameterObject{{
			Name:        "payload",
			In:          "query",
			Description: "pointer-backed payload",
			Explode:     &explode,
			Required:    true,
			Schema: &Schema{
				Type: "object",
				Properties: map[string]*Schema{
					"name": {Type: "string"},
					"tags": {Type: "array", Items: &Schema{Type: "string"}},
				},
				Extensions: map[string]any{
					"x-schema": map[string]any{"shape": "payload"},
				},
			},
			Example: payload,
			Extensions: map[string]*any{
				"x-param": benchmarkAnyPointer(map[string]any{"source": "bench"}),
			},
		}},
		RequestBody: &RequestBodyObject{
			Description: "request body",
			Content: map[string]*MediaType{
				"application/json": {
					Schema:  &Schema{Ref: "#/components/schemas/UserPayload", Example: payload},
					Example: payload,
					Extensions: map[string]any{
						"x-media": map[string]any{"kind": "json"},
					},
				},
			},
			Required: true,
			Extensions: map[string]any{
				"x-body": map[string]any{"mode": "required"},
			},
		},
		Responses: map[string]*ResponseObject{
			"200": {
				Description: "OK",
				Content: map[string]*MediaType{
					"application/json": {
						Schema:  &Schema{Ref: "#/components/schemas/UserPayload"},
						Example: payload,
					},
				},
				Extensions: map[string]any{
					"x-response": map[string]any{"cache": "hit"},
				},
			},
		},
		Callbacks: map[string]*PathItem{
			"onData": {
				Post: &Operation{
					OperationID: "notifyUsers",
					Parameters: []*ParameterObject{{
						Name:    "callbackPayload",
						In:      "query",
						Schema:  &Schema{Type: "string"},
						Example: payload,
					}},
					Responses: map[string]*ResponseObject{"200": {Description: "OK"}},
					Extensions: map[string]any{
						"x-callback": map[string]any{"mode": "async"},
					},
				},
				Extensions: map[string]any{
					"x-path": map[string]any{"kind": "callback"},
				},
			},
		},
		Deprecated: false,
		Security: []*SecurityRequirement{{
			"oauth2": []string{"read", "write"},
		}},
		Servers: []*ServerObject{{
			URL:         "https://api.example.com/{version}",
			Description: "Primary",
			Variables: map[string]*ServerVariable{
				"version": {
					Default:     "v1",
					Description: "API version",
					Enum:        []string{"v1", "v2"},
					Extensions: map[string]any{
						"x-server": map[string]any{"region": "us"},
					},
				},
			},
			Extensions: map[string]any{
				"x-host": map[string]any{"edge": true},
			},
		}},
		Extensions: map[string]any{
			"x-op": map[string]any{"trace": true},
		},
	}
}

func benchmarkAnyPointer(value any) *any {
	v := value
	return &v
}
