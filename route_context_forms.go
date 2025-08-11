package mux

import (
	"strconv"
	"strings"

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
	id, err := uuid.Parse(val)
	return id, err == nil
}

// FormUUIDs parses a list of UUIDs from form parameters.
func (c *DefaultRouteContext) FormUUIDs(name string) ([]uuid.UUID, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, uuid.Parse)
}

// FormInt parses an int from a form parameter.
func (c *DefaultRouteContext) FormInt(name string) (int, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(val)
	return n, err == nil
}

// FormInts parses a list of ints from form parameters.
func (c *DefaultRouteContext) FormInts(name string) ([]int, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, strconv.Atoi)
}

// FormInt16 parses an int16 from a form parameter.
func (c *DefaultRouteContext) FormInt16(name string) (int16, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 16)
	return int16(n), err == nil
}

// FormInt16s parses a list of int16s from form parameters.
func (c *DefaultRouteContext) FormInt16s(name string) ([]int16, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, func(s string) (int16, error) {
		n, err := strconv.ParseInt(s, 10, 16)
		return int16(n), err
	})
}

// FormInt32 parses an int32 from a form parameter.
func (c *DefaultRouteContext) FormInt32(name string) (int32, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 32)
	return int32(n), err == nil
}

// FormInt32s parses a list of int32s from form parameters.
func (c *DefaultRouteContext) FormInt32s(name string) ([]int32, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, func(s string) (int32, error) {
		n, err := strconv.ParseInt(s, 10, 32)
		return int32(n), err
	})
}

// FormInt64 parses an int64 from a form parameter.
func (c *DefaultRouteContext) FormInt64(name string) (int64, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(val, 10, 64)
	return n, err == nil
}

// FormInt64s parses a list of int64s from form parameters.
func (c *DefaultRouteContext) FormInt64s(name string) ([]int64, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
}

// FormBool parses a bool from a form parameter.
func (c *DefaultRouteContext) FormBool(name string) (bool, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return false, false
	}
	b, err := strconv.ParseBool(val)
	return b, err == nil
}

// FormBools parses a list of bools from form parameters.
func (c *DefaultRouteContext) FormBools(name string) ([]bool, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, strconv.ParseBool)
}

// FormFloat32 parses a float32 from a form parameter.
func (c *DefaultRouteContext) FormFloat32(name string) (float32, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(val, 32)
	return float32(f), err == nil
}

// FormFloat32s parses a list of float32s from form parameters.
func (c *DefaultRouteContext) FormFloat32s(name string) ([]float32, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, func(s string) (float32, error) {
		f, err := strconv.ParseFloat(s, 32)
		return float32(f), err
	})
}

// FormFloat64 parses a float64 from a form parameter.
func (c *DefaultRouteContext) FormFloat64(name string) (float64, bool) {
	val, ok := c.FormValue(name)
	if !ok {
		return 0, false
	}
	f, err := strconv.ParseFloat(val, 64)
	return f, err == nil
}

// FormFloat64s parses a list of float64s from form parameters.
func (c *DefaultRouteContext) FormFloat64s(name string) ([]float64, bool) {
	vals, ok := c.FormValues(name)
	if !ok {
		return nil, false
	}
	return parseSlice(vals, func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
}
