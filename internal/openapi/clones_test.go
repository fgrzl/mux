package openapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type clonePayloadNested struct {
	Value string `json:"value"`
}

type clonePayload struct {
	Name   *string             `json:"name"`
	Nested *clonePayloadNested `json:"nested"`
	Tags   []string            `json:"tags"`
}

type cloneNode struct {
	Value string     `json:"value"`
	Next  *cloneNode `json:"next,omitempty"`
}

func TestCloneInfoObjectShouldDeepCopyNestedMetadata(t *testing.T) {
	// Arrange
	info := &InfoObject{
		Title:   "API",
		Version: "1.0",
		Contact: &ContactObject{
			Name: "Support",
			Extensions: map[string]any{
				"x-contact": map[string]any{"tier": "gold"},
			},
		},
		License: &LicenseObject{
			Name: "MIT",
			Extensions: map[string]any{
				"x-license": []string{"approved"},
			},
		},
		Extensions: map[string]any{
			"x-info": map[string]any{"env": "prod"},
		},
	}

	// Act
	clone := CloneInfoObject(info)
	clone.Title = "Changed"
	clone.Contact.Name = "Changed Contact"
	clone.Contact.Extensions["x-contact"].(map[string]any)["tier"] = "silver"
	clone.License.Name = "Changed License"
	clone.License.Extensions["x-license"].([]string)[0] = "rejected"
	clone.Extensions["x-info"].(map[string]any)["env"] = "staging"

	// Assert
	require.NotNil(t, clone)
	assert.NotSame(t, info, clone)
	assert.NotSame(t, info.Contact, clone.Contact)
	assert.NotSame(t, info.License, clone.License)
	assert.Equal(t, "API", info.Title)
	assert.Equal(t, "Support", info.Contact.Name)
	assert.Equal(t, "gold", info.Contact.Extensions["x-contact"].(map[string]any)["tier"])
	assert.Equal(t, "MIT", info.License.Name)
	assert.Equal(t, []string{"approved"}, info.License.Extensions["x-license"].([]string))
	assert.Equal(t, "prod", info.Extensions["x-info"].(map[string]any)["env"])
}

func TestCloneParameterObjectShouldDeepCopyPointerBackedExample(t *testing.T) {
	// Arrange
	name := "Alice"
	param := &ParameterObject{
		Example: &clonePayload{
			Name:   &name,
			Nested: &clonePayloadNested{Value: "root"},
			Tags:   []string{"one"},
		},
	}

	// Act
	clone := CloneParameterObject(param)
	clonedPayload := clone.Example.(*clonePayload)
	*clonedPayload.Name = "Bob"
	clonedPayload.Nested.Value = "child"
	clonedPayload.Tags[0] = "two"

	// Assert
	originalPayload := param.Example.(*clonePayload)
	require.NotNil(t, clone)
	assert.NotSame(t, originalPayload, clonedPayload)
	assert.NotSame(t, originalPayload.Name, clonedPayload.Name)
	assert.NotSame(t, originalPayload.Nested, clonedPayload.Nested)
	assert.Equal(t, "Alice", *originalPayload.Name)
	assert.Equal(t, "root", originalPayload.Nested.Value)
	assert.Equal(t, []string{"one"}, originalPayload.Tags)
}

