package router

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/fgrzl/mux/pkg/binder"
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
	validation    *routing.ValidationState

	// Group-level defaults:
	defaultMiddleware  []Middleware
	defaultServices    map[routing.ServiceKey]any
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

// Safe switches this RouteGroup into non-panicking validation mode and returns it.
// Prefer Configure in new application code when you want a single error return
// for group-level setup.
func (rg *RouteGroup) Safe() *RouteGroup {
	rg.validation = rg.validationState().WithPanicOnError(false)
	return rg
}

// Configure runs startup configuration for this RouteGroup with validation
// errors returned instead of panicking. Nested groups and route builders
// created within the callback share the same validation sink.
func (rg *RouteGroup) Configure(configure func(*RouteGroup)) error {
	if configure == nil {
		return nil
	}

	original := rg.validationState()
	configured := original.WithPanicOnError(false)
	rg.validation = configured
	defer func() {
		rg.validation = original
	}()

	configure(rg)
	return configured.Err()
}

// Errors returns accumulated configuration errors for this RouteGroup tree.
func (rg *RouteGroup) Errors() []error {
	return rg.validationState().Errors()
}

// Err returns accumulated configuration errors for this RouteGroup tree.
func (rg *RouteGroup) Err() error {
	return rg.validationState().Err()
}

func (rg *RouteGroup) validationState() *routing.ValidationState {
	if rg.validation == nil {
		rg.validation = routing.NewValidationState()
	}
	return rg.validation
}

func (rg *RouteGroup) handleValidation(err error) *RouteGroup {
	rg.validationState().Handle(err)
	return rg
}

// Use registers middleware for this RouteGroup and all nested groups/routes
// created from it. Middleware is inherited by child groups and copied into
// each route at registration time.
func (rg *RouteGroup) Use(middleware ...Middleware) *RouteGroup {
	rg.defaultMiddleware = append(rg.defaultMiddleware, middleware...)
	return rg
}

// Services returns a fluent registry for configuring scoped services on this
// RouteGroup.
func (rg *RouteGroup) Services() *routing.ServiceRegistry {
	return routing.NewServiceRegistry(
		func(key routing.ServiceKey, svc any) {
			rg.WithService(key, svc)
		},
		func(key routing.ServiceKey) (any, bool) {
			if rg.defaultServices == nil {
				return nil, false
			}
			svc, ok := rg.defaultServices[key]
			return svc, ok
		},
	)
}

// WithService registers a service for this RouteGroup and all nested
// groups/routes created from it.
func (rg *RouteGroup) WithService(key routing.ServiceKey, svc any) *RouteGroup {
	if key == "" || svc == nil {
		return rg
	}
	if rg.defaultServices == nil {
		rg.defaultServices = make(map[routing.ServiceKey]any)
	}
	rg.defaultServices[key] = svc
	return rg
}

// ---- Chainable Group Setters ----

// WithPathParam adds a required path parameter to all routes in this group.
//
// Parameters:
//   - name: The parameter name as it appears in the route pattern (e.g., "id" for "/users/{id}").
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type. For instance,
//     pass uuid.Nil for UUID parameters, 0 for integers, or "" for strings.
//
// Path parameters are always marked as required in the OpenAPI spec.
func (rg *RouteGroup) WithPathParam(name, description string, example any) *RouteGroup {
	return rg.WithParam(name, "path", description, example, true)
}

// WithPathParamErr adds a required path parameter without panicking on validation failures.
func (rg *RouteGroup) WithPathParamErr(name, description string, example any) (*RouteGroup, error) {
	return rg.WithParamErr(name, "path", description, example, true)
}

// WithQueryParam adds an optional query parameter to all routes in this group.
//
// Parameters:
//   - name: The query parameter name (e.g., "limit" for "?limit=10").
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type. For instance,
//     pass 10 for integer parameters, true for booleans, or "" for strings.
//
// Query parameters added via this method are marked as optional in the OpenAPI spec.
func (rg *RouteGroup) WithQueryParam(name, description string, example any) *RouteGroup {
	return rg.WithParam(name, "query", description, example, false)
}

