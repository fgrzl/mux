package mux

import (
	"strconv"

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
	n, err := strconv.Atoi(val)
	return n, err == nil
}

// HeaderUUID parses a header value into a UUID.
func (c *DefaultRouteContext) HeaderUUID(name string) (uuid.UUID, bool) {
	val, ok := c.Header(name)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(val)
	return id, err == nil
}

// HeaderBool parses a header value into a bool.
func (c *DefaultRouteContext) HeaderBool(name string) (bool, bool) {
	val, ok := c.Header(name)
	if !ok {
		return false, false
	}
	b, err := strconv.ParseBool(val)
	return b, err == nil
}

// HeaderFloat64 parses a header value into a float64.
func (c *DefaultRouteContext) HeaderFloat64(name string) (float64, bool) {
	val, ok := c.Header(name)
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(val, 64)
	return f, err == nil
}
