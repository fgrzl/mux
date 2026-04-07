package registry

import (
	"strings"

	"github.com/fgrzl/mux/internal/routing"
)

// RouteRegistry holds the routing trie and auxiliary fast-path maps used
// to register and look up routes. It stores a root RouteNode and an
// exactRoutes map used for fully static routes to provide a fast-path
// for lookups that don't involve parameters or wildcards.
type RouteRegistry struct {
	root *routing.RouteNode
	// exactRoutes provides a fast-path for fully static routes (no params or wildcards).
	// Keyed by the pattern (as registered) then method.
	exactRoutes map[string]map[string]*routing.RouteOptions
}

// LoadDetails provides additional information about a route lookup.
// Found indicates whether a path matched a terminal node. MethodOK indicates
// whether the requested method is allowed at that node. If Found is true and
// MethodOK is false, Allow contains the precomputed Allow header value.
type LoadDetails struct {
	Found    bool
	MethodOK bool
	Allow    string
}

// NewRouteRegistry creates and initializes an empty RouteRegistry.
// The returned registry contains a root RouteNode and an initialized
// exactRoutes map for static route fast-paths.
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		root:        &routing.RouteNode{Children: make(map[string]*routing.RouteNode)},
		exactRoutes: make(map[string]map[string]*routing.RouteOptions),
	}
}

func newRouteNode() *routing.RouteNode {
	return &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
}