func TestCloneSchemaShouldDeepCopyPointerBackedDefaultAndEnumValues(t *testing.T) {
	// Arrange
	defaultName := "Default"
	enumName := "Enum"
	schema := &Schema{
		Default: &clonePayload{
			Name:   &defaultName,
			Nested: &clonePayloadNested{Value: "default"},
			Tags:   []string{"one"},
		},
		Enum: []any{
			&clonePayload{
				Name:   &enumName,
				Nested: &clonePayloadNested{Value: "enum"},
				Tags:   []string{"two"},
			},
		},
	}

	// Act
	clone := CloneSchema(schema)
	clonedDefault := clone.Default.(*clonePayload)
	clonedEnum := clone.Enum[0].(*clonePayload)
	*clonedDefault.Name = "Changed"
	clonedDefault.Tags[0] = "changed"
	*clonedEnum.Name = "Changed Enum"
	clonedEnum.Nested.Value = "changed enum"

	// Assert
	originalDefault := schema.Default.(*clonePayload)
	originalEnum := schema.Enum[0].(*clonePayload)
	require.NotNil(t, clone)
	assert.NotSame(t, originalDefault, clonedDefault)
	assert.NotSame(t, originalDefault.Name, clonedDefault.Name)
	assert.NotSame(t, originalEnum, clonedEnum)
	assert.NotSame(t, originalEnum.Nested, clonedEnum.Nested)
	assert.Equal(t, "Default", *originalDefault.Name)
	assert.Equal(t, []string{"one"}, originalDefault.Tags)
	assert.Equal(t, "Enum", *originalEnum.Name)
	assert.Equal(t, "enum", originalEnum.Nested.Value)
}

func TestCloneParameterObjectShouldPreserveCircularPointerExamples(t *testing.T) {
	// Arrange
	root := &cloneNode{Value: "root"}
	root.Next = root
	param := &ParameterObject{Example: root}

	// Act
	clone := CloneParameterObject(param)
	clonedRoot := clone.Example.(*cloneNode)
	clonedRoot.Value = "cloned"

	// Assert
	require.NotNil(t, clone)
	assert.NotSame(t, root, clonedRoot)
	assert.Same(t, clonedRoot, clonedRoot.Next)
	assert.Equal(t, "root", root.Value)
	assert.Equal(t, "cloned", clonedRoot.Value)
}

