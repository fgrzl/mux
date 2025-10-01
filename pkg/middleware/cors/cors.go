package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// Options configure the CORS middleware behavior.
type Options struct {
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

// corsMiddleware implements the middleware.
type corsMiddleware struct {
	opts Options
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

// newCORS builds middleware and prepares header values.
func newCORS(opts Options) *corsMiddleware {
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

	allowOrigin := ""
	if m.isOriginAllowed(origin) {
		if len(m.opts.AllowedOrigins) == 1 && m.opts.AllowedOrigins[0] == "*" && !m.opts.AllowCredentials {
			allowOrigin = "*"
		} else {
			allowOrigin = origin
		}
	}

	if allowOrigin != "" {
		res.Header().Set(common.HeaderAccessControlAllowOrigin, allowOrigin)
		if m.opts.AllowCredentials {
			res.Header().Set(common.HeaderAccessControlAllowCredentials, "true")
		}
		if m.expose != "" {
			res.Header().Set(common.HeaderAccessControlExposeHeaders, m.expose)
		}
	}

	// Handle preflight
	if req.Method == http.MethodOptions {
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

		// determine headers to allow
		if m.headers != "" {
			res.Header().Set(common.HeaderAccessControlAllowHeaders, m.headers)
		} else if h := req.Header.Get(common.HeaderAccessControlRequestHeaders); h != "" {
			// reflect requested headers
			res.Header().Set(common.HeaderAccessControlAllowHeaders, h)
		}

		if m.opts.MaxAge > 0 {
			res.Header().Set(common.HeaderAccessControlMaxAge, strconv.Itoa(m.opts.MaxAge))
		}

		// 204 No Content for preflight
		res.WriteHeader(http.StatusNoContent)
		return
	}

	// For simple requests, proceed to next handler which will write the response body.
	next(c)
}

// UseCORS registers the CORS middleware with the router.
func UseCORS(rtr *router.Router, opts Options) {
	rtr.Use(newCORS(opts))
}