// WithQueryParamErr adds an optional query parameter without panicking on validation failures.
func (rg *RouteGroup) WithQueryParamErr(name, description string, example any) (*RouteGroup, error) {
	return rg.WithParamErr(name, "query", description, example, false)
}

// WithRequiredQueryParam adds a required query parameter to all routes in this group.
//
// Parameters:
//   - name: The query parameter name (e.g., "apiKey" for "?apiKey=xyz").
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type. For instance,
//     pass 10 for integer parameters, true for booleans, or "" for strings.
//
// Query parameters added via this method are marked as required in the OpenAPI spec.
func (rg *RouteGroup) WithRequiredQueryParam(name, description string, example any) *RouteGroup {
	return rg.WithParam(name, "query", description, example, true)
}

// WithRequiredQueryParamErr adds a required query parameter without panicking on validation failures.
func (rg *RouteGroup) WithRequiredQueryParamErr(name, description string, example any) (*RouteGroup, error) {
	return rg.WithParamErr(name, "query", description, example, true)
}

// WithHeaderParam adds a header parameter to all routes in this group.
//
// Parameters:
//   - name: The HTTP header name (e.g., "X-API-Version" or "Authorization").
//   - description: Human-readable explanation of the header for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type. For instance,
//     pass "v1" for string headers or 1 for integer headers.
//   - required: If true, the header is marked as required in the OpenAPI spec;
//     if false, it's marked as optional.
func (rg *RouteGroup) WithHeaderParam(name, description string, example any, required bool) *RouteGroup {
	return rg.WithParam(name, "header", description, example, required)
}

// WithHeaderParamErr adds a header parameter without panicking on validation failures.
func (rg *RouteGroup) WithHeaderParamErr(name, description string, example any, required bool) (*RouteGroup, error) {
	return rg.WithParamErr(name, "header", description, example, required)
}

// WithCookieParam adds a cookie parameter to all routes in this group.
//
// Parameters:
//   - name: The cookie name (e.g., "sessionId" or "csrf_token").
//   - description: Human-readable explanation of the cookie for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type. For instance,
//     pass "" for string cookies or 0 for integer cookies.
//   - required: If true, the cookie is marked as required in the OpenAPI spec;
//     if false, it's marked as optional.
func (rg *RouteGroup) WithCookieParam(name, description string, example any, required bool) *RouteGroup {
	return rg.WithParam(name, "cookie", description, example, required)
}

// WithCookieParamErr adds a cookie parameter without panicking on validation failures.
func (rg *RouteGroup) WithCookieParamErr(name, description string, example any, required bool) (*RouteGroup, error) {
	return rg.WithParamErr(name, "cookie", description, example, required)
}

// WithParam adds a parameter of any type/location to all routes in this group.
// This is a low-level method; prefer using WithPathParam, WithQueryParam, etc. for better type safety.
//
// Parameters:
//   - name: The parameter name (e.g., "id", "limit", "X-API-Key").
//   - in: The parameter location. Must be one of: "path", "query", "header", or "cookie".
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type. The type of this value
//     determines the OpenAPI schema (e.g., int → integer, string → string, uuid.UUID → string with format uuid).
//   - required: If true, the parameter is marked as required in the OpenAPI spec;
//     if false, it's marked as optional. Note: path parameters are always required regardless of this value.
func (rg *RouteGroup) WithParam(name, in, description string, example any, required bool) *RouteGroup {
	if _, err := rg.WithParamErr(name, in, description, example, required); err != nil {
		return rg.handleValidation(err)
	}
	return rg
}

