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
// RouteContext defines the contract for a route context that provides
// HTTP request and response handling capabilities, parameter extraction,
// and user authentication.
type RouteContext interface {
	context.Context
	Response() http.ResponseWriter
	SetResponse(http.ResponseWriter)
	Request() *http.Request
	Options() *RouteOptions

	// Core context methods
	User() claims.Principal
	SetUser(user claims.Principal)
	SetService(key ServiceKey, service any)
	GetService(key ServiceKey) (any, bool)

	// Response methods - Basic HTTP responses
	OK(model any)
	JSON(status int, model any)
	Plain(status int, data []byte)
	HTML(status int, html string)
	NoContent()
	NotFound()
	Created(model any)
	Accept(model any)

	// Response methods - Error responses
	BadRequest(title, detail string)
	Unauthorized()
	Forbidden(message string)
	Conflict(title, detail string)
	ServerError(title, detail string)
	Problem(detail *ProblemDetails)

	// Response methods - File and redirects
	File(filePath string)
	Download(filePath string, filename string)
	Redirect(status int, url string)
	TemporaryRedirect(url string)
	PermanentRedirect(url string)

	// Request binding
	Params() RouteParams
	Bind(target any) error

	// Parameter methods
	Param(name string) (string, bool)
	ParamUUID(name string) (uuid.UUID, bool)
	ParamInt(name string) (int, bool)
	ParamInt16(name string) (int16, bool)
	ParamInt32(name string) (int32, bool)
	ParamInt64(name string) (int64, bool)

	// Query parameter methods
	QueryValue(name string) (string, bool)
	QueryValues(name string) ([]string, bool)
	QueryUUID(name string) (uuid.UUID, bool)
	QueryUUIDs(name string) ([]uuid.UUID, bool)
	QueryInt(name string) (int, bool)
	QueryInts(name string) ([]int, bool)
	QueryInt16(name string) (int16, bool)
	QueryInt16s(name string) ([]int16, bool)
	QueryInt32(name string) (int32, bool)
	QueryInt32s(name string) ([]int32, bool)
	QueryInt64(name string) (int64, bool)
	QueryInt64s(name string) ([]int64, bool)
	QueryBool(name string) (bool, bool)
	QueryBools(name string) ([]bool, bool)
	QueryFloat32(name string) (float32, bool)
	QueryFloat32s(name string) ([]float32, bool)
	QueryFloat64(name string) (float64, bool)
	QueryFloat64s(name string) ([]float64, bool)
	GetRedirectURL(defaultRedirect string) string

	// Form parameter methods
	FormValue(name string) (string, bool)
	FormValues(name string) ([]string, bool)
	FormUUID(name string) (uuid.UUID, bool)
	FormUUIDs(name string) ([]uuid.UUID, bool)
	FormInt(name string) (int, bool)
	FormInts(name string) ([]int, bool)
	FormInt16(name string) (int16, bool)
	FormInt16s(name string) ([]int16, bool)
	FormInt32(name string) (int32, bool)
	FormInt32s(name string) ([]int32, bool)
	FormInt64(name string) (int64, bool)
	FormInt64s(name string) ([]int64, bool)
	FormBool(name string) (bool, bool)
	FormBools(name string) ([]bool, bool)
	FormFloat32(name string) (float32, bool)
	FormFloat32s(name string) ([]float32, bool)
	FormFloat64(name string) (float64, bool)
	FormFloat64s(name string) ([]float64, bool)

	// Header methods
	Header(name string) (string, bool)
	HeaderInt(name string) (int, bool)
	HeaderUUID(name string) (uuid.UUID, bool)
	HeaderBool(name string) (bool, bool)
	HeaderFloat64(name string) (float64, bool)

	// Cookie methods
	GetCookie(name string) (string, error)
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool)
	ClearCookie(name string)
	Authenticate(cookieName string, user claims.Principal)
	SignIn(user claims.Principal, redirectUrl string)
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
