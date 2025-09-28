package registry

import (
	"strings"

	"github.com/fgrzl/mux/pkg/routing"
)

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

func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		root:        &routing.RouteNode{Children: make(map[string]*routing.RouteNode)},
		exactRoutes: make(map[string]map[string]*routing.RouteOptions),
	}
}

func (r *RouteRegistry) Root() *routing.RouteNode {
	return r.root
}

func (r *RouteRegistry) Register(pattern string, method string, options *routing.RouteOptions) {
	segments := strings.Split(strings.Trim(pattern, "/"), "/")
	node := r.root

	for _, seg := range segments {
		if seg == "**" {
			if node.CatchAll == nil {
				node.CatchAll = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
			}
			node = node.CatchAll
			break
		} else if seg == "*" {
			if node.Wildcard == nil {
				node.Wildcard = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
			}
			node = node.Wildcard
		} else if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			if node.ParamChild == nil {
				node.ParamChild = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
				node.ParamChild.ParamName = seg[1 : len(seg)-1]
			}
			node = node.ParamChild
		} else {
			if node.Children[seg] == nil {
				node.Children[seg] = &routing.RouteNode{Children: make(map[string]*routing.RouteNode)}
			}
			node = node.Children[seg]
		}
	}

	if node.RouteOptions == nil {
		node.RouteOptions = make(map[string]*routing.RouteOptions)
	}

	node.RouteOptions[method] = options
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
func (r *RouteRegistry) LoadDetailedInto(path string, method string, dst map[string]string) (*routing.RouteOptions, LoadDetails) {
	// Clear params map up-front (if provided)
	if dst != nil {
		for k := range dst {
			delete(dst, k)
		}
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
			s = end
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
func (r *RouteRegistry) LoadInto(path string, method string, dst map[string]string) (*routing.RouteOptions, bool) {
	// Fast path: exact registered static route
	if m, ok := r.exactRoutes[path]; ok {
		if opt, ok2 := m[method]; ok2 {
			// Ensure dst is cleared
			if dst != nil {
				for k := range dst {
					delete(dst, k)
				}
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

	if dst != nil {
		for k := range dst {
			delete(dst, k)
		}
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
			if f.paramKey != "" && dst != nil {
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
				if dst != nil {
					dst[node.ParamChild.ParamName] = seg
				}
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
					return opt, true
				}
			}
			if f.paramKey != "" && dst != nil {
				delete(dst, f.paramKey)
			}
			stack = stack[:len(stack)-1]
		}
	}
	return nil, false
}

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

// findNode traverses the tree by path and returns the terminal node if matched.
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
			s = end // consume rest
			break
		} else {
			return nil
		}
		s = nextIndex
	}
	return n
}

// walk performs a DFS traversal similar to Load but pluggable end-node predicate.
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
