package routing

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/binder"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/cookiekit"
	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/google/uuid"
)

// ServiceKey is an alias to the common.ServiceKey so the routing package can
// refer to it unqualified throughout the codebase.
type ServiceKey = common.ServiceKey

// RouteParams represents path parameters extracted from the URL pattern.
type RouteParams map[string]string

// RouteContext provides context for handling HTTP requests, including access to
// request/response objects, user authentication, route parameters, and services.
// RouteContext provides a comprehensive interface for HTTP request and response handling in mux.
// It exposes methods for accessing route parameters, queries, headers, cookies, forms, and user authentication.
//
// THREAD SAFETY: RouteContext instances are NOT safe for concurrent use by multiple goroutines.
// Each RouteContext is bound to a single request lifecycle and may be pooled for reuse.
// If you need to use the context in a goroutine that may outlive the request handler,
// use the Detach() function to create a safe background copy:
//
//	go func(ctx *routing.DefaultRouteContext) {
//	    // Safe to read request metadata, params, services, and user here.
//	    // Response writes are discarded on detached contexts.
//	}(routing.Detach(c))
//
// WARNING: After the request handler returns, the original RouteContext may be recycled.
// Accessing it after this point leads to undefined behavior. Always use Detach() for
// background work.
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
	// SetRequest replaces the current *http.Request and updates the embedded context.
	SetRequest(*http.Request)
	// Options returns the RouteOptions in effect for this request.
	Options() *RouteOptions

	// Core context methods
	// User returns the authenticated user principal, or nil if unauthenticated.
	User() claims.Principal
	// SetUser sets the authenticated user principal on the context.
	SetUser(user claims.Principal)
	// SetContextValue adds a key-value pair to the request context.
	// This properly updates both the embedded context and the underlying request
	// to ensure consistency when accessing values via either c or c.Request().Context().
	SetContextValue(key, value any)
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
	// MovedPermanently sends a 301 Moved Permanently redirect to the given URL.
	MovedPermanently(url string)
	// Found sends a 302 Found redirect to the given URL.
	Found(url string)
	// SeeOther sends a 303 See Other redirect to the given URL (POST->GET pattern).
	SeeOther(url string)
	// TemporaryRedirect sends a 307 Temporary Redirect to the given URL.
	TemporaryRedirect(url string)
	// PermanentRedirect sends a 308 Permanent Redirect to the given URL.
	PermanentRedirect(url string)

	// Request binding
	// Bind aggregates query, form/body, headers, and route params into the target struct.
	Bind(target any) error

	// Parameter methods
	// ParamsSlice returns the optimized slice-based parameter storage.
	ParamsSlice() *Params
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
	// The SameSite attribute is optional and defaults to Lax when omitted.
	SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool, sameSite ...http.SameSite)
	// ClearCookie removes the named cookie from the client.
	ClearCookie(name string)
	// Authenticate persists the user principal using the named cookie.
	Authenticate(cookieName string, user claims.Principal, opts ...cookiekit.CookieOption)
	// SignIn signs in the user and optionally redirects to the given URL.
	SignIn(user claims.Principal, redirectUrl string, opts ...cookiekit.CookieOption)
	// SignOut clears authentication state for the current user.
	SignOut(redirectUrl string)
}

// NewRouteContext creates a new RouteContext from an HTTP request and response writer.
func NewRouteContext(w http.ResponseWriter, r *http.Request) *DefaultRouteContext {
	return &DefaultRouteContext{
		Context:  r.Context(),
		response: w,
		request:  r,
		// paramsSlice is nil by default, lazily initialized when needed
		// New contexts created directly are not pooled
		wasPooled: false,
	}
}

// contextPool is a pool of DefaultRouteContext objects used to minimize
// per-request allocations. Callers should obtain a context instance using
// AcquireContext(w,r) and return it with ReleaseContext when done. The
// pool stores zeroed or otherwise-reset DefaultRouteContext values and
// the Acquire/Release helpers ensure fields are initialized/cleared as
// appropriate to avoid leaking request-scoped data between requests.
var contextPool = sync.Pool{
	New: func() any { return &DefaultRouteContext{} },
}

