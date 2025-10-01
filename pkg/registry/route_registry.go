package registry

import (
	"strings"

	"github.com/fgrzl/mux/pkg/routing"
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

// splitSegments splits a trimmed path into non-empty segments, ignoring
// consecutive slashes. This centralizes the logic used when registering
// patterns to keep Register shorter and easier to read.
func splitSegments(trimmed string) []string {
	rawSegs := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(rawSegs))
	for _, s := range rawSegs {
		if s == "" {
			continue
		}
		segments = append(segments, s)
	}
	return segments
}

// trimPathBounds returns start and end indices that skip leading and trailing
// slashes for the provided path. This is a small helper to keep traversal
// code concise and consistent.
func trimPathBounds(path string) (start, end int) {
	start = 0
	end = len(path)
	for start < end && path[start] == '/' {
		start++
	}
	for end > start && path[end-1] == '/' {
		end--
	}
	return
}

// scanSegment returns the next path segment starting at s up to end and the
// index after the segment (nextIndex). It does not allocate except for the
// returned slice header.
func scanSegment(path string, s, end int) (seg string, nextIndex int) {
	j := s
	for j < end && path[j] != '/' {
		j++
	}
	return path[s:j], j + 1
}

// Root returns the trie root node for this registry. Callers may use
// the returned node for inspection or debugging; modifications to the
// node are not recommended outside of the registry methods.
// Root returns the trie root node for this registry. Callers may use
// the returned node for inspection or debugging; modifications to the
// node are not recommended outside of the registry methods.
func (r *RouteRegistry) Root() *routing.RouteNode {
	return r.root
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
	// If the trimmed pattern is empty, it refers to the root node
	if trimmed == "" {
		node := r.root
		if node.RouteOptions == nil {
			node.RouteOptions = make(map[string]*routing.RouteOptions)
		}
		node.RouteOptions[method] = options
		// Update metadata on the root node
		node.MethodsMask = methodsMaskFromMap(node.RouteOptions)
		node.AllowHeader = allowHeaderFromMap(node.RouteOptions)
		if !strings.ContainsAny(pattern, "{*}") {
			m := r.exactRoutes[pattern]
			if m == nil {
				m = make(map[string]*routing.RouteOptions)
				r.exactRoutes[pattern] = m
			}
			m[method] = options
		}
		return
	}
	// Split and ignore any empty segments created by consecutive slashes
	rawSegs := strings.Split(trimmed, "/")
	segments := make([]string, 0, len(rawSegs))
	for _, s := range rawSegs {
		if s == "" {
			// Skip empty segments caused by consecutive slashes in pattern.
			// These should be ignored to avoid creating unintended route nodes.
			continue
		}
		segments = append(segments, s)
	}
	node := r.root
	hasParams := false

	for _, seg := range segments {
		if seg == "**" {
			if node.CatchAll == nil {
				node.CatchAll = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
			}
			// update parent fast-path flags
			r.refreshFastPathFlags(node)
			node = node.CatchAll
			break
		} else if seg == "*" {
			if node.Wildcard == nil {
				node.Wildcard = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
			}
			r.refreshFastPathFlags(node)
			node = node.Wildcard
		} else if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			hasParams = true
			if node.ParamChild == nil {
				node.ParamChild = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
				node.ParamChild.ParamName = seg[1 : len(seg)-1]
			}
			r.refreshFastPathFlags(node)
			node = node.ParamChild
		} else {
			if node.Children[seg] == nil {
				node.Children[seg] = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
			}
			r.refreshFastPathFlags(node)
			node = node.Children[seg]
		}
	}

	if node.RouteOptions == nil {
		node.RouteOptions = make(map[string]*routing.RouteOptions)
	}

	node.RouteOptions[method] = options
	if hasParams {
		node.HasParams = true
	}
	// Update cached method metadata (mask and Allow header)
	node.MethodsMask = methodsMaskFromMap(node.RouteOptions)
	node.AllowHeader = allowHeaderFromMap(node.RouteOptions)

	// If the pattern contains no parameter or wildcard tokens, keep a copy
	// in the exactRoutes map as a fast path for Load.
	if !strings.ContainsAny(pattern, "{*}") {
		m := r.exactRoutes[pattern]
		if m == nil {
			m = make(map[string]*routing.RouteOptions)
			r.exactRoutes[pattern] = m
		}
		m[method] = options
	}
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

