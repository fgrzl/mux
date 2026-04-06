package openapi

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCircularReferencesShouldNotCauseInfiniteRecursion verifies that
// types with circular references don't cause infinite recursion during
// schema generation.
func TestCircularReferencesShouldNotCauseInfiniteRecursion(t *testing.T) {
	// Arrange
	type Node struct {
		Value    string  `json:"value"`
		Children []*Node `json:"children,omitempty"`
		Parent   *Node   `json:"parent,omitempty"`
	}

	gen := NewGenerator()
	gen.ensureComponentInit()

	// Create a node with circular reference
	node := &Node{
		Value: "root",
		Children: []*Node{
			{Value: "child1"},
		},
	}
	node.Children[0].Parent = node // circular reference

	// Act - this should not hang or panic
	schema, err := gen.GenerateSchemaForType(reflect.TypeOf(Node{}))

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, schema)

	// Verify the schema was created with proper refs for the circular type
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "children")
	assert.Contains(t, schema.Properties, "parent")

	// Both children and parent should reference Node component (not inline schemas)
	childrenSchema := schema.Properties["children"]
	assert.Equal(t, "array", childrenSchema.Type)
	assert.NotNil(t, childrenSchema.Items)
	assert.Contains(t, childrenSchema.Items.Ref, "Node")

	parentSchema := schema.Properties["parent"]
	assert.Contains(t, parentSchema.Ref, "Node")
}

// TestCircularReferencesInAnyOfShouldWork verifies that circular references
// work correctly even within composite schemas like anyOf.
func TestCircularReferencesInAnyOfShouldWork(t *testing.T) {
	// Arrange
	type TreeNode struct {
		Value string    `json:"value"`
		Left  *TreeNode `json:"left,omitempty"`
		Right *TreeNode `json:"right,omitempty"`
	}

	type LinkedListNode struct {
		Data string          `json:"data"`
		Next *LinkedListNode `json:"next,omitempty"`
	}

	gen := NewGenerator()

	// Create examples with circular refs
	tree := &TreeNode{Value: "root"}
	tree.Left = tree // circular ref

	list := &LinkedListNode{Data: "first"}
	list.Next = list // circular ref

	// Create a route with anyOf containing circular types
	op := &Operation{
		OperationID: "createNode",
		RequestBody: &RequestBodyObject{
			Content: map[string]*MediaType{
				"application/json": {
					Schema: &Schema{
						AnyOf: []*Schema{
							{Ref: "#/components/schemas/TreeNode", Example: tree},
							{Ref: "#/components/schemas/LinkedListNode", Example: list},
						},
					},
					Example: tree,
				},
			},
		},
	}

	routes := []RouteData{
		{Path: "/nodes", Method: "POST", Options: op},
	}

	// Act - this should not hang or panic
	spec, err := gen.GenerateSpecFromRoutes(&InfoObject{
		Title:   "Test API",
		Version: "1.0.0",
	}, routes)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Verify both circular types were registered
	assert.Contains(t, spec.Components.Schemas, "TreeNode")
	assert.Contains(t, spec.Components.Schemas, "LinkedListNode")

	// Verify TreeNode has proper circular refs
	treeSchema := spec.Components.Schemas["TreeNode"]
	assert.NotNil(t, treeSchema)
	assert.Contains(t, treeSchema.Properties, "left")
	assert.Contains(t, treeSchema.Properties, "right")
	assert.Contains(t, treeSchema.Properties["left"].Ref, "TreeNode")
	assert.Contains(t, treeSchema.Properties["right"].Ref, "TreeNode")

	// Verify LinkedListNode has proper circular ref
	listSchema := spec.Components.Schemas["LinkedListNode"]
	assert.NotNil(t, listSchema)
	assert.Contains(t, listSchema.Properties, "next")
	assert.Contains(t, listSchema.Properties["next"].Ref, "LinkedListNode")
}