// AcquireContext gets a DefaultRouteContext from the pool and initializes it
// for the provided http.ResponseWriter and *http.Request.
func AcquireContext(w http.ResponseWriter, r *http.Request) *DefaultRouteContext {
	c := contextPool.Get().(*DefaultRouteContext)
	// Reset minimal fields; leave maps nil to avoid extra work unless needed
	c.Context = r.Context()
	c.response = w
	c.request = r
	c.clientURL = nil
	c.user = nil
	c.options = nil
	c.services = nil
	c.formsParsed = false
	c.paramIndex = nil
	c.maxBodyBytes = 0
	c.bodyLimitApplied = false
	// Acquire paramsSlice from pool for optimized parameter storage
	c.paramsSlice = AcquireParams()
	// Mark as pooled so ReleaseContext knows to return it
	c.wasPooled = true
	return c
}

// ReleaseContext resets sensitive references and returns the context to the pool.
func ReleaseContext(c *DefaultRouteContext) {
	if c == nil {
		return
	}
	// Clear references to avoid leaks between requests
	c.Context = nil
	c.response = nil
	c.request = nil
	c.clientURL = nil
	c.user = nil
	c.options = nil
	if c.paramsSlice != nil {
		// return params slice to pool
		ReleaseParams(c.paramsSlice)
		c.paramsSlice = nil
	}
	c.services = nil
	c.formsParsed = false
	c.paramIndex = nil
	c.maxBodyBytes = 0
	c.bodyLimitApplied = false
	// Only return to the pool if this instance was obtained from it.
	if c.wasPooled {
		// reset the flag to avoid double-put if ReleaseContext is called
		c.wasPooled = false
		contextPool.Put(c)
	}
}

// Detach clones the provided RouteContext into a new non-pooled DefaultRouteContext
// that is safe to use in goroutines that outlive the request. The detached copy
// uses a background context, preserves framework-level state such as user,
// params, services, and route options, and replaces the response writer with a
// no-op writer so background work cannot accidentally write to the original
// HTTP response. The returned context must not be released via ReleaseContext.
func Detach(c RouteContext) *DefaultRouteContext {
	if c == nil {
		return nil
	}
	d, ok := c.(*DefaultRouteContext)
	if !ok {
		return nil
	}
	baseCtx := context.Background()
	if d.user != nil {
		baseCtx = claims.WithUser(baseCtx, d.user)
	}

	var reqClone *http.Request
	if d.request != nil {
		reqClone = d.request.Clone(baseCtx)
		if reqClone.Body != nil {
			reqClone.Body = http.NoBody
		}
		baseCtx = reqClone.Context()
	}
	clone := &DefaultRouteContext{
		Context:      baseCtx,
		response:     &detachedResponseWriter{},
		request:      reqClone,
		clientURL:    d.clientURL,
		user:         d.user,
		options:      d.options,
		formsParsed:  d.formsParsed,
		wasPooled:    false,
		maxBodyBytes: d.maxBodyBytes,
	}
	if d.paramsSlice != nil && d.paramsSlice.Len() > 0 {
		// Clone paramsSlice to avoid sharing references
		clonedSlice := make(Params, d.paramsSlice.Len())
		copy(clonedSlice, *d.paramsSlice)
		clone.paramsSlice = &clonedSlice
	}
	if d.services != nil {
		clone.services = make(map[ServiceKey]any, len(d.services))
		for k, v := range d.services {
			clone.services[k] = v
		}
	}
	if d.paramIndex != nil {
		clone.paramIndex = make(map[string]*openapi.ParameterObject, len(d.paramIndex))
		for k, v := range d.paramIndex {
			clone.paramIndex[k] = v
		}
	}
	return clone
}

