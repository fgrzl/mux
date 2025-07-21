package mux

import (
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// ---- RouteGroup ----

// RouteGroup represents a group of routes with shared configuration and defaults.
type RouteGroup struct {
	prefix       string
	authProvider AuthProvider
	registry     *RouteRegistry

	// Group-level defaults:
	defaultParams      []*ParameterObject
	defaultRoles       []string
	defaultScopes      []string
	defaultPermissions []string
	defaultTags        []string
	defaultSummary     string
	defaultDescription string
	defaultSecurity    []*SecurityRequirement
	defaultAllowAnon   bool
	defaultDeprecated  bool
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
	schema, err := quickSchema(reflect.TypeOf(example))
	if err != nil {
		panic(err)
	}
	rg.defaultParams = append(rg.defaultParams, &ParameterObject{
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
func (rg *RouteGroup) WithSecurity(sec *SecurityRequirement) *RouteGroup {
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

// ---- Route Registration (Apply Defaults) ----

// registerRoute registers a route with all group-level defaults applied.
func (rg *RouteGroup) registerRoute(method, pattern string, handler HandlerFunc) *RouteBuilder {
	pattern = normalizeRoute(pattern, rg.prefix)

	op := Operation{
		Summary:     rg.defaultSummary,
		Description: rg.defaultDescription,
		Deprecated:  rg.defaultDeprecated,
		Responses:   map[string]*ResponseObject{},
	}

	if len(rg.defaultTags) > 0 {
		op.Tags = append([]string{}, rg.defaultTags...)
	}
	if len(rg.defaultSecurity) > 0 {
		op.Security = append([]*SecurityRequirement{}, rg.defaultSecurity...)
	}
	if len(rg.defaultParams) > 0 {
		op.Parameters = append([]*ParameterObject{}, rg.defaultParams...)
	}

	options := &RouteOptions{
		Method:         method,
		Pattern:        pattern,
		Handler:        handler,
		AllowAnonymous: rg.defaultAllowAnon,
		Roles:          append([]string{}, rg.defaultRoles...),
		Scopes:         append([]string{}, rg.defaultScopes...),
		Permissions:    append([]string{}, rg.defaultPermissions...),
		RateLimit:      0,
		RateInterval:   0,
		AuthProvider:   nil,
		Operation:      op,
	}

	rg.registry.Register(pattern, method, options)
	return &RouteBuilder{Options: options}
}

// ---- Route Methods ----

// GET registers a GET route with group defaults.
func (rg *RouteGroup) GET(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodGet, pattern, handler)
}

// POST registers a POST route with group defaults.
func (rg *RouteGroup) POST(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodPost, pattern, handler)
}

// PUT registers a PUT route with group defaults.
func (rg *RouteGroup) PUT(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodPut, pattern, handler)
}

// DELETE registers a DELETE route with group defaults.
func (rg *RouteGroup) DELETE(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodDelete, pattern, handler)
}

// HEAD registers a HEAD route with group defaults.
func (rg *RouteGroup) HEAD(pattern string, handler HandlerFunc) *RouteBuilder {
	return rg.registerRoute(http.MethodHead, pattern, handler)
}

// Healthz registers a /healthz endpoint that always returns ready.
func (rg *RouteGroup) Healthz() *RouteBuilder {
	return rg.HealthzWithReady(func() bool { return true })
}

// HealthzWithReady registers a /healthz endpoint with a custom readiness check.
func (rg *RouteGroup) HealthzWithReady(isReady func() bool) *RouteBuilder {
	return rg.registerRoute(http.MethodGet, "/healthz", func(c *RouteContext) {
		c.Response.Header().Set("Content-Type", "text/plain")
		if isReady() {
			c.Response.WriteHeader(http.StatusOK)
			c.Response.Write([]byte("ok"))
			return
		}
		c.Response.WriteHeader(http.StatusServiceUnavailable)
		c.Response.Write([]byte("not ready"))
	}).AllowAnonymous()
}

// StaticFallback serves static files with a fallback for SPA routing, with directory safety checks.
func (rg *RouteGroup) StaticFallback(pattern, dir, fallback string) *RouteBuilder {
	prefix := strings.TrimSuffix(pattern, "**")
	prefix = strings.TrimRight(prefix, "/")

	handler := func(c *RouteContext) {
		requestPath := c.Request.URL.Path
		trimmed := strings.TrimPrefix(requestPath, prefix)
		trimmed = strings.TrimPrefix(trimmed, "/")
		fullPath := filepath.Join(dir, trimmed)
		absFullPath, err := filepath.Abs(fullPath)
		absDir, dirErr := filepath.Abs(dir)
		if err != nil || dirErr != nil || !strings.HasPrefix(absFullPath, absDir) {
			http.ServeFile(c.Response, c.Request, fallback)
			return
		}
		info, err := os.Stat(absFullPath)
		if err != nil || info.IsDir() {
			http.ServeFile(c.Response, c.Request, fallback)
			return
		}
		// Serve as static
		r := *c.Request
		r.URL.Path = "/" + trimmed
		http.FileServer(http.Dir(dir)).ServeHTTP(c.Response, &r)
	}
	return rg.registerRoute(http.MethodGet, pattern, handler)
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
