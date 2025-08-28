package routing

import (
	"strings"

	"github.com/fgrzl/mux/pkg/binder"
	"github.com/google/uuid"
)

// ensureFormsParsed parses the request forms if they haven't been parsed yet.
// It supports both application/x-www-form-urlencoded and multipart/form-data.
func (c *DefaultRouteContext) ensureFormsParsed() error {
	if c.formsParsed {
		return nil
	}

	ct := c.Request().Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		// Parse multipart form with a 32MB max memory
		if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
			return err
		}
	} else if ct == "application/x-www-form-urlencoded" {
		if err := c.Request().ParseForm(); err != nil {
			return err
		}
	}

	c.formsParsed = true
	return nil
}

// FormValue returns the first value for the given form key.
func (c *DefaultRouteContext) FormValue(name string) (string, bool) {
	if err := c.ensureFormsParsed(); err != nil {
		return "", false
	}

	if c.Request().Form == nil {
		return "", false
	}

	vals, ok := c.Request().Form[name]
	if ok && len(vals) > 0 {
		return vals[0], true
	}
	return "", false
}

// FormValues returns all values for the given form key.
func (c *DefaultRouteContext) FormValues(name string) ([]string, bool) {
	if err := c.ensureFormsParsed(); err != nil {
		return nil, false
	}

	if c.Request().Form == nil {
		return nil, false
	}

	vals, ok := c.Request().Form[name]
	return vals, ok
}

// FormUUID parses a UUID from a form parameter.
func (c *DefaultRouteContext) FormUUID(name string) (uuid.UUID, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return uuid.Nil, false
	}
	return binder.ParseUUIDVal(val)
}

// FormUUIDs parses a list of UUIDs from form parameters.
func (c *DefaultRouteContext) FormUUIDs(name string) ([]uuid.UUID, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseUUIDSlice(vals)
}

// FormInt parses an int from a form parameter.
func (c *DefaultRouteContext) FormInt(name string) (int, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseIntVal(val)
}

// FormInts parses a list of ints from form parameters.
func (c *DefaultRouteContext) FormInts(name string) ([]int, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseIntSlice(vals)
}

// FormInt16 parses an int16 from a form parameter.
func (c *DefaultRouteContext) FormInt16(name string) (int16, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt16Val(val)
}

// FormInt16s parses a list of int16s from form parameters.
func (c *DefaultRouteContext) FormInt16s(name string) ([]int16, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseInt16Slice(vals)
}

// FormInt32 parses an int32 from a form parameter.
func (c *DefaultRouteContext) FormInt32(name string) (int32, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt32Val(val)
}

// FormInt32s parses a list of int32s from form parameters.
func (c *DefaultRouteContext) FormInt32s(name string) ([]int32, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseInt32Slice(vals)
}

// FormInt64 parses an int64 from a form parameter.
func (c *DefaultRouteContext) FormInt64(name string) (int64, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseInt64Val(val)
}

// FormInt64s parses a list of int64s from form parameters.
func (c *DefaultRouteContext) FormInt64s(name string) ([]int64, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseInt64Slice(vals)
}

// FormBool parses a bool from a form parameter.
func (c *DefaultRouteContext) FormBool(name string) (bool, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return false, false
	}
	return binder.ParseBoolVal(val)
}

// FormBools parses a list of bools from form parameters.
func (c *DefaultRouteContext) FormBools(name string) ([]bool, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseBoolSlice(vals)
}

// FormFloat32 parses a float32 from a form parameter.
func (c *DefaultRouteContext) FormFloat32(name string) (float32, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseFloat32Val(val)
}

// FormFloat32s parses a list of float32s from form parameters.
func (c *DefaultRouteContext) FormFloat32s(name string) ([]float32, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseFloat32Slice(vals)
}

// FormFloat64 parses a float64 from a form parameter.
func (c *DefaultRouteContext) FormFloat64(name string) (float64, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	return binder.ParseFloat64Val(val)
}

// FormFloat64s parses a list of float64s from form parameters.
func (c *DefaultRouteContext) FormFloat64s(name string) ([]float64, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return binder.ParseFloat64Slice(vals)
}
