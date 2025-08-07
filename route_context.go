package mux

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/fgrzl/claims"
)

// NewRouteContext creates a new RouteContext from an HTTP request and response writer.
func NewRouteContext(w http.ResponseWriter, r *http.Request) *RouteContext {
	return &RouteContext{
		Context:  r.Context(),
		Response: w,
		Request:  r,
	}
}

// RouteParams represents path parameters extracted from the URL pattern.
type RouteParams map[string]string

// RouteContext provides context for handling HTTP requests, including access to
// request/response objects, user authentication, route parameters, and services.
type RouteContext struct {
	context.Context
	Response  http.ResponseWriter
	Request   *http.Request
	ClientURL *url.URL
	User      claims.Principal
	Options   *RouteOptions
	Params    RouteParams

	services    map[ServiceKey]any
	formsParsed bool
}

// SetUser sets the authenticated user in the RouteContext and updates the context for downstream access.
func (c *RouteContext) SetUser(user claims.Principal) {
	c.User = user
	c.Context = claims.WithUser(c.Context, user)
}

// SetService sets a service in the RouteContext.
func (c *RouteContext) SetService(key ServiceKey, svc any) {
	// Validate inputs
	if key == "" {
		return // Do not allow empty keys
	}
	if svc == nil {
		return // Do not allow nil services
	}
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

// Bind collects input data from query parameters, request body, headers, and path parameters,
// then binds them to the provided model struct. It supports JSON and form-encoded data.
func (c *RouteContext) Bind(model any) error {
	staging := make(map[string]any)

	if err := c.collectRequestData(staging); err != nil {
		return err
	}
	c.collectHeaderData(staging)
	c.collectParamsData(staging)

	marshaledData, err := json.Marshal(staging)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshaledData, model)
}

func (c *RouteContext) collectRequestData(staging map[string]any) error {
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		c.collectQueryParams(staging)
	case http.MethodPut, http.MethodPost:
		return c.collectBodyData(staging)
	}
	return nil
}

func (c *RouteContext) collectQueryParams(staging map[string]any) {
	for key, values := range c.Request.URL.Query() {
		addToStaging(staging, key, values)
	}
}

func (c *RouteContext) collectBodyData(staging map[string]any) error {
	c.Request.Body = http.MaxBytesReader(c.Response, c.Request.Body, 1<<20) // 1MB max
	ct := c.Request.Header.Get(HeaderContentType)
	switch {
	case ct == MimeFormURLEncoded:
		return c.collectFormData(staging)
	case strings.HasPrefix(ct, MimeJSON):
		return c.collectJSONBody(staging)
	default:
		return errors.New("unsupported content type")
	}
}

func (c *RouteContext) collectFormData(staging map[string]any) error {
	if err := c.Request.ParseForm(); err != nil {
		return err
	}
	for key, values := range c.Request.Form {
		addToStaging(staging, key, values)
	}
	return nil
}

func (c *RouteContext) collectJSONBody(staging map[string]any) error {
	bodyMap := make(map[string]any)
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&bodyMap); err != nil {
		return err
	}
	for key, val := range bodyMap {
		staging[key] = val
	}
	return nil
}

func (c *RouteContext) collectHeaderData(staging map[string]any) {
	for key, headerValues := range c.Request.Header {
		addToStaging(staging, key, headerValues)
	}
}

func (c *RouteContext) collectParamsData(staging map[string]any) {
	for key, paramValue := range c.Params {
		staging[key] = paramValue
	}
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
