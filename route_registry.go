package mux

import (
	"strings"
)

type routeNode struct {
	Children     map[string]*routeNode
	ParamChild   *routeNode
	Wildcard     *routeNode // for *
	CatchAll     *routeNode // for **
	ParamName    string
	RouteOptions map[string]*RouteOptions // keyed by method
}

type RouteRegistry struct {
	root *routeNode
}

func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		root: &routeNode{Children: make(map[string]*routeNode)},
	}
}

func (r *RouteRegistry) Register(pattern string, method string, options *RouteOptions) {
	segments := strings.Split(strings.Trim(pattern, "/"), "/")
	node := r.root

	for _, seg := range segments {
		if seg == "**" {
			if node.CatchAll == nil {
				node.CatchAll = &routeNode{Children: make(map[string]*routeNode)}
			}
			node = node.CatchAll
			break
		} else if seg == "*" {
			if node.Wildcard == nil {
				node.Wildcard = &routeNode{Children: make(map[string]*routeNode)}
			}
			node = node.Wildcard
		} else if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			if node.ParamChild == nil {
				node.ParamChild = &routeNode{Children: make(map[string]*routeNode)}
				node.ParamChild.ParamName = seg[1 : len(seg)-1]
			}
			node = node.ParamChild
		} else {
			if node.Children[seg] == nil {
				node.Children[seg] = &routeNode{Children: make(map[string]*routeNode)}
			}
			node = node.Children[seg]
		}
	}

	if node.RouteOptions == nil {
		node.RouteOptions = make(map[string]*RouteOptions)
	}

	node.RouteOptions[method] = options
}

func (r *RouteRegistry) Load(path string, method string) (*RouteOptions, map[string]string, bool) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	params := make(map[string]string)

	var search func(node *routeNode, segs []string) (*RouteOptions, bool)
	search = func(node *routeNode, segs []string) (*RouteOptions, bool) {
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