func TestCloneOperationShouldDeepCopyCallbacksServersAndExtensions(t *testing.T) {
	// Arrange
	name := "stable"
	payload := &clonePayload{
		Name:   &name,
		Nested: &clonePayloadNested{Value: "root"},
		Tags:   []string{"one"},
	}
	op := &Operation{
		OperationID:  "getUsers",
		ExternalDocs: &ExternalDocumentation{URL: "https://example.com/docs", Description: "User docs", Extensions: map[string]any{"x-doc": map[string]any{"audience": "public"}}},
		Callbacks: map[string]*PathItem{
			"onData": {
				Post: &Operation{
					OperationID: "notifyUsers",
					Parameters: []*ParameterObject{{
						Name:    "payload",
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
		Servers: []*ServerObject{{
			URL: "https://api.example.com/{version}",
			Variables: map[string]*ServerVariable{
				"version": {
					Default: "v1",
					Enum:    []string{"v1"},
					Extensions: map[string]any{
						"x-server": map[string]any{"region": "us"},
					},
				},
			},
			Extensions: map[string]any{
				"x-host": map[string]any{"edge": true},
			},
		}},
		Responses: map[string]*ResponseObject{"200": {Description: "OK"}},
		Extensions: map[string]any{
			"x-op": map[string]any{"trace": true},
		},
	}

	// Act
	clone := CloneOperation(op)
	clone.ExternalDocs.Description = "Changed docs"
	clone.ExternalDocs.Extensions["x-doc"].(map[string]any)["audience"] = "private"
	clone.Callbacks["onData"].Post.Extensions["x-callback"].(map[string]any)["mode"] = "sync"
	clonedPayload := clone.Callbacks["onData"].Post.Parameters[0].Example.(*clonePayload)
	*clonedPayload.Name = "changed"
	clonedPayload.Nested.Value = "mutated"
	clonedPayload.Tags[0] = "two"
	clone.Servers[0].Variables["version"].Default = "v2"
	clone.Servers[0].Variables["version"].Extensions["x-server"].(map[string]any)["region"] = "eu"
	clone.Servers[0].Extensions["x-host"].(map[string]any)["edge"] = false
	clone.Extensions["x-op"].(map[string]any)["trace"] = false

	// Assert
	require.NotNil(t, clone)
	assert.NotSame(t, op, clone)
	assert.NotSame(t, op.ExternalDocs, clone.ExternalDocs)
	assert.NotSame(t, op.Callbacks["onData"], clone.Callbacks["onData"])
	assert.NotSame(t, op.Callbacks["onData"].Post, clone.Callbacks["onData"].Post)
	assert.NotSame(t, op.Servers[0], clone.Servers[0])
	assert.Equal(t, "User docs", op.ExternalDocs.Description)
	assert.Equal(t, "public", op.ExternalDocs.Extensions["x-doc"].(map[string]any)["audience"])
	assert.Equal(t, "async", op.Callbacks["onData"].Post.Extensions["x-callback"].(map[string]any)["mode"])
	originalPayload := op.Callbacks["onData"].Post.Parameters[0].Example.(*clonePayload)
	assert.Equal(t, "stable", *originalPayload.Name)
	assert.Equal(t, "root", originalPayload.Nested.Value)
	assert.Equal(t, []string{"one"}, originalPayload.Tags)
	assert.Equal(t, "v1", op.Servers[0].Variables["version"].Default)
	assert.Equal(t, "us", op.Servers[0].Variables["version"].Extensions["x-server"].(map[string]any)["region"])
	assert.Equal(t, true, op.Servers[0].Extensions["x-host"].(map[string]any)["edge"])
	assert.Equal(t, true, op.Extensions["x-op"].(map[string]any)["trace"])
}

func TestCloneOperationShouldPreserveEmptyTagsAndExtensions(t *testing.T) {
	// Arrange
	op := &Operation{
		Tags:       []string{},
		Extensions: map[string]any{},
	}

	// Act
	clone := CloneOperation(op)
	clone.Extensions["x-op"] = true

	// Assert
	require.NotNil(t, clone)
	require.NotNil(t, clone.Tags)
	assert.Empty(t, clone.Tags)
	require.NotNil(t, clone.Extensions)
	assert.Empty(t, op.Extensions)
	assert.Equal(t, true, clone.Extensions["x-op"])
}

func TestCloneOperationShouldDeepCopyEmptyCollections(t *testing.T) {
	// Arrange
	op := &Operation{
		Parameters: make([]*ParameterObject, 0, 1),
		Responses:  map[string]*ResponseObject{},
		Callbacks:  map[string]*PathItem{},
		Security:   make([]*SecurityRequirement, 0, 1),
		Servers:    make([]*ServerObject, 0, 1),
	}

	// Act
	clone := CloneOperation(op)
	clone.Parameters = append(clone.Parameters, &ParameterObject{Name: "clone-param"})
	clone.Responses["200"] = &ResponseObject{Description: "clone-response"}
	clone.Callbacks["clone"] = &PathItem{}
	clone.Security = append(clone.Security, &SecurityRequirement{})
	clone.Servers = append(clone.Servers, &ServerObject{URL: "https://clone.example.com"})

	op.Parameters = append(op.Parameters, &ParameterObject{Name: "source-param"})
	op.Security = append(op.Security, &SecurityRequirement{})
	op.Servers = append(op.Servers, &ServerObject{URL: "https://source.example.com"})

	// Assert
	require.NotNil(t, clone)
	require.NotNil(t, clone.Parameters)
	require.NotNil(t, clone.Responses)
	require.NotNil(t, clone.Callbacks)
	require.NotNil(t, clone.Security)
	require.NotNil(t, clone.Servers)
	require.Len(t, clone.Parameters, 1)
	require.Len(t, op.Parameters, 1)
	assert.Equal(t, "clone-param", clone.Parameters[0].Name)
	assert.Equal(t, "source-param", op.Parameters[0].Name)
	assert.Contains(t, clone.Responses, "200")
	assert.Empty(t, op.Responses)
	assert.Contains(t, clone.Callbacks, "clone")
	assert.Empty(t, op.Callbacks)
	require.Len(t, clone.Security, 1)
	require.Len(t, op.Security, 1)
	require.Len(t, clone.Servers, 1)
	require.Len(t, op.Servers, 1)
	assert.Equal(t, "https://clone.example.com", clone.Servers[0].URL)
	assert.Equal(t, "https://source.example.com", op.Servers[0].URL)
}

func TestCloneOperationShouldDeepCopyNestedEmptyCollections(t *testing.T) {
	// Arrange
	callback := &PathItem{
		Parameters: make([]*ParameterObject, 0, 1),
		Servers:    make([]*ServerObject, 0, 1),
		Extensions: map[string]any{},
	}
	operationSecurity := SecurityRequirement{}
	op := &Operation{
		Parameters: []*ParameterObject{{
			Name:       "filter",
			In:         "query",
			Examples:   map[string]*ExampleObject{},
			Content:    map[string]*MediaType{},
			Extensions: map[string]*any{},
			Schema: &Schema{
				Properties:    map[string]*Schema{},
				Required:      make([]string, 0, 1),
				Enum:          make([]any, 0, 1),
				OneOf:         make([]*Schema, 0, 1),
				AnyOf:         make([]*Schema, 0, 1),
				AllOf:         make([]*Schema, 0, 1),
				Discriminator: &Discriminator{PropertyName: "kind", Mapping: map[string]string{}},
				Extensions:    map[string]any{},
			},
		}},
		Responses: map[string]*ResponseObject{
			"200": {
				Description: "OK",
				Headers:     map[string]*HeaderObject{},
				Content: map[string]*MediaType{
					"application/json": {
						Examples: map[string]*ExampleObject{},
						Encoding: map[string]*EncodingObject{
							"payload": {
								Headers:    map[string]*HeaderObject{},
								Extensions: map[string]any{},
							},
						},
						Extensions: map[string]any{},
					},
				},
				Links:      map[string]*LinkObject{},
				Extensions: map[string]any{},
			},
		},
		Callbacks: map[string]*PathItem{"cb": callback},
		Security:  []*SecurityRequirement{&operationSecurity},
		Servers: []*ServerObject{{
			URL: "https://api.example.com",
			Variables: map[string]*ServerVariable{
				"stage": {
					Default:    "prod",
					Enum:       make([]string, 0, 1),
					Extensions: map[string]any{},
				},
			},
			Extensions: map[string]any{},
		}},
	}

	// Act
	clone := CloneOperation(op)
	cloneParam := clone.Parameters[0]
	cloneSchema := cloneParam.Schema
	cloneResponse := clone.Responses["200"]
	cloneMedia := cloneResponse.Content["application/json"]
	cloneEncoding := cloneMedia.Encoding["payload"]
	cloneCallback := clone.Callbacks["cb"]
	cloneServerVar := clone.Servers[0].Variables["stage"]

	cloneParam.Examples["clone"] = &ExampleObject{Summary: "clone"}
	cloneParam.Content["application/json"] = &MediaType{}
	cloneParam.Extensions["x-clone"] = nil
	cloneSchema.Properties["clone"] = &Schema{Type: "string"}
	cloneSchema.Required = append(cloneSchema.Required, "clone")
	cloneSchema.Enum = append(cloneSchema.Enum, "clone")
	cloneSchema.OneOf = append(cloneSchema.OneOf, &Schema{Type: "string"})
	cloneSchema.AnyOf = append(cloneSchema.AnyOf, &Schema{Type: "number"})
	cloneSchema.AllOf = append(cloneSchema.AllOf, &Schema{Type: "boolean"})
	cloneSchema.Discriminator.Mapping["clone"] = "#/components/schemas/Clone"
	cloneSchema.Extensions["x-clone"] = true
	cloneResponse.Headers["X-Clone"] = &HeaderObject{Description: "clone"}
	cloneResponse.Links["clone"] = &LinkObject{Description: "clone"}
	cloneResponse.Extensions["x-clone"] = true
	cloneMedia.Examples["clone"] = &ExampleObject{Summary: "clone"}
	cloneMedia.Extensions["x-clone"] = true
	cloneEncoding.Headers["X-Clone"] = &HeaderObject{Description: "clone"}
	cloneEncoding.Extensions["x-clone"] = true
	cloneCallback.Parameters = append(cloneCallback.Parameters, &ParameterObject{Name: "clone-callback"})
	cloneCallback.Servers = append(cloneCallback.Servers, &ServerObject{URL: "https://clone-callback.example.com"})
	cloneCallback.Extensions["x-clone"] = true
	cloneServerVar.Enum = append(cloneServerVar.Enum, "clone")
	cloneServerVar.Extensions["x-clone"] = true

	op.Parameters[0].Examples["source"] = &ExampleObject{Summary: "source"}
	op.Parameters[0].Content["text/plain"] = &MediaType{}
	op.Parameters[0].Extensions["x-source"] = nil
	op.Parameters[0].Schema.Properties["source"] = &Schema{Type: "integer"}
	op.Parameters[0].Schema.Required = append(op.Parameters[0].Schema.Required, "source")
	op.Parameters[0].Schema.Enum = append(op.Parameters[0].Schema.Enum, "source")
	op.Parameters[0].Schema.OneOf = append(op.Parameters[0].Schema.OneOf, &Schema{Type: "integer"})
	op.Parameters[0].Schema.AnyOf = append(op.Parameters[0].Schema.AnyOf, &Schema{Type: "object"})
	op.Parameters[0].Schema.AllOf = append(op.Parameters[0].Schema.AllOf, &Schema{Type: "array"})
	op.Parameters[0].Schema.Discriminator.Mapping["source"] = "#/components/schemas/Source"
	op.Parameters[0].Schema.Extensions["x-source"] = true
	op.Responses["200"].Headers["X-Source"] = &HeaderObject{Description: "source"}
	op.Responses["200"].Links["source"] = &LinkObject{Description: "source"}
	op.Responses["200"].Extensions["x-source"] = true
	op.Responses["200"].Content["application/json"].Examples["source"] = &ExampleObject{Summary: "source"}
	op.Responses["200"].Content["application/json"].Extensions["x-source"] = true
	op.Responses["200"].Content["application/json"].Encoding["payload"].Headers["X-Source"] = &HeaderObject{Description: "source"}
	op.Responses["200"].Content["application/json"].Encoding["payload"].Extensions["x-source"] = true
	op.Callbacks["cb"].Parameters = append(op.Callbacks["cb"].Parameters, &ParameterObject{Name: "source-callback"})
	op.Callbacks["cb"].Servers = append(op.Callbacks["cb"].Servers, &ServerObject{URL: "https://source-callback.example.com"})
	op.Callbacks["cb"].Extensions["x-source"] = true
	op.Servers[0].Variables["stage"].Enum = append(op.Servers[0].Variables["stage"].Enum, "source")
	op.Servers[0].Variables["stage"].Extensions["x-source"] = true

	// Assert
	assert.Contains(t, cloneParam.Examples, "clone")
	assert.NotContains(t, op.Parameters[0].Examples, "clone")
	assert.Contains(t, op.Parameters[0].Examples, "source")
	assert.NotContains(t, cloneParam.Examples, "source")
	assert.Contains(t, cloneParam.Content, "application/json")
	assert.NotContains(t, op.Parameters[0].Content, "application/json")
	assert.Contains(t, op.Parameters[0].Content, "text/plain")
	assert.NotContains(t, cloneParam.Content, "text/plain")
	assert.Contains(t, cloneParam.Extensions, "x-clone")
	assert.NotContains(t, op.Parameters[0].Extensions, "x-clone")
	assert.Contains(t, op.Parameters[0].Schema.Properties, "source")
	assert.NotContains(t, cloneSchema.Properties, "source")
	assert.Equal(t, []string{"clone"}, cloneSchema.Required)
	assert.Equal(t, []string{"source"}, op.Parameters[0].Schema.Required)
	assert.Equal(t, []any{"clone"}, cloneSchema.Enum)
	assert.Equal(t, []any{"source"}, op.Parameters[0].Schema.Enum)
	assert.Len(t, cloneSchema.OneOf, 1)
	assert.Len(t, op.Parameters[0].Schema.OneOf, 1)
	assert.Equal(t, "string", cloneSchema.OneOf[0].Type)
	assert.Equal(t, "integer", op.Parameters[0].Schema.OneOf[0].Type)
	assert.Equal(t, "#/components/schemas/Clone", cloneSchema.Discriminator.Mapping["clone"])
	assert.NotContains(t, op.Parameters[0].Schema.Discriminator.Mapping, "clone")
	assert.Contains(t, cloneResponse.Headers, "X-Clone")
	assert.NotContains(t, op.Responses["200"].Headers, "X-Clone")
	assert.Contains(t, cloneResponse.Links, "clone")
	assert.NotContains(t, op.Responses["200"].Links, "clone")
	assert.Contains(t, cloneMedia.Examples, "clone")
	assert.NotContains(t, op.Responses["200"].Content["application/json"].Examples, "clone")
	assert.Contains(t, cloneEncoding.Headers, "X-Clone")
	assert.NotContains(t, op.Responses["200"].Content["application/json"].Encoding["payload"].Headers, "X-Clone")
	require.Len(t, cloneCallback.Parameters, 1)
	require.Len(t, op.Callbacks["cb"].Parameters, 1)
	assert.Equal(t, "clone-callback", cloneCallback.Parameters[0].Name)
	assert.Equal(t, "source-callback", op.Callbacks["cb"].Parameters[0].Name)
	require.Len(t, cloneCallback.Servers, 1)
	require.Len(t, op.Callbacks["cb"].Servers, 1)
	assert.Equal(t, "https://clone-callback.example.com", cloneCallback.Servers[0].URL)
	assert.Equal(t, "https://source-callback.example.com", op.Callbacks["cb"].Servers[0].URL)
	assert.Equal(t, []string{"clone"}, cloneServerVar.Enum)
	assert.Equal(t, []string{"source"}, op.Servers[0].Variables["stage"].Enum)
	assert.Contains(t, cloneServerVar.Extensions, "x-clone")
	assert.NotContains(t, op.Servers[0].Variables["stage"].Extensions, "x-clone")
}

func TestCloneSecurityRequirementShouldPreserveNilAndEmptyMaps(t *testing.T) {
	// Arrange
	var nilRequirement SecurityRequirement
	emptyRequirement := SecurityRequirement{}

	// Act
	nilClone := CloneSecurityRequirement(&nilRequirement)
	emptyClone := CloneSecurityRequirement(&emptyRequirement)

	// Assert
	require.NotNil(t, nilClone)
	assert.Nil(t, *nilClone)
	require.NotNil(t, emptyClone)
	require.NotNil(t, *emptyClone)
	assert.Empty(t, *emptyClone)
}

func TestCloneOperationShouldPreserveCallbackCycles(t *testing.T) {
	// Arrange
	op := &Operation{
		OperationID: "root",
		Responses:   map[string]*ResponseObject{"200": {Description: "OK"}},
	}
	callback := &PathItem{}
	op.Callbacks = map[string]*PathItem{"self": callback}
	callback.Post = op

	// Act
	clone := CloneOperation(op)
	clone.OperationID = "cloned"

	// Assert
	require.NotNil(t, clone)
	assert.NotSame(t, op, clone)
	require.Contains(t, clone.Callbacks, "self")
	assert.NotSame(t, callback, clone.Callbacks["self"])
	assert.Same(t, clone, clone.Callbacks["self"].Post)
	assert.Equal(t, "root", op.OperationID)
	assert.Equal(t, "cloned", clone.OperationID)
}
