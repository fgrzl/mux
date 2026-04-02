package routing

import (
	"github.com/fgrzl/mux/pkg/binder"
	"github.com/google/uuid"
)

// QueryValue returns the first value for the given query key.
func (c *DefaultRouteContext) QueryValue(name string) (string, bool) {
	vals, ok := c.Request().URL.Query()[name]
	if ok && len(vals) > 0 {
		return vals[0], true
	}
	return "", false
}

// QueryValues returns all values for the given query key.
func (c *DefaultRouteContext) QueryValues(name string) ([]string, bool) {
	vals, ok := c.Request().URL.Query()[name]
	return vals, ok
}

// QueryUUID parses a UUID from a query parameter.
func (c *DefaultRouteContext) QueryUUID(name string) (uuid.UUID, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return uuid.Nil, false
	}
	return binder.ParseUUIDVal(val)
}

// QueryUUIDs parses a list of UUIDs from query parameters.
func (c *DefaultRouteContext) QueryUUIDs(name string) ([]uuid.UUID, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseUUIDSlice(vals)
}

// QueryInt parses an int from a query parameter.
func (c *DefaultRouteContext) QueryInt(name string) (int, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseIntVal(val)
}

// QueryInts parses a list of ints from query parameters.
func (c *DefaultRouteContext) QueryInts(name string) ([]int, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseIntSlice(vals)
}

// QueryInt16 parses an int16 from a query parameter.
func (c *DefaultRouteContext) QueryInt16(name string) (int16, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt16Val(val)
}

// QueryInt16s parses a list of int16s from query parameters.
func (c *DefaultRouteContext) QueryInt16s(name string) ([]int16, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseInt16Slice(vals)
}

// QueryInt32 parses an int32 from a query parameter.
func (c *DefaultRouteContext) QueryInt32(name string) (int32, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt32Val(val)
}

// QueryInt32s parses a list of int32s from query parameters.
func (c *DefaultRouteContext) QueryInt32s(name string) ([]int32, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseInt32Slice(vals)
}

// QueryInt64 parses an int64 from a query parameter.
func (c *DefaultRouteContext) QueryInt64(name string) (int64, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt64Val(val)
}

// QueryInt64s parses a list of int64s from query parameters.
func (c *DefaultRouteContext) QueryInt64s(name string) ([]int64, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseInt64Slice(vals)
}

// QueryBool parses a bool from a query parameter.
func (c *DefaultRouteContext) QueryBool(name string) (bool, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return false, false
	}
	return binder.ParseBoolVal(val)
}

// QueryBools parses a list of bools from query parameters.
func (c *DefaultRouteContext) QueryBools(name string) ([]bool, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseBoolSlice(vals)
}

// QueryFloat32 parses a float32 from a query parameter.
func (c *DefaultRouteContext) QueryFloat32(name string) (float32, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseFloat32Val(val)
}

// QueryFloat32s parses a list of float32s from query parameters.
func (c *DefaultRouteContext) QueryFloat32s(name string) ([]float32, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseFloat32Slice(vals)
}

// QueryFloat64 parses a float64 from a query parameter.
func (c *DefaultRouteContext) QueryFloat64(name string) (float64, bool) {
	val, ok := c.QueryValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseFloat64Val(val)
}

// QueryFloat64s parses a list of float64s from query parameters.
func (c *DefaultRouteContext) QueryFloat64s(name string) ([]float64, bool) {
	vals, ok := c.QueryValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseFloat64Slice(vals)
}

// GetRedirectURL returns the first matching redirect-related query value or the fallback.
func (c *DefaultRouteContext) GetRedirectURL(defaultRedirect string) string {
	candidates := []string{
		"redirect_uri", // OAuth2 convention
		"redirect_url",
		"return_url",
		"returnUrl",
		"return_to",
		"redirect",
		"return",
	}
	for _, key := range candidates {
		if url, ok := c.QueryValue(key); ok {
			return url
		}
	}
	return defaultRedirect
}
