package mux

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type RouteContext struct {
	Response http.ResponseWriter
	Request  *http.Request
	User     ClaimsPrincipal
	Options  *RouteOptions
	Params   RouteParams
}

func NewRouteContext(w http.ResponseWriter, r *http.Request) *RouteContext {
	return &RouteContext{
		Response: w,
		Request:  r,
	}
}

type ProblemDetails struct {
	Type     string  `json:"type"`               // A URI reference to the error type
	Title    string  `json:"title"`              // A brief summary of the error
	Status   int     `json:"status"`             // HTTP status code
	Detail   string  `json:"detail"`             // Specific details about the error
	Instance *string `json:"instance,omitempty"` // A URI identifying the error instance, nullable
}

type RouteParams map[string]string

type RouteOptions struct {
	Method         string
	Pattern        string
	Handler        HandlerFunc
	AllowAnonymous bool
	Roles          []string
	Scopes         []string
	Permissions    []string
}

func (c *RouteContext) Bind(model any) error {
	// Populate the values map with query parameters, headers, and body content
	staging := make(map[string]any)

	// Handle different HTTP methods
	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		// Process query string for GET, HEAD, DELETE requests
		for key, values := range c.Request.URL.Query() {
			addToStaging(staging, key, values)
		}

	case http.MethodPut, http.MethodPost:

		ct := c.Request.Header.Get("Content-Type")
		// Handle POST/PUT request bodies, both for form data and JSON content
		if ct == "application/x-www-form-urlencoded" {
			// Handle form URL encoded
			if err := c.Request.ParseForm(); err != nil {
				return err
			}
			// Populate the map with form values
			for key, values := range c.Request.Form {
				addToStaging(staging, key, values)
			}
		} else if strings.HasPrefix(ct, "application/json") {
			// Handle JSON content
			bodyMap := make(map[string]any)
			decoder := json.NewDecoder(c.Request.Body)
			if err := decoder.Decode(&bodyMap); err != nil {
				return err
			}
			// Add the JSON fields into the values map
			for key, val := range bodyMap {
				staging[key] = val
			}
		} else {
			// Unsupported content type
			return errors.New("unsupported content type")
		}
	}

	// Check headers
	for key, headerValues := range c.Request.Header {
		addToStaging(staging, key, headerValues)
	}

	// Extract route parameters (assuming these are set somewhere in the context)
	for key, paramValue := range c.Params {
		staging[key] = paramValue
	}

	// Now that we have populated the values map with all necessary info, unmarshal into the model
	// Marshal the values map into JSON, then unmarshal into the provided model
	marshaledData, err := json.Marshal(staging)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(marshaledData, model); err != nil {
		return err
	}

	return nil
}

func (c *RouteContext) Param(name string) (string, bool) {
	if val, ok := c.Params[name]; ok {
		return val, ok
	}
	return "", false
}

func (c *RouteContext) QueryValue(name string) (string, bool) {
	query := c.Request.URL.Query()
	if val, ok := query[name]; ok {
		return val[0], ok
	}
	return "", false
}

func (c *RouteContext) QueryValues(name string) ([]string, bool) {
	query := c.Request.URL.Query()
	if val, ok := query[name]; ok {
		return val, ok
	}
	return nil, false
}

func (c *RouteContext) ServerError(title, detail string) {
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusInternalServerError,
		Instance: getInstanceURI(c.Request),
	})
}

func (c *RouteContext) BadRequest(title, detail string) {
	problem := ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusBadRequest,
		Instance: getInstanceURI(c.Request),
	}

	c.Problem(&problem)
}

func (c *RouteContext) Conflict(title, detail string) {
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusConflict,
		Instance: getInstanceURI(c.Request),
	})
}

func (c *RouteContext) OK(model any) {
	r := c.Response
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(http.StatusOK)
	json.NewEncoder(r).Encode(model)
}

func (c *RouteContext) Created(model any) {
	r := c.Response
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(http.StatusCreated)
	json.NewEncoder(r).Encode(model)
}

func (c *RouteContext) Accept(model any) {
	r := c.Response
	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(http.StatusAccepted)
}

func (c *RouteContext) NoContent() {
	r := c.Response
	r.WriteHeader(http.StatusNoContent)
}

func (c *RouteContext) NotFound() {
	http.NotFound(c.Response, c.Request)
}

func (c *RouteContext) Unauthorized() {
	http.Error(c.Response, "", http.StatusUnauthorized)
}

func (c *RouteContext) Forbidden(message string) {
	http.Error(c.Response, "", http.StatusForbidden)
}

func (c *RouteContext) Problem(problem *ProblemDetails) {
	r := c.Response
	r.Header().Set("Content-Type", "application/problem+json")
	r.WriteHeader(problem.Status)
	json.NewEncoder(r).Encode(problem)
}

func addToStaging(staging map[string]any, key string, values []string) {
	// Store single or multiple values in the staging map
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
