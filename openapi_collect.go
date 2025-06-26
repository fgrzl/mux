package mux

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

// routeData represents a route for OpenAPI generation.
type routeData struct {
	Path    string
	Method  string
	Options *RouteOptions
}

// collectRoutes collects all routes from a routeNode tree.
func collectRoutes(node *routeNode) ([]routeData, error) {
	if node == nil {
		return nil, fmt.Errorf("route node is nil")
	}
	var routes []routeData
	var walk func(string, *routeNode) error
	walk = func(prefix string, n *routeNode) error {
		for method, opt := range n.RouteOptions {
			if method == "" {
				return fmt.Errorf("empty method in route options at path %q", prefix)
			}
			routes = append(routes, routeData{
				Path:    cleanPath(prefix),
				Method:  strings.ToUpper(method),
				Options: opt,
			})
		}
		for seg, child := range n.Children {
			if seg == "" {
				return fmt.Errorf("empty segment in children at path %q", prefix)
			}
			if err := walk(path.Join(prefix, seg), child); err != nil {
				return err
			}
		}
		if n.ParamChild != nil {
			if n.ParamChild.ParamName == "" {
				return fmt.Errorf("empty param name at path %q", prefix)
			}
			if err := walk(path.Join(prefix, "{"+n.ParamChild.ParamName+"}"), n.ParamChild); err != nil {
				return err
			}
		}
		if n.Wildcard != nil {
			if err := walk(path.Join(prefix, "*"), n.Wildcard); err != nil {
				return err
			}
		}
		if n.CatchAll != nil {
			if err := walk(path.Join(prefix, "**"), n.CatchAll); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk("", node); err != nil {
		return nil, err
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	return routes, nil
}

// cleanPath ensures consistent path formatting.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	p = path.Clean("/" + strings.Trim(p, "/"))
	if p == "." {
		return "/"
	}
	return p
}
