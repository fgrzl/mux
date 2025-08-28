package routing

import (
	"time"

	openapi "github.com/fgrzl/mux/pkg/openapi"
)

// HandlerFunc is the handler signature used by routing package.
// It accepts a RouteContext so implementations can be defined within
// the routing package without importing mux, avoiding cycles.
type HandlerFunc func(RouteContext)

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
}
