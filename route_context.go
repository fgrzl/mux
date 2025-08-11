package mux

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/fgrzl/claims"
	"github.com/google/uuid"
)

// NewRouteContext creates a new RouteContext from an HTTP request and response writer.
func NewRouteContext(w http.ResponseWriter, r *http.Request) *DefaultRouteContext {
	return &DefaultRouteContext{
		Context:  r.Context(),
		response: w,
		request:  r,
	}
}

// RouteParams represents path parameters extracted from the URL pattern.
type RouteParams map[string]string

// RouteContext provides context for handling HTTP requests, including access to
// request/response objects, user authentication, route parameters, and services.
// RouteContext provides a comprehensive interface for HTTP request and response handling in mux.
// It exposes methods for accessing route parameters, queries, headers, cookies, forms, and user authentication.
// RouteContext is designed to be used per-request and is safe for concurrent use by multiple goroutines.
//
// Typical usage includes extracting parameters, binding request data, managing authentication, and sending responses.
//
// Implementations must embed context.Context and provide all methods below.
type RouteContext interface {
	// Context is the embedded request-scoped context for cancellation, deadlines, and values.
	context.Context

	// Response returns the current http.ResponseWriter for writing the response.
	Response() http.ResponseWriter
	// SetResponse replaces the http.ResponseWriter used to write the response.
	SetResponse(http.ResponseWriter)
	// Request returns the underlying *http.Request for this RouteContext.
	Request() *http.Request
	// Options returns the RouteOptions in effect for this request.
	Options() *RouteOptions

	// Core context methods
	// User returns the authenticated user principal, or nil if unauthenticated.
	User() claims.Principal
	// SetUser sets the authenticated user principal on the context.
	SetUser(user claims.Principal)
	// SetService stores a service instance by key for later retrieval in handlers.
	SetService(key ServiceKey, service any)
	// GetService retrieves a previously stored service by key.
	GetService(key ServiceKey) (any, bool)

	// Response methods - Basic HTTP responses
	// OK writes a 200 OK response with the provided model serialized (typically JSON).
	OK(model any)
	// JSON writes a JSON response with the given HTTP status code and model.
	JSON(status int, model any)
	// Plain writes a plain-text/bytes response with the given HTTP status code.
	Plain(status int, data []byte)
	// HTML writes an HTML response with the given HTTP status code and body.
	HTML(status int, html string)
	// NoContent writes a 204 No Content response.
	NoContent()
	// NotFound writes a 404 Not Found response.
	NotFound()
	// Created writes a 201 Created response with the provided model.
	Created(model any)
	// Accept writes a 202 Accepted response with the provided model.
	Accept(model any)

	// Response methods - Error responses
	// BadRequest writes a 400 Problem Details response with title and detail.
	BadRequest(title, detail string)
	// Unauthorized writes a 401 Unauthorized response.
	Unauthorized()
	// Forbidden writes a 403 Forbidden response with an optional message.
	Forbidden(message string)
	// Conflict writes a 409 Conflict Problem Details response.
	Conflict(title, detail string)
	// ServerError writes a 500 Internal Server Error Problem Details response.
	ServerError(title, detail string)
	// Problem writes a Problem Details response using the provided detail object.
	Problem(detail *ProblemDetails)

	// Response methods - File and redirects
	// File streams a file from disk to the response.
	File(filePath string)
	// Download sends a file as an attachment with the given download filename.
	Download(filePath string, filename string)
	// Redirect sends a redirect response with the given status code to the url.
	Redirect(status int, url string)
	// TemporaryRedirect sends a 307 Temporary Redirect to the given URL.
	TemporaryRedirect(url string)
	// PermanentRedirect sends a 308 Permanent Redirect to the given URL.
	PermanentRedirect(url string)

	// Request binding
	// Params returns the path parameters captured for this request.
	Params() RouteParams
	// Bind aggregates query, form/body, headers, and route params into the target struct.
	Bind(target any) error

	// Parameter methods
	// Param returns a single path parameter by name.
	Param(name string) (string, bool)
	// ParamUUID parses a path parameter as a UUID.
	ParamUUID(name string) (uuid.UUID, bool)
	// ParamInt parses a path parameter as int.
	ParamInt(name string) (int, bool)
	// ParamInt16 parses a path parameter as int16.
	ParamInt16(name string) (int16, bool)
	// ParamInt32 parses a path parameter as int32.
	ParamInt32(name string) (int32, bool)
	// ParamInt64 parses a path parameter as int64.
	ParamInt64(name string) (int64, bool)

	// Query parameter methods
	// QueryValue returns a query parameter by name.
	QueryValue(name string) (string, bool)
	// QueryValues returns all values for a query parameter.
	QueryValues(name string) ([]string, bool)
	// QueryUUID parses a query parameter as a UUID.
	QueryUUID(name string) (uuid.UUID, bool)
	// QueryUUIDs parses a query parameter as multiple UUIDs.
	QueryUUIDs(name string) ([]uuid.UUID, bool)
	// QueryInt parses a query parameter as int.
	QueryInt(name string) (int, bool)
	// QueryInts parses a query parameter as []int.
	QueryInts(name string) ([]int, bool)
	// QueryInt16 parses a query parameter as int16.
	QueryInt16(name string) (int16, bool)
	// QueryInt16s parses a query parameter as []int16.
	QueryInt16s(name string) ([]int16, bool)
	// QueryInt32 parses a query parameter as int32.
	QueryInt32(name string) (int32, bool)
	// QueryInt32s parses a query parameter as []int32.
	QueryInt32s(name string) ([]int32, bool)
	// QueryInt64 parses a query parameter as int64.
	QueryInt64(name string) (int64, bool)
	// QueryInt64s parses a query parameter as []int64.
	QueryInt64s(name string) ([]int64, bool)
	// QueryBool parses a query parameter as bool.
	QueryBool(name string) (bool, bool)
	// QueryBools parses a query parameter as []bool.
	QueryBools(name string) ([]bool, bool)
	// QueryFloat32 parses a query parameter as float32.
	QueryFloat32(name string) (float32, bool)
	// QueryFloat32s parses a query parameter as []float32.
	QueryFloat32s(name string) ([]float32, bool)
	// QueryFloat64 parses a query parameter as float64.
	QueryFloat64(name string) (float64, bool)
	// QueryFloat64s parses a query parameter as []float64.
	QueryFloat64s(name string) ([]float64, bool)
	// GetRedirectURL returns a safe URL to redirect to, falling back to defaultRedirect.
	GetRedirectURL(defaultRedirect string) string

	// Form parameter methods
	// FormValue returns a form value by name from parsed body.
	FormValue(name string) (string, bool)
	// FormValues returns all values for a form field.
	FormValues(name string) ([]string, bool)
	// FormUUID parses a form value as a UUID.
	FormUUID(name string) (uuid.UUID, bool)
	// FormUUIDs parses a form value as multiple UUIDs.
	FormUUIDs(name string) ([]uuid.UUID, bool)
	// FormInt parses a form value as int.
	FormInt(name string) (int, bool)
	// FormInts parses a form value as []int.
	FormInts(name string) ([]int, bool)
	// FormInt16 parses a form value as int16.
	FormInt16(name string) (int16, bool)
	// FormInt16s parses a form value as []int16.
	FormInt16s(name string) ([]int16, bool)
	// FormInt32 parses a form value as int32.
	FormInt32(name string) (int32, bool)
	// FormInt32s parses a form value as []int32.
	FormInt32s(name string) ([]int32, bool)
	// FormInt64 parses a form value as int64.
	FormInt64(name string) (int64, bool)
	// FormInt64s parses a form value as []int64.
	FormInt64s(name string) ([]int64, bool)
	// FormBool parses a form value as bool.
	FormBool(name string) (bool, bool)
	// FormBools parses a form value as []bool.
	FormBools(name string) ([]bool, bool)
	// FormFloat32 parses a form value as float32.
	FormFloat32(name string) (float32, bool)
	// FormFloat32s parses a form value as []float32.
	FormFloat32s(name string) ([]float32, bool)
	// FormFloat64 parses a form value as float64.
	FormFloat64(name string) (float64, bool)
	// FormFloat64s parses a form value as []float64.
	FormFloat64s(name string) ([]float64, bool)

	// Header methods
	// Header returns a request header value by name.
	Header(name string) (string, bool)
	// HeaderInt parses a request header as int.
	HeaderInt(name string) (int, bool)
	// HeaderUUID parses a request header as uuid.UUID.
	HeaderUUID(name string) (uuid.UUID, bool)
	// HeaderBool parses a request header as bool.
	HeaderBool(name string) (bool, bool)
	// HeaderFloat64 parses a request header as float64.
	HeaderFloat64(name string) (float64, bool)

	// Cookie methods
	// GetCookie returns the cookie value by name or an error if missing.
	GetCookie(name string) (string, error)
	// SetCookie sets a cookie with attributes including maxAge, path, domain, and flags.
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
	// ClearCookie removes the named cookie from the client.
	ClearCookie(name string)
	// Authenticate persists the user principal using the named cookie.
	Authenticate(cookieName string, user claims.Principal)
	// SignIn signs in the user and optionally redirects to the given URL.
	SignIn(user claims.Principal, redirectUrl string)
	// SignOut clears authentication state for the current user.
	SignOut()
}

