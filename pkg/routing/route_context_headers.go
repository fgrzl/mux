package routing

import (
	"github.com/fgrzl/mux/pkg/binder"
	"github.com/google/uuid"
)

// Header returns a raw header value.
func (c *DefaultRouteContext) Header(name string) (string, bool) {
	val := c.Request().Header.Get(name)
	return val, val != ""
}

// HeaderInt parses a header value into an int.
func (c *DefaultRouteContext) HeaderInt(name string) (int, bool) {
	val, ok := c.Header(name)
	if !ok {
		return 0, false
	}
	return binder.ParseIntVal(val)
}

// HeaderUUID parses a header value into a UUID.
func (c *DefaultRouteContext) HeaderUUID(name string) (uuid.UUID, bool) {
	val, ok := c.Header(name)
	if !ok {
		return uuid.Nil, false
	}
	return binder.ParseUUIDVal(val)
}

// HeaderBool parses a header value into a bool.
func (c *DefaultRouteContext) HeaderBool(name string) (bool, bool) {
	val, ok := c.Header(name)
	if !ok {
		return false, false
	}
	return binder.ParseBoolVal(val)
}

// HeaderFloat64 parses a header value into a float64.
func (c *DefaultRouteContext) HeaderFloat64(name string) (float64, bool) {
	val, ok := c.Header(name)
	if !ok {
		return 0, false
	}
	return binder.ParseFloat64Val(val)
}
