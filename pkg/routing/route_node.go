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
	// HasOnlyCatchAll is a precomputed fast-path used by the registry traversal
	// to detect nodes whose only possible next step is the CatchAll edge. When true,
	// traversal can short-circuit immediately without scanning remaining segments.
	HasOnlyCatchAll bool
	// HasOnlyWildcardTerminal indicates that from this node the only possible next
	// step is the Wildcard edge and that edge leads to a terminal node for the
	// pattern (e.g., "/files/*"). Traversal can short-circuit by returning the
	// wildcard node directly without scanning remaining segments.
	HasOnlyWildcardTerminal bool
	// HasParams indicates that the pattern leading to this node included one or more
	// path parameters (e.g., {id}). Callers can use this to decide whether to build
	// a params map for the happy path without performing another traversal.
	HasParams bool
	// ParamCount stores the number of path parameters in the pattern leading to this node.
	// This enables pre-allocation of parameter maps with the correct capacity to avoid
	// map growth allocations during request handling.
	ParamCount int
}