type DefaultRouteContext struct {
	context.Context
	response    http.ResponseWriter
	request     *http.Request
	clientURL   *url.URL
	user        claims.Principal
	options     *RouteOptions
	params      RouteParams
	services    map[ServiceKey]any
	formsParsed bool
}

func (c *DefaultRouteContext) Response() http.ResponseWriter {
	return c.response
}

func (c *DefaultRouteContext) SetResponse(w http.ResponseWriter) {
	c.response = w
}

func (c *DefaultRouteContext) Request() *http.Request {
	return c.request
}

func (c *DefaultRouteContext) Options() *RouteOptions {
	return c.options
}

func (c *DefaultRouteContext) ClientURL() *url.URL {
	return c.clientURL
}

func (c *DefaultRouteContext) Params() RouteParams {
	return c.params
}

// GetUser returns the authenticated user from the RouteContext.
func (c *DefaultRouteContext) User() claims.Principal {
	return c.user
}

// SetUser sets the authenticated user in the RouteContext and updates the context for downstream access.
func (c *DefaultRouteContext) SetUser(user claims.Principal) {
	c.user = user
	c.Context = claims.WithUser(c.Context, user)
}

// SetService sets a service in the RouteContext.
func (c *DefaultRouteContext) SetService(key ServiceKey, svc any) {
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
func (c *DefaultRouteContext) GetService(key ServiceKey) (any, bool) {
	if c.services == nil {
		return nil, false
	}
	svc, ok := c.services[key]
	return svc, ok
}

// Bind collects input data from query parameters, request body, headers, and path parameters,
// then binds them to the provided model struct. It supports JSON and form-encoded data.
func (c *DefaultRouteContext) Bind(model any) error {
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

func (c *DefaultRouteContext) collectRequestData(staging map[string]any) error {
	switch c.request.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		c.collectQueryParams(staging)
	case http.MethodPut, http.MethodPost:
		return c.collectBodyData(staging)
	}
	return nil
}

func (c *DefaultRouteContext) collectQueryParams(staging map[string]any) {
	for key, values := range c.request.URL.Query() {
		addToStaging(staging, key, values)
	}
}

func (c *DefaultRouteContext) collectBodyData(staging map[string]any) error {
	c.request.Body = http.MaxBytesReader(c.Response(), c.request.Body, 1<<20) // 1MB max
	ct := c.request.Header.Get(HeaderContentType)
	switch {
	case ct == MimeFormURLEncoded:
		return c.collectFormData(staging)
	case strings.HasPrefix(ct, MimeJSON):
		return c.collectJSONBody(staging)
	default:
		return errors.New("unsupported content type")
	}
}

func (c *DefaultRouteContext) collectFormData(staging map[string]any) error {
	if err := c.request.ParseForm(); err != nil {
		return err
	}
	for key, values := range c.request.Form {
		addToStaging(staging, key, values)
	}
	return nil
}

func (c *DefaultRouteContext) collectJSONBody(staging map[string]any) error {
	bodyMap := make(map[string]any)
	decoder := json.NewDecoder(c.request.Body)
	if err := decoder.Decode(&bodyMap); err != nil {
		return err
	}
	for key, val := range bodyMap {
		staging[key] = val
	}
	return nil
}

func (c *DefaultRouteContext) collectHeaderData(staging map[string]any) {
	for key, headerValues := range c.request.Header {
		addToStaging(staging, key, headerValues)
	}
}

func (c *DefaultRouteContext) collectParamsData(staging map[string]any) {
	for key, paramValue := range c.params {
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
