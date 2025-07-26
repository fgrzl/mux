package mux

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/fgrzl/claims"
)

func NewRouteContext(w http.ResponseWriter, r *http.Request) *RouteContext {
	return &RouteContext{
		Context:  r.Context(),
		Response: w,
		Request:  r,
	}
}

type RouteParams map[string]string

type RouteContext struct {
	context.Context
	Response http.ResponseWriter
	Request  *http.Request
	User     claims.Principal
	Options  *RouteOptions
	Params   RouteParams

	services map[ServiceKey]any
}

// SetService sets a service in the RouteContext.
func (c *RouteContext) SetService(key ServiceKey, svc any) {
	if c.services == nil {
		c.services = make(map[ServiceKey]any)
	}
	c.services[key] = svc
}

// GetService retrieves a service from the RouteContext.
func (c *RouteContext) GetService(key ServiceKey) (any, bool) {
	if c.services == nil {
		return nil, false
	}
	svc, ok := c.services[key]
	return svc, ok
}

func (c *RouteContext) Bind(model any) error {
	staging := make(map[string]any)

	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		for key, values := range c.Request.URL.Query() {
			addToStaging(staging, key, values)
		}
	case http.MethodPut, http.MethodPost:
		c.Request.Body = http.MaxBytesReader(c.Response, c.Request.Body, 1<<20) // 1MB max
		ct := c.Request.Header.Get("Content-Type")
		if ct == "application/x-www-form-urlencoded" {
			if err := c.Request.ParseForm(); err != nil {
				return err
			}
			for key, values := range c.Request.Form {
				addToStaging(staging, key, values)
			}
		} else if strings.HasPrefix(ct, "application/json") {
			bodyMap := make(map[string]any)
			decoder := json.NewDecoder(c.Request.Body)
			if err := decoder.Decode(&bodyMap); err != nil {
				return err
			}
			for key, val := range bodyMap {
				staging[key] = val
			}
		} else {
			return errors.New("unsupported content type")
		}
	}

	for key, headerValues := range c.Request.Header {
		addToStaging(staging, key, headerValues)
	}
	for key, paramValue := range c.Params {
		staging[key] = paramValue
	}

	marshaledData, err := json.Marshal(staging)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshaledData, model)
}

func addToStaging(staging map[string]any, key string, values []string) {
	if len(values) == 1 {
		staging[key] = values[0]
	} else {
		staging[key] = values
	}
}

func getInstanceURI(r *http.Request) *string {
	instanceURI := r.RequestURI
	return &instanceURI
}

func parseSlice[T any](vals []string, parse func(string) (T, error)) ([]T, bool) {
	result := make([]T, 0, len(vals))
	for _, val := range vals {
		parsed, err := parse(val)
		if err != nil {
			return nil, false
		}
		result = append(result, parsed)
	}
	return result, true
}
