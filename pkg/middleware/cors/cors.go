package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// ---- Functional Options ----

// CORSOption is a function type for configuring CORS options.
type CORSOption func(*CORSOptions)

// WithAllowedOrigins sets the origins that are allowed for CORS requests.
// Use ["*"] or leave empty to allow any origin.
func WithAllowedOrigins(origins ...string) CORSOption {
	return func(o *CORSOptions) {
		o.AllowedOrigins = origins
	}
}

// WithAllowedMethods sets the HTTP methods allowed for CORS requests.
// If not set, defaults to GET, POST, PUT, DELETE, OPTIONS.
func WithAllowedMethods(methods ...string) CORSOption {
	return func(o *CORSOptions) {
		o.AllowedMethods = methods
	}
}

// WithAllowedHeaders sets the headers allowed for cross-origin requests.
// If not set, incoming requested headers will be reflected.
func WithAllowedHeaders(headers ...string) CORSOption {
	return func(o *CORSOptions) {
		o.AllowedHeaders = headers
	}
}

// WithExposeHeaders sets the headers that are safe to expose to the browser.
func WithExposeHeaders(headers ...string) CORSOption {
	return func(o *CORSOptions) {
		o.ExposeHeaders = headers
	}
}

// WithCredentials enables Access-Control-Allow-Credentials.
func WithCredentials(allow bool) CORSOption {
	return func(o *CORSOptions) {
		o.AllowCredentials = allow
	}
}

// WithMaxAge sets the value (in seconds) for Access-Control-Max-Age on preflight responses.
func WithMaxAge(seconds int) CORSOption {
	return func(o *CORSOptions) {
		o.MaxAge = seconds
	}
}

// ---- Options ----

// CORSOptions configure the CORS middleware behavior.
type CORSOptions struct {
	// AllowedOrigins lists origins that are allowed. Use ["*"] or leave empty to allow any origin.
	AllowedOrigins []string
	// AllowedMethods lists HTTP methods to allow in CORS. If empty, defaults to common methods.
	AllowedMethods []string
	// AllowedHeaders lists headers allowed for cross-origin requests. If empty, incoming requested headers will be reflected.
	AllowedHeaders []string
	// ExposeHeaders lists headers that are safe to expose to the browser.
	ExposeHeaders []string
	// AllowCredentials enables Access-Control-Allow-Credentials.
	AllowCredentials bool
	// MaxAge is the value (in seconds) for Access-Control-Max-Age on preflight responses.
	MaxAge int
}

// Options is deprecated. Use CORSOptions instead.
// Kept for backward compatibility.
type Options = CORSOptions

// ---- Middleware ----

// ---- Middleware ----

// corsMiddleware implements the middleware.
type corsMiddleware struct {
	opts CORSOptions
	// normalized comma-joined values
	methods string
	headers string
	expose  string
	// precomputed origin checks
	allowedMap   map[string]struct{}
	permissive   bool // true when AllowedOrigins is empty -> allow any origin
	hasWildcard  bool // true when AllowedOrigins contains '*'
	wildcardSole bool // true when AllowedOrigins is exactly ['*']
}

// defaultMethods returns a sensible default set if none provided.
func defaultMethods() []string {
	return []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions}
}

// newCORSMiddleware builds middleware and prepares header values.
func newCORSMiddleware(opts CORSOptions) *corsMiddleware {
	cm := &corsMiddleware{opts: opts}
	methods := opts.AllowedMethods
	if len(methods) == 0 {
		methods = defaultMethods()
	}
	cm.methods = strings.Join(methods, ", ")

	if len(opts.AllowedHeaders) > 0 {
		cm.headers = strings.Join(opts.AllowedHeaders, ", ")
	}
	if len(opts.ExposeHeaders) > 0 {
		cm.expose = strings.Join(opts.ExposeHeaders, ", ")
	}

	// prepare origin lookup structures
	if len(opts.AllowedOrigins) == 0 {
		// permissive by default (back-compat)
		cm.permissive = true
	} else {
		// Only create the map when we actually need to store explicit origins.
		// Wildcard entries are handled separately.
		cm.allowedMap = make(map[string]struct{}, len(opts.AllowedOrigins))
		for _, a := range opts.AllowedOrigins {
			if a == "*" {
				cm.hasWildcard = true
				continue
			}
			// normalize to lower-case for faster comparison
			cm.allowedMap[strings.ToLower(a)] = struct{}{}
		}
		if cm.hasWildcard && len(opts.AllowedOrigins) == 1 {
			cm.wildcardSole = true
		}
	}
	return cm
}