type DefaultRouteContext struct {
	context.Context
	response    http.ResponseWriter
	request     *http.Request
	clientURL   *url.URL
	user        claims.Principal
	options     *RouteOptions
	paramsSlice *Params // Optimized slice-based parameter storage
	services    map[ServiceKey]any
	formsParsed bool
	// wasPooled indicates whether this instance was obtained from the pool.
	// It is used to prevent double-returns to the object pool: if true,
	// ReleaseContext will return it to the pool; otherwise it will not be pooled.
	wasPooled bool
	// runtime cache for quick parameter lookups (key: strings.ToLower(in+":"+name))
	paramIndex map[string]*openapi.ParameterObject
	// maxBodyBytes limits body size for bind operations. 0 means default (1MB).
	maxBodyBytes int64
	// bodyLimitApplied tracks whether MaxBytesReader has been applied to prevent double-wrapping.
	bodyLimitApplied bool
}

type detachedResponseWriter struct{ header http.Header }

func (w *detachedResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *detachedResponseWriter) Write(p []byte) (int, error) { return len(p), nil }

func (w *detachedResponseWriter) WriteHeader(statusCode int) {}

func (c *DefaultRouteContext) Response() http.ResponseWriter {
	return c.response
}

func (c *DefaultRouteContext) SetResponse(w http.ResponseWriter) {
	c.response = w
}

func (c *DefaultRouteContext) Request() *http.Request {
	return c.request
}

// SetRequest replaces the current request and updates the embedded context.
func (c *DefaultRouteContext) SetRequest(r *http.Request) {
	c.request = r
	if r != nil {
		c.Context = r.Context()
	}
}

func (c *DefaultRouteContext) Options() *RouteOptions {
	return c.options
}

// SetOptions sets the current RouteOptions on the context.
func (c *DefaultRouteContext) SetOptions(o *RouteOptions) {
	c.options = o
}

// SetParamsSlice sets the path parameters using the optimized slice-based storage.
func (c *DefaultRouteContext) SetParamsSlice(p *Params) {
	c.paramsSlice = p
}

// ClientURL returns the configured client URL for building absolute links.
func (c *DefaultRouteContext) ClientURL() *url.URL {
	return c.clientURL
}

// SetClientURL sets the client URL used for absolute URL generation.
func (c *DefaultRouteContext) SetClientURL(u *url.URL) {
	c.clientURL = u
}

// SetMaxBodyBytes sets the maximum allowed request body size for this context.
// A value <= 0 causes a default of 1MB to be applied during binding.
func (c *DefaultRouteContext) SetMaxBodyBytes(n int64) { c.maxBodyBytes = n }

// ParamsSlice returns the optimized slice-based parameter storage.
func (c *DefaultRouteContext) ParamsSlice() *Params {
	return c.paramsSlice
}

// GetUser returns the authenticated user from the RouteContext.
func (c *DefaultRouteContext) User() claims.Principal {
	return c.user
}

// SetUser sets the authenticated user in the RouteContext and updates the context for downstream access.
func (c *DefaultRouteContext) SetUser(user claims.Principal) {
	c.user = user
	c.Context = claims.WithUser(c.Context, user)
	// Update the request's context so c.Request().Context() returns the same values
	if c.request != nil {
		c.request = c.request.WithContext(c.Context)
	}
}

