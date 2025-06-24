package mux

import (
	"strconv"

	"github.com/google/uuid"
)

// Param returns the raw string value of a route parameter.
func (c *RouteContext) Param(name string) (string, bool) {
	val, ok := c.Params[name]
	return val, ok
}

// ParamUUID parses a UUID from a route parameter.
func (c *RouteContext) ParamUUID(name string) (uuid.UUID, bool) {
	val, ok := c.Param(name)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(val)
	return id, err == nil
}

// ParamInt parses an int from a route parameter.
func (c *RouteContext) ParamInt(name string) (int, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(val)
	return n, err == nil
}

// ParamInt16 parses an int16 from a route parameter.
func (c *RouteContext) ParamInt16(name string) (int16, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 16)
	return int16(n), err == nil
}

// ParamInt32 parses an int32 from a route parameter.
func (c *RouteContext) ParamInt32(name string) (int32, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 32)
	return int32(n), err == nil
}

// ParamInt64 parses an int64 from a route parameter.
func (c *RouteContext) ParamInt64(name string) (int64, bool) {
	val, ok := c.Param(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 64)
	return n, err == nil
}
