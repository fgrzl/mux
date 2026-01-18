package router

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fgrzl/mux/pkg/builder"
	openapi "github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/registry"
	"github.com/fgrzl/mux/pkg/routing"
)

// ---- RouteGroup ----

// RouteGroup represents a group of routes with shared configuration and defaults.
//
// IMPORTANT: Route registration (GET, POST, PUT, DELETE, etc.) must be done during
// application startup, before calling http.ListenAndServe() or starting any concurrent
// request handling. The route registry is not protected by a mutex for performance
// reasons, so registering routes while the router is handling requests will cause
// data races and undefined behavior.
//
// Safe usage pattern:
//
//	rtr := router.NewRouter()
//	api := rtr.NewRouteGroup("/api/v1")
//	api.GET("/users", listUsers)    // OK: during startup
//	api.POST("/users", createUser)  // OK: during startup
//	http.ListenAndServe(":8080", rtr) // Now serving requests
type RouteGroup struct {
	prefix        string
	routeRegistry *registry.RouteRegistry

	// Group-level defaults:
	defaultParams      []*openapi.ParameterObject
	defaultRoles       []string
	defaultScopes      []string
	defaultPermissions []string
	defaultTags        []string
	defaultSummary     string
	defaultDescription string
	defaultSecurity    []*openapi.SecurityRequirement
	defaultAllowAnon   bool
	defaultDeprecated  bool
}

func (rg *RouteGroup) RouteRegistry() *registry.RouteRegistry {
	return rg.routeRegistry
}

// ---- Chainable Group Setters ----

// WithPathParam adds a required path parameter with an example value.
func (rg *RouteGroup) WithPathParam(name string, example any) *RouteGroup {
	return rg.WithParam(name, "path", example, true)
}

// WithQueryParam adds an optional query parameter with an example value.
func (rg *RouteGroup) WithQueryParam(name string, example any) *RouteGroup {
	return rg.WithParam(name, "query", example, false)
}

// WithRequiredQueryParam adds a required query parameter with an example value.
func (rg *RouteGroup) WithRequiredQueryParam(name string, example any) *RouteGroup {
	return rg.WithParam(name, "query", example, true)
}

// WithHeaderParam adds a header parameter with an example value.
func (rg *RouteGroup) WithHeaderParam(name string, example any, required bool) *RouteGroup {
	return rg.WithParam(name, "header", example, required)
}

// WithCookieParam adds a cookie parameter with an example value.
func (rg *RouteGroup) WithCookieParam(name string, example any, required bool) *RouteGroup {
	return rg.WithParam(name, "cookie", example, required)
}

// WithParam adds a parameter of any type/location to the group defaults.
func (rg *RouteGroup) WithParam(name, in string, example any, required bool) *RouteGroup {
	schema, err := builder.QuickSchema(reflect.TypeOf(example))
	if err != nil {
		panic(err)
	}
	rg.defaultParams = append(rg.defaultParams, &openapi.ParameterObject{
		Name:     name,
		In:       in,
		Required: required,
		Schema:   schema,
	})
	return rg
}

// RequireRoles adds required roles to the group defaults.
func (rg *RouteGroup) RequireRoles(roles ...string) *RouteGroup {
	rg.defaultRoles = append(rg.defaultRoles, roles...)
	return rg
}

// RequireScopes adds required scopes to the group defaults.
func (rg *RouteGroup) RequireScopes(scopes ...string) *RouteGroup {
	rg.defaultScopes = append(rg.defaultScopes, scopes...)
	return rg
}

// RequirePermission adds required permissions to the group defaults.
func (rg *RouteGroup) RequirePermission(perms ...string) *RouteGroup {
	rg.defaultPermissions = append(rg.defaultPermissions, perms...)
	return rg
}

// WithTags adds tags to the group defaults.
func (rg *RouteGroup) WithTags(tags ...string) *RouteGroup {
	rg.defaultTags = append(rg.defaultTags, tags...)
	return rg
}

// WithSummary sets the summary for the group defaults.
func (rg *RouteGroup) WithSummary(s string) *RouteGroup {
	rg.defaultSummary = s
	return rg
}

// WithDescription sets the description for the group defaults.
func (rg *RouteGroup) WithDescription(desc string) *RouteGroup {
	rg.defaultDescription = desc
	return rg
}

// WithSecurity adds a security requirement to the group defaults.
func (rg *RouteGroup) WithSecurity(sec *openapi.SecurityRequirement) *RouteGroup {
	rg.defaultSecurity = append(rg.defaultSecurity, sec)
	return rg
}

// AllowAnonymous allows anonymous access for the group.
func (rg *RouteGroup) AllowAnonymous() *RouteGroup {
	rg.defaultAllowAnon = true
	return rg
}