// SetContextValue adds a key-value pair to the request context.
// This properly updates both the embedded context and the underlying request
// to ensure consistency when accessing values via either c or c.Request().Context().
func (c *DefaultRouteContext) SetContextValue(key, value any) {
	c.Context = context.WithValue(c.Context, key, value)
	// Update the request's context so c.Request().Context() returns the same values
	if c.request != nil {
		c.request = c.request.WithContext(c.Context)
	}
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

// Bind collects input data from request sources according to the HTTP method,
// then binds them to the provided model struct.
//
// Query parameters, headers, and path parameters are bound for all request
// methods. Request bodies are bound only for POST, PUT, and PATCH using JSON,
// application/x-www-form-urlencoded, or multipart/form-data payloads. Methods
// such as GET, HEAD, and DELETE do not bind request bodies.
//
// If a request declares RequestBody.Required=true but sends an empty body, Bind returns ErrMissingBody.
// Callers can check for this using errors.Is(err, ErrMissingBody) and map it to a custom response,
// or let it propagate to a central error handler that converts it to a 400 Problem Details response.
func (c *DefaultRouteContext) Bind(model any) error {
	staging := make(map[string]any)

	if err := c.collectRequestData(staging); err != nil {
		// Automatically map ErrMissingBody to a 400 Problem Details response
		// so handlers don't need to check for it explicitly. This provides
		// consistent error handling across the framework.
		if IsMissingBodyError(err) {
			c.Problem(&ProblemDetails{
				Type:   "request-body-required",
				Title:  "Bad Request",
				Status: http.StatusBadRequest,
				Detail: "Request body is required",
			})
			return err
		}
		return err
	}
	if err := c.collectHeaderData(staging); err != nil {
		return err
	}
	if err := c.collectParamsData(staging); err != nil {
		return err
	}

	// If the JSON body was a top-level array, collectJSONBody stores it under
	// the special key "__root_json_array". In that case we should marshal the
	// array value directly instead of the staging map so the target model can
	// be an array/slice type.
	if root, ok := staging["__root_json_array"]; ok {
		marshaledData, err := json.Marshal(root)
		if err != nil {
			return err
		}
		return json.Unmarshal(marshaledData, model)
	}

	marshaledData, err := json.Marshal(staging)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshaledData, model)
}

func (c *DefaultRouteContext) collectRequestData(staging map[string]any) error {
	if err := c.collectQueryParams(staging); err != nil {
		return err
	}
	if !methodAllowsBodyBinding(c.request.Method) {
		return nil
	}
	return c.collectBodyData(staging)
}