func splitSegments(trimmed string) []string {
	if trimmed == "" {
		return nil
	}
	raw := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(raw))
	for _, s := range raw {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

func isParamSegment(seg string) bool {
	return strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") && len(seg) > 2
}

// Note: helper functions for splitting/scanning path segments were removed
// as they were unused and the registry performs in-place scanning for
// performance and clarity.

// Root returns the trie root node for this registry. Callers may use
// the returned node for inspection or debugging; modifications to the
// node are not recommended outside of the registry methods.
// Root returns the trie root node for this registry. Callers may use
// the returned node for inspection or debugging; modifications to the
// node are not recommended outside of the registry methods.
func (r *RouteRegistry) Root() *routing.RouteNode {
	return r.root
}

// LoadExact performs a fast exact-match lookup for static routes without any
// trie traversal. Returns the RouteOptions and true if an exact match is found
// for both the path and method. This is the fastest possible lookup path.
func (r *RouteRegistry) LoadExact(path string, method string) (*routing.RouteOptions, bool) {
	if m, ok := r.exactRoutes[path]; ok {
		if opt, ok := m[method]; ok {
			return opt, true
		}
	}
	return nil, false
}

// Register adds a route for the given pattern and HTTP method to the
// registry. The pattern may contain parameter tokens ({name}), a single
// segment wildcard (*), or a catch-all (**) at the end. The provided
// options are stored and per-node metadata (MethodsMask, AllowHeader)
// is updated to accelerate lookups.
// Register adds a route for the given pattern and HTTP method to the
// registry. The pattern may contain parameter tokens ({name}), a single
// segment wildcard (*), or a catch-all (**) at the end. The provided
// options are stored and per-node metadata (MethodsMask, AllowHeader)
// is updated to accelerate lookups.
func (r *RouteRegistry) Register(pattern string, method string, options *routing.RouteOptions) {
	trimmed := strings.Trim(pattern, "/")
	if trimmed == "" {
		r.registerRootRoute(pattern, method, options)
		return
	}

	segments := splitSegments(trimmed)
	node, hasParams, paramCount := r.walkOrCreateNodes(segments)
	r.assignRouteOptions(node, method, options, hasParams, paramCount)

	if !strings.ContainsAny(pattern, "{*}") {
		r.storeExactRoute(pattern, method, options)
	}
}

func (r *RouteRegistry) registerRootRoute(pattern, method string, options *routing.RouteOptions) {
	r.assignRouteOptions(r.root, method, options, false, 0)
	if !strings.ContainsAny(pattern, "{*}") {
		r.storeExactRoute(pattern, method, options)
	}
}

// HasRoute reports whether the given registered pattern already has a handler
// for method.
func (r *RouteRegistry) HasRoute(pattern, method string) bool {
	node := r.findRegisteredNode(pattern)
	if node == nil || node.RouteOptions == nil {
		return false
	}
	_, ok := node.RouteOptions[method]
	return ok
}

// Unregister removes the handler for the given registered pattern and method.
// It returns true when a route was removed.
func (r *RouteRegistry) Unregister(pattern, method string) bool {
	node := r.findRegisteredNode(pattern)
	if node == nil || node.RouteOptions == nil {
		return false
	}
	if _, ok := node.RouteOptions[method]; !ok {
		return false
	}

	delete(node.RouteOptions, method)
	if len(node.RouteOptions) == 0 {
		node.RouteOptions = nil
		node.MethodsMask = 0
		node.AllowHeader = ""
	} else {
		node.MethodsMask = methodsMaskFromMap(node.RouteOptions)
		node.AllowHeader = allowHeaderFromMap(node.RouteOptions)
	}

	if m, ok := r.exactRoutes[pattern]; ok {
		delete(m, method)
		if len(m) == 0 {
			delete(r.exactRoutes, pattern)
		}
	}

	return true
}

func (r *RouteRegistry) findRegisteredNode(pattern string) *routing.RouteNode {
	trimmed := strings.Trim(pattern, "/")
	if trimmed == "" {
		return r.root
	}

	node := r.root
	for _, seg := range splitSegments(trimmed) {
		switch {
		case seg == "**":
			node = node.CatchAll
		case seg == "*":
			node = node.Wildcard
		case isParamSegment(seg):
			node = node.ParamChild
		default:
			node = node.Children[seg]
		}
		if node == nil {
			return nil
		}
	}

	return node
}

func (r *RouteRegistry) walkOrCreateNodes(segments []string) (*routing.RouteNode, bool, int) {
	node := r.root
	hasParams := false
	paramCount := 0
	for _, seg := range segments {
		next, updatedHasParams, count, done := r.advanceNode(node, seg, hasParams, paramCount)
		node = next
		hasParams = updatedHasParams
		paramCount = count
		if done {
			break
		}
	}
	return node, hasParams, paramCount
}

func (r *RouteRegistry) advanceNode(node *routing.RouteNode, seg string, hasParams bool, paramCount int) (*routing.RouteNode, bool, int, bool) {
	switch {
	case seg == "**":
		if node.CatchAll == nil {
			node.CatchAll = newRouteNode()
		}
		r.refreshFastPathFlags(node)
		return node.CatchAll, hasParams, paramCount, true
	case seg == "*":
		if node.Wildcard == nil {
			node.Wildcard = newRouteNode()
		}
		r.refreshFastPathFlags(node)
		return node.Wildcard, hasParams, paramCount, false
	case isParamSegment(seg):
		hasParams = true
		paramCount++
		if node.ParamChild == nil {
			node.ParamChild = newRouteNode()
			// Use interned string for common parameter names to reduce allocations
			paramName := seg[1 : len(seg)-1]
			node.ParamChild.ParamName = routing.InternString(paramName)
		}
		r.refreshFastPathFlags(node)
		return node.ParamChild, hasParams, paramCount, false
	default:
		if node.Children[seg] == nil {
			node.Children[seg] = newRouteNode()
		}
		r.refreshFastPathFlags(node)
		return node.Children[seg], hasParams, paramCount, false
	}
}

func (r *RouteRegistry) assignRouteOptions(node *routing.RouteNode, method string, options *routing.RouteOptions, hasParams bool, paramCount int) {
	if node.RouteOptions == nil {
		node.RouteOptions = make(map[string]*routing.RouteOptions)
	}
	node.RouteOptions[method] = options
	if hasParams {
		node.HasParams = true
		node.ParamCount = paramCount
	}
	node.MethodsMask = methodsMaskFromMap(node.RouteOptions)
	node.AllowHeader = allowHeaderFromMap(node.RouteOptions)
}

func (r *RouteRegistry) storeExactRoute(pattern, method string, options *routing.RouteOptions) {
	m := r.exactRoutes[pattern]
	if m == nil {
		m = make(map[string]*routing.RouteOptions)
		r.exactRoutes[pattern] = m
	}
	m[method] = options
}

// refreshFastPathFlags recomputes fast-path flags for n based on its children pointers.
// refreshFastPathFlags recomputes fast-path flags for n based on its children pointers.
func (r *RouteRegistry) refreshFastPathFlags(n *routing.RouteNode) {
	if n == nil {
		return
	}
	// HasOnlyCatchAll when there is a CatchAll and no other possible next step
	if n.CatchAll != nil && len(n.Children) == 0 && n.ParamChild == nil && n.Wildcard == nil {
		n.HasOnlyCatchAll = true
	} else {
		n.HasOnlyCatchAll = false
	}
	// HasOnlyWildcardTerminal when there is a Wildcard and no other next step, and the wildcard node
	// is a terminal for some method (i.e., has RouteOptions). This allows short-circuiting patterns like /files/*.
	if n.Wildcard != nil && len(n.Children) == 0 && n.ParamChild == nil && n.CatchAll == nil {
		n.HasOnlyWildcardTerminal = len(n.Wildcard.RouteOptions) > 0
	} else {
		n.HasOnlyWildcardTerminal = false
	}
}

// trimPathIndices returns the start and end indices of path with leading and
// trailing slashes ignored. This avoids repeating the same trimming logic in
// multiple traversal functions.
func trimPathIndices(path string) (int, int) {
	start := 0
	end := len(path)
	for start < end && path[start] == '/' {
		start++
	}
	for end > start && path[end-1] == '/' {
		end--
	}
	return start, end
}

// scanSegment scans path starting at s up to end and returns the segment and
// the index to continue from (just after the next '/').
func scanSegment(path string, s, end int) (seg string, nextIndex int) {
	j := s
	for j < end && path[j] != '/' {
		j++
	}
	seg = path[s:j]
	nextIndex = j + 1
	return
}

// chooseNextEdge selects the next route node for the given segment according
// to the registry precedence: static child, param child, wildcard, catch-all.
// It returns the selected node and a string describing the edge type.
func chooseNextEdge(n *routing.RouteNode, seg string) (*routing.RouteNode, string) {
	if child, ok := n.Children[seg]; ok {
		return child, "child"
	}
	if n.ParamChild != nil {
		return n.ParamChild, "param"
	}
	if n.Wildcard != nil {
		return n.Wildcard, "wildcard"
	}
	if n.CatchAll != nil {
		return n.CatchAll, "catchall"
	}
	return nil, ""
}

// LoadDetailedIntoSlice performs a route lookup, fills any path params into dst
// (resetting it first), and returns the matched RouteOptions along with
// LoadDetails describing the match.
func (r *RouteRegistry) LoadDetailedIntoSlice(path string, method string, dst *routing.Params) (*routing.RouteOptions, LoadDetails) {
	// Exact static fast-path
	if m, ok := r.exactRoutes[path]; ok {
		if dst != nil {
			dst.Reset()
		}
		details := LoadDetails{Found: true}
		if opt, ok2 := m[method]; ok2 {
			details.MethodOK = true
			return opt, details
		}
		details.Allow = allowHeaderFromMap(m)
		return nil, details
	}
	// Match a node and collect params
	node, matched := r.matchNodeIntoSlice(path, dst)
	if !matched || node == nil || len(node.RouteOptions) == 0 {
		return nil, LoadDetails{Found: false}
	}
	if opt, ok := node.RouteOptions[method]; ok {
		return opt, LoadDetails{Found: true, MethodOK: true}
	}
	return nil, LoadDetails{Found: true, MethodOK: false, Allow: node.AllowHeader}
}

// matchNodeIntoSlice traverses the registry and populates a Params slice,
// avoiding hash computation and map allocation overhead.
// The dst slice is reset (length set to 0) before populating.
// This version inlines critical path operations to reduce function call overhead.
func (r *RouteRegistry) matchNodeIntoSlice(path string, dst *routing.Params) (*routing.RouteNode, bool) {
	if dst != nil {
		dst.Reset()
	}

	// Inline trimPathIndices for hot path
	start := 0
	end := len(path)
	for start < end && path[start] == '/' {
		start++
	}
	for end > start && path[end-1] == '/' {
		end--
	}

	n := r.root
	s := start

	for s < end {
		// Early short-circuits using precomputed flags
		if n.HasOnlyCatchAll {
			return n.CatchAll, true
		}
		if n.HasOnlyWildcardTerminal {
			return n.Wildcard, true
		}

		// Inline scanSegment - find next '/' or end
		j := s
		for j < end && path[j] != '/' {
			j++
		}
		seg := path[s:j]

		// Inline chooseNextEdge with precedence: static > param > wildcard > catch-all
		var next *routing.RouteNode

		if child, ok := n.Children[seg]; ok {
			// Static child match (most common case)
			n = child
		} else if n.ParamChild != nil {
			// Parameter match
			next = n.ParamChild
			if dst != nil {
				// Append directly to slice - much faster than map insertion
				*dst = append(*dst, routing.Param{Key: next.ParamName, Value: seg})
			}
			n = next
		} else if n.Wildcard != nil {
			// Wildcard match
			n = n.Wildcard
		} else if n.CatchAll != nil {
			// Catch-all match - consumes remainder
			return n.CatchAll, true
		} else {
			// No match found
			if dst != nil {
				dst.Reset()
			}
			return nil, false
		}

		// Advance to next segment (skip '/')
		s = j + 1
	}

	return n, true
}

// LoadIntoSlice performs a route lookup and writes any extracted path
// parameters into the provided slice. The destination slice is reset before
// populating. Returns the matched RouteOptions and ok=true when found.
func (r *RouteRegistry) LoadIntoSlice(path string, method string, dst *routing.Params) (*routing.RouteOptions, bool) {
	// Fast path: exact registered static route (no params to extract)
	if m, ok := r.exactRoutes[path]; ok {
		if opt, ok2 := m[method]; ok2 {
			// Ensure dst is cleared
			if dst != nil {
				dst.Reset()
			}
			return opt, true
		}
	}
	// Match with parameter extraction
	node, matched := r.matchNodeIntoSlice(path, dst)
	if !matched || node == nil || len(node.RouteOptions) == 0 {
		return nil, false
	}
	if opt, ok := node.RouteOptions[method]; ok {
		return opt, true
	}
	return nil, false
}

// TryMatchMethods returns the list of allowed HTTP methods for a given path
// if the path matches any registered route. If the path does not match, ok=false.
// TryMatchMethods returns the list of allowed HTTP methods for a given path
// if the path matches any registered route. If the path does not match, ok=false.
func (r *RouteRegistry) TryMatchMethods(path string) (methods []string, ok bool) {
	// exact fast-path for static routes
	if m, ok := r.exactRoutes[path]; ok {
		out := make([]string, 0, len(m))
		for method := range m {
			out = append(out, method)
		}
		return out, true
	}
	// Single traversal to locate the terminal node for this path
	node := r.findNode(path)
	if node == nil || len(node.RouteOptions) == 0 {
		return nil, false
	}
	out := make([]string, 0, len(node.RouteOptions))
	for method := range node.RouteOptions {
		out = append(out, method)
	}
	return out, true
}

// TryGetAllowHeader returns a precomputed Allow header value for a matched path.
// If no route matches the path, ok=false.
// TryGetAllowHeader returns a precomputed Allow header value for a matched path.
// If no route matches the path, ok=false.
func (r *RouteRegistry) TryGetAllowHeader(path string) (string, bool) {
	if m, ok := r.exactRoutes[path]; ok {
		return allowHeaderFromMap(m), true
	}
	node := r.findNode(path)
	if node == nil || len(node.RouteOptions) == 0 {
		return "", false
	}
	if node.AllowHeader != "" {
		return node.AllowHeader, true
	}
	// Fallback compute if not cached (should not happen normally)
	return allowHeaderFromMap(node.RouteOptions), true
}

// methodsMaskFromMap builds a bitmask for common HTTP methods from a map key set.
// methodsMaskFromMap builds a bitmask for common HTTP methods from a map key set.
func methodsMaskFromMap(m map[string]*routing.RouteOptions) uint32 {
	var mask uint32
	for method := range m {
		switch method {
		case "GET":
			mask |= 1 << 0
		case "POST":
			mask |= 1 << 1
		case "PUT":
			mask |= 1 << 2
		case "DELETE":
			mask |= 1 << 3
		case "PATCH":
			mask |= 1 << 4
		case "HEAD":
			mask |= 1 << 5
		case "OPTIONS":
			mask |= 1 << 6
		case "CONNECT":
			mask |= 1 << 7
		case "TRACE":
			mask |= 1 << 8
		}
	}
	return mask
}

// allowHeaderFromMap joins the method keys in a predictable, stable order.
// allowHeaderFromMap joins the method keys in a predictable, stable order.
func allowHeaderFromMap(m map[string]*routing.RouteOptions) string {
	if len(m) == 0 {
		return ""
	}
	// Emit methods in a conventional order
	order := []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "CONNECT", "TRACE"}
	out := make([]string, 0, len(m))
	for _, meth := range order {
		if _, ok := m[meth]; ok {
			out = append(out, meth)
		}
	}
	// If any non-standard methods were registered, include them deterministically
	for k := range m {
		known := false
		for _, s := range order {
			if s == k {
				known = true
				break
			}
		}
		if !known {
			out = append(out, k)
		}
	}
	return strings.Join(out, ", ")
}

