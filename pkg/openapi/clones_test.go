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