// Deprecated marks the group as deprecated.
func (rg *RouteGroup) Deprecated() *RouteGroup {
	rg.defaultDeprecated = true
	return rg
}

// ---- Nested Group Creation ----

// copyDefaults copies all default settings from source to target RouteGroup.
func (target *RouteGroup) copyDefaults(source *RouteGroup) {
	// For startup-only registration we can avoid copying slices and keep
	// references to reduce allocations. This assumes defaults won't be
	// mutated concurrently after group creation.
	target.defaultParams = source.defaultParams
	target.defaultRoles = source.defaultRoles
	target.defaultScopes = source.defaultScopes
	target.defaultPermissions = source.defaultPermissions
	target.defaultTags = source.defaultTags
	target.defaultSecurity = source.defaultSecurity
	target.defaultSummary = source.defaultSummary
	target.defaultDescription = source.defaultDescription
	target.defaultAllowAnon = source.defaultAllowAnon
	target.defaultDeprecated = source.defaultDeprecated
}

// newRouteGroupBase creates a new RouteGroup with basic initialization.
func newRouteGroupBase(prefix string, registry *registry.RouteRegistry) *RouteGroup {
	return &RouteGroup{
		prefix:        prefix,
		routeRegistry: registry,
	}
}

// NewRouteGroup creates a new RouteGroup with an extended prefix and inherited defaults.
// The new group inherits all defaults from the parent and uses the same registry and auth provider.
func (rg *RouteGroup) NewRouteGroup(prefix string) *RouteGroup {
	// Use the existing normalizeRoute function to properly join the prefixes
	extendedPrefix := normalizeRoute(prefix, rg.prefix)

	// Create new group with basic initialization
	newGroup := newRouteGroupBase(extendedPrefix, rg.routeRegistry)

	// Copy all defaults from parent
	newGroup.copyDefaults(rg)

	return newGroup
}

// ---- Route Registration (Apply Defaults) ----

// registerRoute registers a route with all group-level defaults applied.
func (rg *RouteGroup) registerRoute(method, pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	pattern = normalizeRoute(pattern, rg.prefix)

	op := openapi.Operation{
		Summary:     rg.defaultSummary,
		Description: rg.defaultDescription,
		Deprecated:  rg.defaultDeprecated,
		Responses:   map[string]*openapi.ResponseObject{},
	}

	// Use references to group defaults to avoid allocations during registration.
	if len(rg.defaultTags) > 0 {
		op.Tags = rg.defaultTags
	}
	if len(rg.defaultSecurity) > 0 {
		op.Security = rg.defaultSecurity
	}
	if len(rg.defaultParams) > 0 {
		op.Parameters = rg.defaultParams
	}

	options := &routing.RouteOptions{
		Method:         method,
		Pattern:        pattern,
		Handler:        handler,
		AllowAnonymous: rg.defaultAllowAnon,
		Roles:          rg.defaultRoles,
		Scopes:         rg.defaultScopes,
		Permissions:    rg.defaultPermissions,
		RateLimit:      0,
		RateInterval:   0,
		Operation:      op,
	}
	// Build an initial ParamIndex from group defaults if any are present
	if len(op.Parameters) > 0 {
		options.ParamIndex = routing.BuildParamIndex(op.Parameters)
	}

	rg.routeRegistry.Register(pattern, method, options)
	return &builder.RouteBuilder{Options: options}
}

// ---- Route Methods ----

// GET registers a GET route with group defaults.
func (rg *RouteGroup) GET(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodGet, pattern, handler)
}

// POST registers a POST route with group defaults.
func (rg *RouteGroup) POST(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodPost, pattern, handler)
}

// PUT registers a PUT route with group defaults.
func (rg *RouteGroup) PUT(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodPut, pattern, handler)
}

// DELETE registers a DELETE route with group defaults.
func (rg *RouteGroup) DELETE(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodDelete, pattern, handler)
}

// HEAD registers a HEAD route with group defaults.
func (rg *RouteGroup) HEAD(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodHead, pattern, handler)
}

// Healthz registers a /healthz endpoint that always returns ready.
func (rg *RouteGroup) Healthz() *builder.RouteBuilder {
	return rg.HealthzWithReady(func(c routing.RouteContext) bool { return true })
}

// HealthzWithReady registers a /healthz endpoint with a custom readiness check.
func (rg *RouteGroup) HealthzWithReady(isReady func(c routing.RouteContext) bool) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodGet, "/healthz", func(c routing.RouteContext) {
		if isReady(c) {
			c.Plain(http.StatusOK, []byte("ok"))
			return
		}
		c.Plain(http.StatusServiceUnavailable, []byte("not ready"))
	}).AllowAnonymous()
}

