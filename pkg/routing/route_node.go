package routing

type RouteNode struct {
	Children     map[string]*RouteNode
	ParamChild   *RouteNode
	Wildcard     *RouteNode // for *
	CatchAll     *RouteNode // for **
	ParamName    string
	RouteOptions map[string]*RouteOptions // keyed by method
	// Cached method metadata for performance (populated by registry on register)
	MethodsMask uint32 // bitmask of allowed methods for this node
	AllowHeader string // pre-joined Allow header value (e.g., "GET, POST")
}