// Load performs a route lookup for the supplied path and method. It returns
// the matched RouteOptions, an extracted params map (or nil for no params),
// and ok=true when a matching route exists for the path and method. For
// static registered patterns an internal fast-path is used to avoid trie
// traversal and allocations.
// Load performs a route lookup for the supplied path and method. It returns
// the matched RouteOptions, an extracted params map (or nil for no params),
// and ok=true when a matching route exists for the path and method. For
// static registered patterns an internal fast-path is used to avoid trie
// traversal and allocations.
func (r *RouteRegistry) Load(path string, method string) (*routing.RouteOptions, map[string]string, bool) {
	// Fast path: exact registered static route
	if m, ok := r.exactRoutes[path]; ok {
		if opt, ok2 := m[method]; ok2 {
			return opt, nil, true
		}
	}

	opt, params, ok := r.walk(path, func(n *routing.RouteNode, method string) (*routing.RouteOptions, bool) {
		if n.RouteOptions == nil {
			return nil, false
		}
		o, ok := n.RouteOptions[method]
		return o, ok
	}, method)
	return opt, params, ok
}

// LoadDetailedInto performs a route lookup, fills any path params into dst (clearing it first),
// and returns the matched RouteOptions along with LoadDetails describing the match.
// If the path matches but the method is not allowed, RouteOptions is nil and details.Allow
// contains the precomputed Allow header value.
// LoadDetailedInto performs a route lookup, fills any path params into dst (clearing it first),
// and returns the matched RouteOptions along with LoadDetails describing the match.
// If the path matches but the method is not allowed, RouteOptions is nil and details.Allow
// contains the precomputed Allow header value.
func (r *RouteRegistry) LoadDetailedInto(path string, method string, dst map[string]string) (*routing.RouteOptions, LoadDetails) {
	// Clear params map up-front (if provided)
	for k := range dst {
		delete(dst, k)
	}
	// Exact static fast-path
	if m, ok := r.exactRoutes[path]; ok {
		details := LoadDetails{Found: true}
		if opt, ok2 := m[method]; ok2 {
			details.MethodOK = true
			return opt, details
		}
		details.Allow = allowHeaderFromMap(m)
		return nil, details
	}
	// Match a node and collect params greedily with precedence: static > param > wildcard > catch-all
	node, _ := r.matchNodeInto(path, dst)
	if node == nil || len(node.RouteOptions) == 0 {
		return nil, LoadDetails{Found: false}
	}
	if opt, ok := node.RouteOptions[method]; ok {
		return opt, LoadDetails{Found: true, MethodOK: true}
	}
	return nil, LoadDetails{Found: true, MethodOK: false, Allow: node.AllowHeader}
}

// matchNodeInto traverses the routing tree by path and returns the terminal node if matched.
// It fills any path params into dst (if non-nil), clearing dst is the caller's responsibility.
// matchNodeInto traverses the routing tree by path and returns the terminal node if matched.
// It fills any path params into dst (if non-nil), clearing dst is the caller's responsibility.
func (r *RouteRegistry) matchNodeInto(path string, dst map[string]string) (*routing.RouteNode, bool) {
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
		j := s
		for j < end && path[j] != '/' {
			j++
		}
		seg := path[s:j]
		nextIndex := j + 1
		if child, ok := n.Children[seg]; ok {
			n = child
		} else if n.ParamChild != nil {
			if dst != nil {
				dst[n.ParamChild.ParamName] = seg
			}
			n = n.ParamChild
		} else if n.Wildcard != nil {
			n = n.Wildcard
		} else if n.CatchAll != nil {
			n = n.CatchAll
			break
		} else {
			return nil, false
		}
		s = nextIndex
	}
	return n, true
}

// LoadInto is like Load but writes any extracted route parameters into the provided
// map to avoid per-call map allocations. The provided map will be cleared first.
// It returns the matched RouteOptions and ok=true when a matching route exists.
// LoadInto is like Load but writes any extracted route parameters into the provided
// map to avoid per-call map allocations. The provided map will be cleared first.
// It returns the matched RouteOptions and ok=true when a matching route exists.
func (r *RouteRegistry) LoadInto(path string, method string, dst map[string]string) (*routing.RouteOptions, bool) {
	// Fast path: exact registered static route
	if m, ok := r.exactRoutes[path]; ok {
		if opt, ok2 := m[method]; ok2 {
			// Ensure dst is cleared
			for k := range dst {
				delete(dst, k)
			}
			return opt, true
		}
	}
	opt, ok := r.walkInto(path, method, dst, func(n *routing.RouteNode, method string) (*routing.RouteOptions, bool) {
		if n.RouteOptions == nil {
			return nil, false
		}
		o, ok := n.RouteOptions[method]
		return o, ok
	})
	return opt, ok
}

