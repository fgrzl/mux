package routing

import (
	"github.com/fgrzl/mux/pkg/binder"
	"github.com/google/uuid"
)

// Param returns the raw string value of a route parameter.
func (c *DefaultRouteContext) Param(name string) (string, bool) {
	val, ok := c.params[name]
	return val, ok
}

// ParamUUID parses a UUID from a route parameter.
func (c *DefaultRouteContext) ParamUUID(name string) (uuid.UUID, bool) {
	val, ok := c.Param(name)
	if !ok {
		return uuid.Nil, false
	}
	return binder.ParseUUIDVal(val)
}

// ParamInt parses an int from a route parameter.
func (c *DefaultRouteContext) ParamInt(name string) (int, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	return binder.ParseIntVal(val)
}

// ParamInt16 parses an int16 from a route parameter.
func (c *DefaultRouteContext) ParamInt16(name string) (int16, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt16Val(val)
}

// ParamInt32 parses an int32 from a route parameter.
func (c *DefaultRouteContext) ParamInt32(name string) (int32, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt32Val(val)
}

// ParamInt64 parses an int64 from a route parameter.
func (c *DefaultRouteContext) ParamInt64(name string) (int64, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt64Val(val)
}
