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
	// Walk to a node and collect methods at match
	_, _, matched := r.walk(path, func(n *routing.RouteNode, _ string) (*routing.RouteOptions, bool) {
		if len(n.RouteOptions) == 0 {
			return nil, false
		}
		return &routing.RouteOptions{}, true // sentinel
	}, "")
	if !matched {
		return nil, false
	}
	// Retrieve methods at the matched node (opt is sentinel; node not returned)
	// Re-walk to get the node so we can enumerate methods
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

// findNode traverses the tree by path and returns the terminal node if matched.
func (r *RouteRegistry) findNode(path string) *routing.RouteNode {
	_, _, ok := r.walk(path, func(n *routing.RouteNode, _ string) (*routing.RouteOptions, bool) {
		return &routing.RouteOptions{}, true // we only care about terminal node existence
	}, "")
	if !ok {
		return nil
	}
	// Sadly walk does not currently return the node; for simplicity, re-implement a minimal traversal
	// We'll keep this helper simple: return nil if not found; otherwise the exact terminal node

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
	stack := make([]frame, 0, 16)
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
					params = make(map[string]string)
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
