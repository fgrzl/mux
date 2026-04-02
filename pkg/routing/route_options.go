package routing

import (
	"strings"
	"time"

	openapi "github.com/fgrzl/mux/pkg/openapi"
)

// HandlerFunc is the handler signature used by routing package.
// It accepts a RouteContext so implementations can be defined within
// the routing package without importing mux, avoiding cycles.
type HandlerFunc = func(RouteContext)

// RouteOptions holds both runtime routing data (handler, auth etc.) and the
// OpenAPI Operation object that will be rendered into the spec.
// The Operation from the openapi package is embedded so code that expects
// fields like Parameters/Responses continues to work.
type RouteOptions struct {
	// ---- runtime routing metadata ----
	Method  string
	Pattern string
	Handler HandlerFunc

	// ---- runtime operations ----
	AllowAnonymous bool
	Roles          []string
	Scopes         []string
	Permissions    []string
	RateLimit      int
	RateInterval   time.Duration

	// ---- OpenAPI documentation ----
	openapi.Operation

	// ParamIndex is a runtime index of parameters for fast lookups.
	// Key format: strings.ToLower(in+":"+name)
	ParamIndex map[string]*openapi.ParameterObject
}

// BuildParamIndex constructs a lowercase parameter index keyed by "in:name".
// It returns nil if params is empty or nil.
func BuildParamIndex(params []*openapi.ParameterObject) map[string]*openapi.ParameterObject {
	if len(params) == 0 {
		return nil
	}
	idx := make(map[string]*openapi.ParameterObject, len(params))
	for _, p := range params {
		if p == nil {
			continue
		}
		key := strings.ToLower(p.In + ":" + p.Name)
		idx[key] = p
	}
	return idx
}
