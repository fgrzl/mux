package routing

type RouteNode struct {
	Children     map[string]*RouteNode
	ParamChild   *RouteNode
	Wildcard     *RouteNode // for *
	CatchAll     *RouteNode // for **
	ParamName    string
	RouteOptions map[string]*RouteOptions // keyed by method
}
