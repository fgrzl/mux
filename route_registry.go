package mux

import (
	"fmt"
	"strings"
)

// RouteRegistry holds the root node of the trie for route registration and lookup
type RouteRegistry struct {
	root *node
}

// node represents a single route segment in the trie
type node struct {
	children map[string]*node
	options  map[string]*RouteOptions
	isParam  bool
	paramKey string
}

// newNode initializes a new trie node with empty maps for children and options
func newNode() *node {
	return &node{
		children: make(map[string]*node),
		options:  make(map[string]*RouteOptions), // Initialize the options map
	}
}

// NewRouteRegistry creates and returns a new RouteRegistry
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		root: newNode(),
	}
}

// Register registers a new route with the given pattern, method, and options
func (rr *RouteRegistry) Register(pattern string, method string, options *RouteOptions) error {
	// Prevent empty pattern
	if pattern == "" {
		return fmt.Errorf("route pattern cannot be empty")
	}

	// Split the pattern into route segments
	segments := strings.Split(strings.Trim(pattern, "/"), "/")
	current := rr.root

	// Traverse the route segments
	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") { // Dynamic parameter
			paramKey := segment[1 : len(segment)-1] // Extract parameter name
			if current.children["*param"] == nil {
				// Create a node for the parameter if it doesn't exist
				current.children["*param"] = newNode()
				current.children["*param"].isParam = true
				current.children["*param"].paramKey = paramKey
			}
			current = current.children["*param"]
		} else { // Static segment
			if current.children[segment] == nil {
				// Create a new node for static segment if it doesn't exist
				current.children[segment] = newNode()
			}
			current = current.children[segment]
		}
	}

	// Check if the route already exists for the given method
	if current.options[method] != nil {
		return fmt.Errorf("route %s %s already exists", method, pattern)
	}

	// Register the options for the given method
	current.options[method] = options
	return nil
}

// Load loads the RouteOptions for the given path and method
func (rr *RouteRegistry) Load(path string, method string) (*RouteOptions, RouteParams, bool) {
	// Split the path into segments
	segments := strings.Split(strings.Trim(path, "/"), "/")
	current := rr.root
	params := make(map[string]string)

	// Traverse the path segments
	for _, segment := range segments {
		if current.children[segment] != nil { // Static match
			current = current.children[segment]
		} else if current.children["*param"] != nil { // Parameter match
			current = current.children["*param"]
			params[current.paramKey] = segment
		} else {
			// If no match is found, return false
			return nil, nil, false
		}
	}

	options := current.options[method]
	if options == nil {
		return nil, nil, false
	}

	return options, params, true
}