func methodAllowsBodyBinding(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

func (c *DefaultRouteContext) collectQueryParams(staging map[string]any) error {
	for rawKey, values := range c.request.URL.Query() {
		// deep-object handling: dot-notation or bracket-notation
		if root, path := parseDeepKey(rawKey); len(path) > 0 {
			// only handle deep objects when root parameter is declared as object
			if param := c.lookupParameter(root, "query"); param != nil {
				isObject := false
				if param.Schema != nil && param.Schema.Type == "object" {
					isObject = true
				}
				if !isObject && param.Example != nil {
					exT := reflect.TypeOf(param.Example)
					if exT.Kind() == reflect.Ptr {
						exT = exT.Elem()
					}
					if exT.Kind() == reflect.Struct {
						isObject = true
					}
				}
				if isObject {
					// ensure values is split for CSV if schema for property indicates array
					// for dot/bracket keys path[0] is property name
					propName := path[0]
					propSchema := (func() *openapi.Schema {
						if param.Schema != nil {
							return param.Schema.Properties[propName]
						}
						return nil
					})()
					// if property expects array and single CSV string provided, split
					if propSchema != nil && propSchema.Type == "array" && len(values) == 1 && strings.Contains(values[0], ",") {
						values = splitAndTrim(values[0])
					}
					var parsed any
					var err error
					// prefer schema-based parsing
					if propSchema != nil {
						parsed, err = binder.ParseValueBySchema(values, propSchema)
					} else if param.Example != nil {
						// try to locate the example field within the Example struct and parse by that example
						exVal := reflect.ValueOf(param.Example)
						if exVal.Kind() == reflect.Ptr {
							exVal = exVal.Elem()
						}
						if exVal.IsValid() && exVal.Kind() == reflect.Struct {
							// find field by json tag or name
							exType := exVal.Type()
							var fieldExample any
							for i := 0; i < exType.NumField(); i++ {
								f := exType.Field(i)
								tag := f.Tag.Get("json")
								name := f.Name
								if tag != "" {
									// tag may contain options like `json:"name,omitempty"`
									parts := strings.Split(tag, ",")
									if parts[0] == propName {
										fieldExample = exVal.Field(i).Interface()
										break
									}
								}
								if strings.EqualFold(name, propName) {
									fieldExample = exVal.Field(i).Interface()
									break
								}
							}
							if fieldExample != nil {
								p := &openapi.ParameterObject{Example: fieldExample}
								if parsedVal, ok := binder.ParseByExample(values[0], p); ok {
									parsed = parsedVal
								}
							}
						}
					} else {
						parsed, err = binder.ParseValueBySchema(values, propSchema)
					}
					if err != nil {
						return fmt.Errorf("query param %q.%s: %w", root, propName, err)
					}
					// set nested map structure
					setNestedMap(staging, root, path, parsed)
					continue
				}
			}
		}

		key := rawKey
		// try to find a declared parameter for this query key
		if param := c.lookupParameter(key, "query"); param != nil {
			if handled, err := binder.ProcessParamAndSet(staging, key, values, "query", param); err != nil {
				return err
			} else if handled {
				continue
			}
		}
		addToStaging(staging, key, values)
	}
	return nil
}

func (c *DefaultRouteContext) collectBodyData(staging map[string]any) error {
	// Apply MaxBytesReader only once to prevent double-wrapping which can cause
	// the effective limit to be applied multiple times incorrectly.
	if !c.bodyLimitApplied {
		maxBytes := c.maxBodyBytes
		if maxBytes <= 0 {
			maxBytes = 1 << 20 // 1MB default
		}
		c.request.Body = http.MaxBytesReader(c.Response(), c.request.Body, maxBytes)
		c.bodyLimitApplied = true
	}

	// If the route explicitly requires a request body according to its
	// OpenAPI RequestBody.Required flag, perform a light-weight presence
	// check before attempting to parse the body. This gives a stable,
	// descriptive error for completely empty bodies instead of low-level
	// decoder errors like io.EOF. The presence check applies across
	// content types (JSON and form) and only runs when RequestBody.Required
	// is true.
	if c.options != nil && c.options.RequestBody != nil && c.options.RequestBody.Required {
		// Wrap the current Body in a buffered reader and attempt to Peek(1).
		// If Peek returns io.EOF, the body is empty.
		br := bufio.NewReader(c.request.Body)
		if _, err := br.Peek(1); err != nil {
			if err == io.EOF {
				return ErrMissingBody
			}
			// Propagate other read errors
			return err
		}
		// Put a nondestructive wrapper back so downstream readers (json.Decoder,
		// ParseForm) can consume the body normally.
		c.request.Body = io.NopCloser(br)
	}

	ct := c.request.Header.Get(common.HeaderContentType)
	switch {
	case strings.HasPrefix(ct, common.MimeFormURLEncoded), strings.HasPrefix(ct, common.MimeMultipartFormData):
		return c.collectFormData(staging)
	case strings.HasPrefix(ct, common.MimeJSON):
		return c.collectJSONBody(staging)
	default:
		return errors.New("unsupported content type")
	}
}

func (c *DefaultRouteContext) collectFormData(staging map[string]any) error {
	ct := c.request.Header.Get(common.HeaderContentType)
	if strings.HasPrefix(ct, common.MimeMultipartFormData) {
		if err := c.request.ParseMultipartForm(32 << 20); err != nil {
			return err
		}
		if c.request.MultipartForm == nil {
			return nil
		}
		for key, values := range c.request.MultipartForm.Value {
			addToStaging(staging, key, values)
		}
	} else {
		if err := c.request.ParseForm(); err != nil {
			return err
		}
		for key, values := range c.request.PostForm {
			addToStaging(staging, key, values)
		}
	}
	return nil
}

func (c *DefaultRouteContext) collectJSONBody(staging map[string]any) error {
	// Read and decode JSON body. Support object roots for mixed-source binding
	// and array roots for slice targets.
	decoder := json.NewDecoder(c.request.Body)
	// Use interface{} to accept either map or slice
	var bodyAny any
	if err := decoder.Decode(&bodyAny); err != nil {
		// Treat empty body as acceptable when request body is not required.
		// JSON decoder returns io.EOF for an empty body; normalize that to
		// no-op so callers receive zero-values instead of a low-level error.
		if err == io.EOF {
			return nil
		}
		return err
	}
	switch v := bodyAny.(type) {
	case map[string]any:
		for key, val := range v {
			staging[key] = val
		}
		return nil
	case []any:
		staging["__root_json_array"] = v
		return nil
	default:
		return fmt.Errorf("unsupported JSON root type %T", v)
	}
}

func (c *DefaultRouteContext) collectHeaderData(staging map[string]any) error {
	for key, headerValues := range c.request.Header {
		// process header parameter consistently via helper
		if handled, hadParam, err := c.processParamForStaging(staging, key, headerValues, "header"); err != nil {
			return err
		} else if hadParam && handled {
			continue
		}
		addToStaging(staging, key, headerValues)
	}
	return nil
}

func (c *DefaultRouteContext) collectParamsData(staging map[string]any) error {
	if c.paramsSlice != nil {
		for i := 0; i < c.paramsSlice.Len(); i++ {
			p := (*c.paramsSlice)[i]
			if handled, hadParam, err := c.processParamForStaging(staging, p.Key, []string{p.Value}, "path"); err != nil {
				return err
			} else if hadParam && handled {
				continue
			}
			staging[p.Key] = p.Value
		}
	}
	return nil
}

// processParamForStaging centralizes the common pattern of looking up a declared
// parameter and invoking binder.ProcessParamAndSet. It returns three values:
// (handled, hadParam, err) where "hadParam" indicates whether a ParameterObject
// existed for the given key/in and "handled" indicates whether the binder
// handled and wrote a value into the staging map.
func (c *DefaultRouteContext) processParamForStaging(staging map[string]any, key string, values []string, in string) (bool, bool, error) {
	if param := c.lookupParameter(key, in); param != nil {
		handled, err := binder.ProcessParamAndSet(staging, key, values, in, param)
		return handled, true, err
	}
	return false, false, nil
}

// lookupParameter finds a ParameterObject in the current RouteOptions by name and location (in).
func (c *DefaultRouteContext) lookupParameter(name, in string) *openapi.ParameterObject {
	if c.options == nil || c.options.Parameters == nil {
		return nil
	}
	// Prefer a precomputed per-route index when available
	if c.options.ParamIndex != nil {
		if p, ok := c.options.ParamIndex[strings.ToLower(in+":"+name)]; ok {
			return p
		}
		return nil
	}
	// Fallback: build index once per request
	if c.paramIndex == nil {
		c.paramIndex = BuildParamIndex(c.options.Parameters)
	}
	return c.paramIndex[strings.ToLower(in+":"+name)]
}

// parseByExample attempts to parse a single string value into the type suggested by the ParameterObject's Example or Schema.
// These helpers are implemented in internal/binder.

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

// parseDeepKey returns the root name and a path slice for dotted or bracket keys.
// Examples:
//
//	"user.name" -> ("user", ["name"])
//	"user[address][city]" -> ("user", ["address","city"])
func parseDeepKey(key string) (string, []string) {
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		return parts[0], parts[1:]
	}
	// bracket notation: user[name] or user[address][city]
	if idx := strings.Index(key, "["); idx != -1 {
		root := key[:idx]
		rest := key[idx:]
		path := []string{}
		var b strings.Builder
		inBracket := false
		for _, r := range rest {
			switch r {
			case '[':
				inBracket = true
				b.Reset()
			case ']':
				inBracket = false
				path = append(path, b.String())
			default:
				if inBracket {
					b.WriteRune(r)
				}
			}
		}
		return root, path
	}
	return key, nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

// parseValueBySchema attempts to coerce raw string values into a typed value
// guided by the provided Schema (which may be nil). Returns error on parse failure.
// parseValueBySchema is implemented in binding_convert.go

// setNestedMap sets staging[root][path[0]][path[1]]... = value, creating maps as needed
func setNestedMap(staging map[string]any, root string, path []string, value any) {
	// ensure root map exists
	rootMap, _ := staging[root].(map[string]any)
	if rootMap == nil {
		rootMap = map[string]any{}
	}

	node := rootMap
	// walk/create intermediate maps
	for i := 0; i < len(path)-1; i++ {
		key := path[i]
		next, _ := node[key].(map[string]any)
		if next == nil {
			next = map[string]any{}
			node[key] = next
		}
		node = next
	}

	// set final value
	if len(path) > 0 {
		node[path[len(path)-1]] = value
	}
	staging[root] = rootMap
}
