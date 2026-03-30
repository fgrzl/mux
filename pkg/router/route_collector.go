package router

import (
	"fmt"
	"path"
	"sort"
	"strings"

	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/routing"
)

// collectRoutesFromNode traverses the router's internal routing.RouteNode tree and
// produces a slice of openapi.RouteData.
func collectRoutesFromNode(node *routing.RouteNode) ([]openapi.RouteData, error) {
	if node == nil {
		return nil, fmt.Errorf("route node is nil")
	}
	var routes []openapi.RouteData
	var walk func(string, *routing.RouteNode) error
	walk = func(prefix string, n *routing.RouteNode) error {
		for method, opt := range n.RouteOptions {
			if method == "" {
				return fmt.Errorf("empty method in route options at path %q", prefix)
			}
			routes = append(routes, openapi.RouteData{
				Path:    cleanPath(prefix),
				Method:  strings.ToUpper(method),
				Options: openapi.CloneOperation(&opt.Operation),
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

// cleanPath ensures consistent path formatting (copied from the former openapi helper).
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