// walkInto traverses the tree and fills params into dst (which may be nil) without allocating
// a new map. dst is cleared on entry if non-nil.
// walkInto traverses the tree and fills params into dst (which may be nil) without allocating
// a new map. dst is cleared on entry if non-nil.
func (r *RouteRegistry) walkInto(path string, method string, dst map[string]string, atEnd func(*routing.RouteNode, string) (*routing.RouteOptions, bool)) (*routing.RouteOptions, bool) {
	// We'll scan the path in-place between indices [start,end)
	start := 0
	end := len(path)
	for start < end && path[start] == '/' {
		start++
	}
	for end > start && path[end-1] == '/' {
		end--
	}

	for k := range dst {
		delete(dst, k)
	}

	type frame struct {
		node     *routing.RouteNode
		s        int
		stage    int
		paramKey string
	}
	var stackBuf [16]frame
	stack := stackBuf[:0]
	stack = append(stack, frame{node: r.root, s: start, stage: 0})
	for len(stack) > 0 {
		f := &stack[len(stack)-1]
		node := f.node
		s := f.s
		if s >= end {
			if opt, ok := atEnd(node, method); ok {
				return opt, true
			}
			if f.paramKey != "" {
				delete(dst, f.paramKey)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		// Early short-circuit using precomputed flag
		if node.HasOnlyCatchAll {
			if opt, ok := atEnd(node.CatchAll, method); ok {
				return opt, true
			}
			if f.paramKey != "" {
				delete(dst, f.paramKey)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		// Fast-path for wildcard terminal node to avoid scanning remaining
		// segments when the wildcard edge is the only possible next step and
		// it represents a terminal route for some methods.
		if node.HasOnlyWildcardTerminal {
			if opt, ok := atEnd(node.Wildcard, method); ok {
				return opt, true
			}
			if f.paramKey != "" {
				delete(dst, f.paramKey)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		j := s
		for j < end && path[j] != '/' {
			j++
		}
		seg := path[s:j]
		nextIndex := j + 1
		switch f.stage {
		case 0:
			f.stage = 1
			if child, ok := node.Children[seg]; ok {
				stack = append(stack, frame{node: child, s: nextIndex, stage: 0})
				continue
			}
			fallthrough
		case 1:
			f.stage = 2
			if node.ParamChild != nil {
				dst[node.ParamChild.ParamName] = seg
				stack = append(stack, frame{node: node.ParamChild, s: nextIndex, stage: 0, paramKey: node.ParamChild.ParamName})
				continue
			}
			fallthrough
		case 2:
			f.stage = 3
			if node.Wildcard != nil {
				// Wildcard consumes this segment; if terminal-only fast-path was not true,
				// we still need to continue traversal
				stack = append(stack, frame{node: node.Wildcard, s: nextIndex, stage: 0})
				continue
			}
			fallthrough
		case 3:
			if node.CatchAll != nil {
				if opt, ok := atEnd(node.CatchAll, method); ok {
					return opt, true
				}
			}
			if f.paramKey != "" {
				delete(dst, f.paramKey)
			}
			stack = stack[:len(stack)-1]
		}
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
	// We'll scan the path in-place between indices [start,end)
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
		// Early short-circuit using precomputed flags
		if n.HasOnlyCatchAll {
			return n.CatchAll
		}
		if n.HasOnlyWildcardTerminal {
			return n.Wildcard
		}
		j := s
		for j < end && path[j] != '/' {
			j++
		}
		seg := path[s:j]
		nextIndex := j + 1
		if child, ok := n.Children[seg]; ok {
			n = child
		} else if n.ParamChild != nil {
			n = n.ParamChild
		} else if n.Wildcard != nil {
			n = n.Wildcard
		} else if n.CatchAll != nil {
			n = n.CatchAll
			break
		} else {
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

// FindNodeInto traverses the routing tree for the given path and fills any
// path parameters into dst (clearing it first). It returns the terminal
// RouteNode when the path matches a registered node (even if that node
// doesn't have RouteOptions). If no node matches, it returns nil.
//
// This is an exported convenience wrapper around the internal matchNodeInto
// implementation to allow callers to perform a single traversal and then
// examine the node's RouteOptions without repeating work.
func (r *RouteRegistry) FindNodeInto(path string, dst map[string]string) *routing.RouteNode {
	node, ok := r.matchNodeInto(path, dst)
	if !ok || node == nil {
		return nil
	}
	return node
}

// walk performs a pluggable depth-first search (DFS) traversal of the route tree.
// It traverses the tree according to the given path and invokes the provided
// atEnd function when a terminal node is reached. This allows custom logic to be
// applied at the end of traversal, such as matching a route or collecting
// parameters.
//
// Parameters:
//   - path: The URL path to traverse, which may include parameters or wildcards.
//   - atEnd: A function called at each terminal node, receiving the current node
//     and the HTTP method. It should return a pointer to RouteOptions and a
//     boolean indicating whether a match was found.
//   - method: The HTTP method to match (e.g., "GET", "POST").
//
// Returns:
//   - *routing.RouteOptions: The matched route options if found, or nil.
//   - map[string]string: A map of extracted route parameters, if any (may be nil).
//   - bool: True if a matching route was found, false otherwise.
func (r *RouteRegistry) walk(path string, atEnd func(*routing.RouteNode, string) (*routing.RouteOptions, bool), method string) (*routing.RouteOptions, map[string]string, bool) {
	// We'll scan the path in-place between indices [start,end)
	// Skip leading and trailing slashes without allocating a trimmed string
	start := 0
	end := len(path)
	for start < end && path[start] == '/' {
		start++
	}
	for end > start && path[end-1] == '/' {
		end--
	}

	var params map[string]string

	type frame struct {
		node     *routing.RouteNode
		s        int
		stage    int
		paramKey string
	}
	// Use a small, stack-allocated buffer to avoid heap allocation for typical
	// path depths. If more frames are needed, append will grow capacity on heap.
	var stackBuf [16]frame
	stack := stackBuf[:0]
	stack = append(stack, frame{node: r.root, s: start, stage: 0})
	for len(stack) > 0 {
		f := &stack[len(stack)-1]
		node := f.node
		s := f.s
		if s >= end {
			if opt, ok := atEnd(node, method); ok {
				return opt, params, true
			}
			if f.paramKey != "" && params != nil {
				delete(params, f.paramKey)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		// Early short-circuit using precomputed flag
		if node.HasOnlyCatchAll {
			if opt, ok := atEnd(node.CatchAll, method); ok {
				return opt, params, true
			}
			if f.paramKey != "" && params != nil {
				delete(params, f.paramKey)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		// Fast-path for wildcard terminal: if the only possible next step is the
		// wildcard edge and that edge is terminal for some methods, we can
		// short-circuit and invoke atEnd on the wildcard node directly which
		// avoids scanning remaining segments and avoids param handling.
		if node.HasOnlyWildcardTerminal {
			if opt, ok := atEnd(node.Wildcard, method); ok {
				return opt, params, true
			}
			if f.paramKey != "" && params != nil {
				delete(params, f.paramKey)
			}
			stack = stack[:len(stack)-1]
			continue
		}
		j := s
		for j < end && path[j] != '/' {
			j++
		}
		seg := path[s:j]
		nextIndex := j + 1
		switch f.stage {
		case 0:
			f.stage = 1
			if child, ok := node.Children[seg]; ok {
				stack = append(stack, frame{node: child, s: nextIndex, stage: 0})
				continue
			}
			fallthrough
		case 1:
			f.stage = 2
			if node.ParamChild != nil {
				if params == nil {
					// Pre-size for the common case of a small number of params
					params = make(map[string]string, 1)
				}
				params[node.ParamChild.ParamName] = seg
				stack = append(stack, frame{node: node.ParamChild, s: nextIndex, stage: 0, paramKey: node.ParamChild.ParamName})
				continue
			}
			fallthrough
		case 2:
			f.stage = 3
			if node.Wildcard != nil {
				stack = append(stack, frame{node: node.Wildcard, s: nextIndex, stage: 0})
				continue
			}
			fallthrough
		case 3:
			if node.CatchAll != nil {
				if opt, ok := atEnd(node.CatchAll, method); ok {
					return opt, params, true
				}
			}
			if f.paramKey != "" && params != nil {
				delete(params, f.paramKey)
			}
			stack = stack[:len(stack)-1]
		}
	}
	return nil, params, false
}