// WithParamErr adds a parameter without panicking on validation failures.
func (rg *RouteGroup) WithParamErr(name, in, description string, example any, required bool) (*RouteGroup, error) {
	if name == "" || in == "" {
		return rg, fmt.Errorf("parameter name and 'in' cannot be empty")
	}
	if !isValidGroupParameterIn(in) {
		return rg, fmt.Errorf("invalid parameter 'in': %q", in)
	}

	schema, err := builder.QuickSchema(reflect.TypeOf(example))
	if err != nil {
		return rg, err
	}
	conv := binder.MakeConverter(reflect.TypeOf(example), schema)
	rg.defaultParams = append(rg.defaultParams, openapi.CloneParameterObject(&openapi.ParameterObject{
		Name:        name,
		In:          in,
		Description: description,
		Required:    required || in == "path",
		Schema:      schema,
		Example:     example,
		Converter:   conv,
	}))
	return rg, nil
}

func isValidGroupParameterIn(in string) bool {
	switch in {
	case "query", "header", "path", "cookie":
		return true
	default:
		return false
	}
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
	rg.defaultSecurity = append(rg.defaultSecurity, openapi.CloneSecurityRequirement(sec))
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
	target.defaultMiddleware = slices.Clone(source.defaultMiddleware)
	target.validation = source.validationState().Clone()
	if len(source.defaultServices) > 0 {
		target.defaultServices = cloneGroupServices(source.defaultServices)
	}
	if len(source.defaultParams) > 0 {
		target.defaultParams = make([]*openapi.ParameterObject, len(source.defaultParams))
		for index, param := range source.defaultParams {
			target.defaultParams[index] = openapi.CloneParameterObject(param)
		}
	}
	target.defaultRoles = slices.Clone(source.defaultRoles)
	target.defaultScopes = slices.Clone(source.defaultScopes)
	target.defaultPermissions = slices.Clone(source.defaultPermissions)
	target.defaultTags = slices.Clone(source.defaultTags)
	if len(source.defaultSecurity) > 0 {
		target.defaultSecurity = make([]*openapi.SecurityRequirement, len(source.defaultSecurity))
		for index, sec := range source.defaultSecurity {
			target.defaultSecurity[index] = openapi.CloneSecurityRequirement(sec)
		}
	}
	target.defaultSummary = source.defaultSummary
	target.defaultDescription = source.defaultDescription
	target.defaultAllowAnon = source.defaultAllowAnon
	target.defaultDeprecated = source.defaultDeprecated
}