// Livez registers a /livez endpoint that always returns live.
func (rg *RouteGroup) Livez() *builder.RouteBuilder {
	return rg.LivezWithCheck(func(c routing.RouteContext) bool { return true })
}

// LivezWithCheck registers a /livez endpoint with a custom liveness check.
func (rg *RouteGroup) LivezWithCheck(isLive func(c routing.RouteContext) bool) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodGet, "/livez", func(c routing.RouteContext) {
		if isLive(c) {
			c.Plain(http.StatusOK, []byte("ok"))
			return
		}
		c.Plain(http.StatusServiceUnavailable, []byte("not live"))
	}).AllowAnonymous()
}

// Readyz registers a /readyz endpoint that always returns ready.
func (rg *RouteGroup) Readyz() *builder.RouteBuilder {
	return rg.ReadyzWithCheck(func(c routing.RouteContext) bool { return true })
}

// ReadyzWithCheck registers a /readyz endpoint with a custom readiness check.
func (rg *RouteGroup) ReadyzWithCheck(isReady func(c routing.RouteContext) bool) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodGet, "/readyz", func(c routing.RouteContext) {
		if isReady(c) {
			c.Plain(http.StatusOK, []byte("ok"))
			return
		}
		c.Plain(http.StatusServiceUnavailable, []byte("not ready"))
	}).AllowAnonymous()
}

// Startupz registers a /startupz endpoint that always returns started.
func (rg *RouteGroup) Startupz() *builder.RouteBuilder {
	return rg.StartupzWithCheck(func(c routing.RouteContext) bool { return true })
}

// StartupzWithCheck registers a /startupz endpoint with a custom startup check.
func (rg *RouteGroup) StartupzWithCheck(hasStarted func(c routing.RouteContext) bool) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodGet, "/startupz", func(c routing.RouteContext) {
		if hasStarted(c) {
			c.Plain(http.StatusOK, []byte("ok"))
			return
		}
		c.Plain(http.StatusServiceUnavailable, []byte("not started"))
	}).AllowAnonymous()
}

// StaticFallback serves static files with a fallback for SPA routing, with directory safety checks.
func (rg *RouteGroup) StaticFallback(pattern, dir, fallback string) *builder.RouteBuilder {
	// Determine the URL prefix (pattern without the trailing "**"), and normalize
	prefix := strings.TrimSuffix(pattern, "**")
	prefix = strings.TrimRight(prefix, "/")

	// Resolve absolute directory and fallback file for robust file serving
	absDir, _ := filepath.Abs(dir)
	// Always serve from absolute directory to avoid CWD surprises
	fs := http.FileServer(http.Dir(absDir))
	// Resolve fallback to a file within absDir; assume fallback file name lives under dir
	fallbackAbs := filepath.Join(absDir, filepath.Base(fallback))

	// Catch-all handler: serve static file if it exists within dir, otherwise serve fallback
	handler := func(c routing.RouteContext) {
		requestPath := c.Request().URL.Path
		trimmed := strings.TrimPrefix(requestPath, prefix)
		trimmed = strings.TrimPrefix(trimmed, "/")
		absFullPath := filepath.Join(absDir, trimmed)
		// Safety: ensure the resolved path is within the static directory
		if !strings.HasPrefix(absFullPath, absDir) {
			http.ServeFile(c.Response(), c.Request(), fallbackAbs)
			return
		}
		info, err := os.Stat(absFullPath)
		if err != nil || info.IsDir() {
			http.ServeFile(c.Response(), c.Request(), fallbackAbs)
			return
		}
		// Serve as static (adjust URL.Path for the FileServer)
		r := *c.Request()
		r.URL.Path = "/" + trimmed
		fs.ServeHTTP(c.Response(), &r)
	}

	// Also register a base path route that directly serves the fallback file so
	// requests to the group root (e.g., "/" or "/app/") return the SPA entry.
	// We intentionally allow anonymous access for SPA assets by default.
	baseHandler := func(c routing.RouteContext) {
		http.ServeFile(c.Response(), c.Request(), fallbackAbs)
	}
	// Empty route allows normalizeRoute to map to group base (e.g., prefix or "/").
	rg.registerRoute(http.MethodGet, "", baseHandler).AllowAnonymous()

	// Register the catch-all route and return its builder for further chaining.
	return rg.registerRoute(http.MethodGet, pattern, handler).AllowAnonymous()
}

// ---- Utilities ----

// normalizeRoute joins and cleans up the route and prefix.
func normalizeRoute(route, prefix string) string {
	prefix = strings.TrimRight(prefix, "/")
	route = strings.TrimLeft(route, "/")
	if !strings.HasPrefix(route, prefix) {
		route = prefix + "/" + route
	}
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	route = strings.ReplaceAll(route, "//", "/")
	return route
}
