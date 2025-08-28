package registry

import (
	"strings"

	"github.com/fgrzl/mux/internal/routing"
)

type RouteRegistry struct {
	root *routing.RouteNode
}

func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		root: &routing.RouteNode{Children: make(map[string]*routing.RouteNode)},
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
}

func (r *RouteRegistry) Load(path string, method string) (*routing.RouteOptions, map[string]string, bool) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	params := make(map[string]string)

	var search func(node *routing.RouteNode, segs []string) (*routing.RouteOptions, bool)
	search = func(node *routing.RouteNode, segs []string) (*routing.RouteOptions, bool) {
		if len(segs) == 0 {
			opt, ok := node.RouteOptions[method]
			return opt, ok
		}

		seg := segs[0]
		rest := segs[1:]

		// Exact match
		if child, ok := node.Children[seg]; ok {
			if opt, found := search(child, rest); found {
				return opt, true
			}
		}

		// Param match
		if node.ParamChild != nil {
			params[node.ParamChild.ParamName] = seg
			if opt, found := search(node.ParamChild, rest); found {
				return opt, true
			}
			delete(params, node.ParamChild.ParamName)
		}

		// Wildcard match
		if node.Wildcard != nil {
			if opt, found := search(node.Wildcard, rest); found {
				return opt, true
			}
		}

		// Catch-all match (**) — always matches
		if node.CatchAll != nil {
			opt, ok := node.CatchAll.RouteOptions[method]
			return opt, ok
		}

		return nil, false
	}

	opt, ok := search(r.root, segments)
	return opt, params, ok
}