func cloneGroupServices(services map[routing.ServiceKey]any) map[routing.ServiceKey]any {
	if len(services) == 0 {
		return nil
	}
	cloned := make(map[routing.ServiceKey]any, len(services))
	for key, svc := range services {
		if key == "" || svc == nil {
			continue
		}
		cloned[key] = svc
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

func routeBuilderValidationState(rb *builder.RouteBuilder) *routing.ValidationState {
	if rb.Validation == nil {
		rb.Validation = routing.NewValidationState()
	}
	return rb.Validation
}

func cloneDetachedRouteOptions(source *routing.RouteOptions) *routing.RouteOptions {
	if source == nil {
		return nil
	}

	operation := openapi.CloneOperation(&source.Operation)
	if operation == nil {
		operation = &openapi.Operation{Responses: map[string]*openapi.ResponseObject{}}
	}

	cloned := &routing.RouteOptions{
		Method:         strings.ToUpper(source.Method),
		Pattern:        source.Pattern,
		Handler:        source.Handler,
		AllowAnonymous: source.AllowAnonymous,
		Roles:          slices.Clone(source.Roles),
		Scopes:         slices.Clone(source.Scopes),
		Permissions:    slices.Clone(source.Permissions),
		RateLimit:      source.RateLimit,
		RateInterval:   source.RateInterval,
		Operation:      *operation,
	}
	cloned.SetMiddleware(slices.Clone(source.Middleware))
	cloned.SetServices(source.Services)
	cloned.ParamIndex = routing.BuildParamIndex(cloned.Parameters)
	if cloned.Responses == nil {
		cloned.Responses = map[string]*openapi.ResponseObject{}
	}
	return cloned
}

func mergeRouteOptions(target, source *routing.RouteOptions) {
	if target == nil || source == nil {
		return
	}

	target.AllowAnonymous = target.AllowAnonymous || source.AllowAnonymous
	target.Roles = append(target.Roles, slices.Clone(source.Roles)...)
	target.Scopes = append(target.Scopes, slices.Clone(source.Scopes)...)
	target.Permissions = append(target.Permissions, slices.Clone(source.Permissions)...)
	if source.RateLimit > 0 {
		target.RateLimit = source.RateLimit
	}
	if source.RateInterval > 0 {
		target.RateInterval = source.RateInterval
	}
	target.AppendMiddleware(slices.Clone(source.Middleware)...)
	for key, service := range source.Services {
		target.SetService(key, service)
	}

	operation := openapi.CloneOperation(&source.Operation)
	if operation == nil {
		return
	}

	if operation.OperationID != "" {
		target.OperationID = operation.OperationID
	}
	if operation.Summary != "" {
		target.Summary = operation.Summary
	}
	if operation.Description != "" {
		target.Description = operation.Description
	}
	if len(operation.Tags) > 0 {
		target.Tags = append(target.Tags, operation.Tags...)
	}
	if operation.ExternalDocs != nil {
		target.ExternalDocs = operation.ExternalDocs
	}
	if len(operation.Parameters) > 0 {
		target.Parameters = append(target.Parameters, operation.Parameters...)
	}
	if operation.RequestBody != nil {
		target.RequestBody = operation.RequestBody
	}
	if len(operation.Responses) > 0 {
		if target.Responses == nil {
			target.Responses = make(map[string]*openapi.ResponseObject, len(operation.Responses))
		}
		for code, response := range operation.Responses {
			target.Responses[code] = response
		}
	}
	if len(operation.Callbacks) > 0 {
		if target.Callbacks == nil {
			target.Callbacks = make(map[string]*openapi.PathItem, len(operation.Callbacks))
		}
		for key, callback := range operation.Callbacks {
			target.Callbacks[key] = callback
		}
	}
	target.Deprecated = target.Deprecated || operation.Deprecated
	if len(operation.Security) > 0 {
		target.Security = append(target.Security, operation.Security...)
	}
	if len(operation.Servers) > 0 {
		target.Servers = operation.Servers
	}
	if len(operation.Extensions) > 0 {
		if target.Extensions == nil {
			target.Extensions = make(map[string]any, len(operation.Extensions))
		}
		for key, value := range operation.Extensions {
			target.Extensions[key] = value
		}
	}
	target.ParamIndex = routing.BuildParamIndex(target.Parameters)
}

func (rg *RouteGroup) newRouteOptions(method, pattern string, handler routing.HandlerFunc) *routing.RouteOptions {
	method = strings.ToUpper(method)
	pattern = normalizeRoute(pattern, rg.prefix)

	op := openapi.Operation{
		Summary:     rg.defaultSummary,
		Description: rg.defaultDescription,
		Deprecated:  rg.defaultDeprecated,
		Responses:   map[string]*openapi.ResponseObject{},
	}

	if len(rg.defaultTags) > 0 {
		op.Tags = slices.Clone(rg.defaultTags)
	}
	if len(rg.defaultSecurity) > 0 {
		op.Security = make([]*openapi.SecurityRequirement, len(rg.defaultSecurity))
		for index, sec := range rg.defaultSecurity {
			op.Security[index] = openapi.CloneSecurityRequirement(sec)
		}
	}
	if len(rg.defaultParams) > 0 {
		op.Parameters = make([]*openapi.ParameterObject, len(rg.defaultParams))
		for index, param := range rg.defaultParams {
			op.Parameters[index] = openapi.CloneParameterObject(param)
		}
	}

	options := &routing.RouteOptions{
		Method:         method,
		Pattern:        pattern,
		Handler:        handler,
		AllowAnonymous: rg.defaultAllowAnon,
		Roles:          slices.Clone(rg.defaultRoles),
		Scopes:         slices.Clone(rg.defaultScopes),
		Permissions:    slices.Clone(rg.defaultPermissions),
		RateLimit:      0,
		RateInterval:   0,
		Operation:      op,
	}
	if len(op.Parameters) > 0 {
		options.ParamIndex = routing.BuildParamIndex(op.Parameters)
	}
	options.SetMiddleware(slices.Clone(rg.defaultMiddleware))
	options.SetServices(cloneGroupServices(rg.defaultServices))
	return options
}

func (rg *RouteGroup) registerBuiltRoute(rb *builder.RouteBuilder) *builder.RouteBuilder {
	if rb == nil {
		rb = &builder.RouteBuilder{}
	}
	validation := routeBuilderValidationState(rb)
	if rb.Options == nil {
		rb.Options = &routing.RouteOptions{Operation: openapi.Operation{Responses: map[string]*openapi.ResponseObject{}}}
	}
	rb.Options.Method = strings.ToUpper(rb.Options.Method)
	if rb.Options.Responses == nil {
		rb.Options.Responses = map[string]*openapi.ResponseObject{}
	}
	rb.Options.ParamIndex = routing.BuildParamIndex(rb.Options.Parameters)
	if validation.Err() != nil {
		return rb
	}
	if rb.Options.Method == "" || rb.Options.Pattern == "" {
		validation.Handle(fmt.Errorf("route method and pattern cannot be empty"))
		return rb
	}
	if rg.routeRegistry.HasRoute(rb.Options.Pattern, rb.Options.Method) {
		validation.Handle(fmt.Errorf("route %s %s is already registered", rb.Options.Method, rb.Options.Pattern))
		return rb
	}

	registered := false
	rb.Validation = validation.WithErrorHook(func(error) {
		if !registered {
			return
		}
		rg.routeRegistry.Unregister(rb.Options.Pattern, rb.Options.Method)
		registered = false
	})

	rg.routeRegistry.Register(rb.Options.Pattern, rb.Options.Method, rb.Options)
	registered = true
	return rb
}

// newRouteGroupBase creates a new RouteGroup with basic initialization.
func newRouteGroupBase(prefix string, registry *registry.RouteRegistry) *RouteGroup {
	return &RouteGroup{
		prefix:        prefix,
		routeRegistry: registry,
		validation:    routing.NewValidationState(),
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
	rb := &builder.RouteBuilder{
		Options:    rg.newRouteOptions(method, pattern, handler),
		Validation: rg.validationState().Clone(),
	}
	return rg.registerBuiltRoute(rb)
}

// HandleRoute attaches a preconfigured detached RouteBuilder to this group.
// Prefer GET, POST, and related helpers for normal application code.
func (rg *RouteGroup) HandleRoute(route *builder.RouteBuilder, handler routing.HandlerFunc) *builder.RouteBuilder {
	rb := &builder.RouteBuilder{Validation: rg.validationState().Clone()}
	if route == nil {
		rb.Options = &routing.RouteOptions{Operation: openapi.Operation{Responses: map[string]*openapi.ResponseObject{}}}
		routeBuilderValidationState(rb).Handle(fmt.Errorf("route builder cannot be nil"))
		return rb
	}

	source := cloneDetachedRouteOptions(route.Options)
	if source == nil {
		rb.Options = &routing.RouteOptions{Operation: openapi.Operation{Responses: map[string]*openapi.ResponseObject{}}}
		routeBuilderValidationState(rb).Handle(fmt.Errorf("route options cannot be nil"))
		return rb
	}
	if handler != nil {
		source.Handler = handler
	}

	rb.Options = rg.newRouteOptions(source.Method, source.Pattern, source.Handler)
	mergeRouteOptions(rb.Options, source)
	for _, err := range route.Errors() {
		routeBuilderValidationState(rb).Handle(err)
	}
	if rb.Err() != nil {
		return rb
	}
	return rg.registerBuiltRoute(rb)
}

// ---- Route Methods ----

// Handle registers a route using a standard-library http.Handler.
func (rg *RouteGroup) Handle(method, pattern string, handler http.Handler) *builder.RouteBuilder {
	return rg.registerRoute(method, pattern, routing.HTTPHandler(handler))
}

// HandleFunc registers a route using a standard-library http.HandlerFunc.
func (rg *RouteGroup) HandleFunc(method, pattern string, handler http.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(method, pattern, routing.HTTPHandlerFunc(handler))
}

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

// PATCH registers a PATCH route with group defaults.
func (rg *RouteGroup) PATCH(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodPatch, pattern, handler)
}

// DELETE registers a DELETE route with group defaults.
func (rg *RouteGroup) DELETE(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodDelete, pattern, handler)
}

// HEAD registers a HEAD route with group defaults.
func (rg *RouteGroup) HEAD(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodHead, pattern, handler)
}