// isOriginAllowed returns true if origin matches allowed list or wildcard present.
func (m *corsMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	if m.permissive || m.hasWildcard {
		return true
	}
	if m.allowedMap == nil {
		return false
	}
	_, ok := m.allowedMap[strings.ToLower(origin)]
	return ok
}

// Invoke implements the Middleware interface.
func (m *corsMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	req := c.Request()
	res := c.Response()
	origin := req.Header.Get(common.HeaderOrigin)

	if origin == "" {
		// Not a CORS request
		next(c)
		return
	}

	allowOrigin := m.determineAllowedOrigin(origin)
	if allowOrigin != "" {
		m.setResponseHeaders(res, allowOrigin)
	}

	// Handle preflight
	if req.Method == http.MethodOptions {
		m.handlePreflight(c, next, allowOrigin, req, res)
		return
	}

	// For simple requests, proceed to next handler which will write the response body.
	next(c)
}

// determineAllowedOrigin checks if the origin is allowed and returns the appropriate value.
func (m *corsMiddleware) determineAllowedOrigin(origin string) string {
	if !m.isOriginAllowed(origin) {
		return ""
	}

	// Use wildcard if configured and no credentials
	if len(m.opts.AllowedOrigins) == 1 && m.opts.AllowedOrigins[0] == "*" && !m.opts.AllowCredentials {
		return "*"
	}

	return origin
}

// setResponseHeaders sets the CORS response headers.
func (m *corsMiddleware) setResponseHeaders(res http.ResponseWriter, allowOrigin string) {
	res.Header().Set(common.HeaderAccessControlAllowOrigin, allowOrigin)

	if m.opts.AllowCredentials {
		res.Header().Set(common.HeaderAccessControlAllowCredentials, "true")
	}

	if m.expose != "" {
		res.Header().Set(common.HeaderAccessControlExposeHeaders, m.expose)
	}
}

// handlePreflight handles CORS preflight requests.
func (m *corsMiddleware) handlePreflight(c routing.RouteContext, next router.HandlerFunc, allowOrigin string, req *http.Request, res http.ResponseWriter) {
	// Only respond to preflight if origin is allowed
	if allowOrigin == "" {
		// Not allowed; continue to next which may return 403 or 200
		next(c)
		return
	}

	// Requested method
	reqMethod := req.Header.Get(common.HeaderAccessControlRequestMethod)
	if reqMethod == "" {
		// malformed preflight; continue
		next(c)
		return
	}

	res.Header().Set(common.HeaderAccessControlAllowMethods, m.methods)

	// Determine headers to allow
	m.setAllowedHeaders(req, res)

	if m.opts.MaxAge > 0 {
		res.Header().Set(common.HeaderAccessControlMaxAge, strconv.Itoa(m.opts.MaxAge))
	}

	// 204 No Content for preflight
	res.WriteHeader(http.StatusNoContent)
}

// setAllowedHeaders sets the Access-Control-Allow-Headers header.
func (m *corsMiddleware) setAllowedHeaders(req *http.Request, res http.ResponseWriter) {
	if m.headers != "" {
		res.Header().Set(common.HeaderAccessControlAllowHeaders, m.headers)
		return
	}

	// Reflect requested headers if none configured
	if requestedHeaders := req.Header.Get(common.HeaderAccessControlRequestHeaders); requestedHeaders != "" {
		res.Header().Set(common.HeaderAccessControlAllowHeaders, requestedHeaders)
	}
}

// UseCORS registers the CORS middleware with the router using functional options.
func UseCORS(rtr *router.Router, opts ...CORSOption) {
	options := &CORSOptions{}
	for _, opt := range opts {
		opt(options)
	}
	rtr.Use(newCORSMiddleware(*options))
}

// newCORS is deprecated. Use newCORSMiddleware instead.
// Kept for backward compatibility.
func newCORS(opts CORSOptions) *corsMiddleware {
	return newCORSMiddleware(opts)
}
