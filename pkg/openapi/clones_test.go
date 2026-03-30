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