// OPTIONS registers an OPTIONS route with group defaults.
func (rg *RouteGroup) OPTIONS(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodOptions, pattern, handler)
}

// TRACE registers a TRACE route with group defaults.
func (rg *RouteGroup) TRACE(pattern string, handler routing.HandlerFunc) *builder.RouteBuilder {
	return rg.registerRoute(http.MethodTrace, pattern, handler)
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
	// Resolve fallback to a file within absDir while preserving nested paths.
	fallbackAbs := resolveStaticFallbackPath(absDir, dir, fallback)

	// Catch-all handler: serve static file if it exists within dir, otherwise serve fallback
	handler := func(c routing.RouteContext) {
		requestPath := c.Request().URL.Path
		trimmed := strings.TrimPrefix(requestPath, prefix)
		trimmed = strings.TrimPrefix(trimmed, "/")
		absFullPath := filepath.Join(absDir, trimmed)
		// Safety: ensure the resolved path is within the static directory
		if !isPathWithinDir(absDir, absFullPath) {
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
	normalizedPrefix := normalizeRoutePrefix(prefix)
	normalizedRoute := normalizeRoutePath(route)

	if normalizedRoute == "" {
		if normalizedPrefix == "" {
			return "/"
		}
		return normalizedPrefix + "/"
	}
	if normalizedPrefix == "" {
		return normalizedRoute
	}
	if hasRoutePrefix(normalizedRoute, normalizedPrefix) {
		return normalizedRoute
	}

	joined := normalizedPrefix + "/" + strings.TrimLeft(normalizedRoute, "/")
	return strings.ReplaceAll(joined, "//", "/")
}

func normalizeRoutePrefix(prefix string) string {
	prefix = normalizeRoutePath(prefix)
	return strings.TrimRight(prefix, "/")
}

func normalizeRoutePath(route string) string {
	if route == "" {
		return ""
	}
	keepTrailingSlash := strings.HasSuffix(route, "/")
	route = strings.TrimLeft(route, "/")
	if route == "" {
		return "/"
	}
	route = "/" + strings.ReplaceAll(route, "//", "/")
	if keepTrailingSlash && !strings.HasSuffix(route, "/") {
		route += "/"
	}
	return route
}

func hasRoutePrefix(route, prefix string) bool {
	route = strings.TrimRight(route, "/")
	prefix = strings.TrimRight(prefix, "/")
	return route == prefix || strings.HasPrefix(route, prefix+"/")
}

func resolveStaticFallbackPath(absDir, dir, fallback string) string {
	fallbackPath := fallback
	if !filepath.IsAbs(fallbackPath) {
		fallbackPath = filepath.Clean(filepath.FromSlash(fallbackPath))
		dirClean := filepath.Clean(filepath.FromSlash(dir))
		prefixes := []string{dirClean, filepath.Base(dirClean)}
		for _, prefix := range prefixes {
			if prefix == "." || prefix == string(filepath.Separator) || prefix == "" {
				continue
			}
			if fallbackPath == prefix {
				fallbackPath = "."
				break
			}
			trimPrefix := prefix + string(filepath.Separator)
			if strings.HasPrefix(fallbackPath, trimPrefix) {
				fallbackPath = strings.TrimPrefix(fallbackPath, trimPrefix)
				break
			}
		}
		fallbackPath = filepath.Join(absDir, fallbackPath)
	}

	fallbackPath = filepath.Clean(fallbackPath)
	if !isPathWithinDir(absDir, fallbackPath) {
		return filepath.Join(absDir, filepath.Base(fallbackPath))
	}
	return fallbackPath
}

func isPathWithinDir(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
