package openapi

// RouteData is a small value type used by the generator. Route collection
// is performed by the mux adapter so the openapi package does not depend on
// router internals.
type RouteData struct {
	Path   string
	Method string
	// Options should carry the OpenAPI Operation information; we store a pointer
	// to openapi.Operation so the generator can use it without depending on
	// the router's RouteOptions type.
	Options *Operation
}