// findNode traverses the route tree by the given path and returns the terminal
// node if the path matches a node in the tree. If no matching node is found,
// it returns nil.
// findNode traverses the route tree by the given path and returns the terminal
// node if the path matches a node in the tree. If no matching node is found,
// it returns nil.
func (r *RouteRegistry) findNode(path string) *routing.RouteNode {
	start, end := trimPathIndices(path)
	n := r.root
	s := start
	for s < end {
		// Early short-circuit using precomputed flags
		if n.HasOnlyCatchAll {
			return n.CatchAll
		}
		if n.HasOnlyWildcardTerminal {
			return n.Wildcard
		}
		seg, nextIndex := scanSegment(path, s, end)
		next, edge := chooseNextEdge(n, seg)
		switch edge {
		case "child", "param", "wildcard":
			n = next
		case "catchall":
			return next
		default:
			return nil
		}
		s = nextIndex
	}
	return n
}

// FindNode performs a non-allocating traversal for the given path and returns
// the terminal RouteNode when the path matches a node in the tree. This
// function does not populate any params map and therefore avoids allocations
// when callers only need to inspect the matched node (for example to read
// Allow header metadata on method mismatch).
func (r *RouteRegistry) FindNode(path string) *routing.RouteNode {
	return r.findNode(path)
}

// FindNodeIntoSlice traverses the routing tree for the given path and fills any
// path parameters into dst (resetting it first). Returns the terminal RouteNode
// or nil if no match.
func (r *RouteRegistry) FindNodeIntoSlice(path string, dst *routing.Params) *routing.RouteNode {
	node, ok := r.matchNodeIntoSlice(path, dst)
	if !ok || node == nil {
		return nil
	}
	return node
}
