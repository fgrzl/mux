package authorization

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// forbiddenMessage is a shared message used when access is denied.
const forbiddenMessage = "You do not have the necessary permissions to access this resource."

// Well-known scope constants used by tests and examples.
const (
	ScopeAPIRead  = "api:read"
	ScopeAPIWrite = "api:write"
	ScopeAPIAdmin = "api:admin"
)

// ---- Functional Options ----

// AuthZOption represents a functional option for configuring AuthorizationOptions.
type AuthZOption func(*AuthorizationOptions)

// WithRoles appends one or more roles to the middleware's required roles.
func WithRoles(roles ...string) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.Roles = append(o.Roles, roles...)
	}
}

// WithScopes appends one or more scopes to the middleware's required scopes.
func WithScopes(scopes ...string) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.Scopes = append(o.Scopes, scopes...)
	}
}

// WithPermissions appends one or more permissions to the middleware's required permissions.
func WithPermissions(perms ...string) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.Permissions = append(o.Permissions, perms...)
	}
}

// WithRoleChecker overrides the default role checking behavior with a custom function.
func WithRoleChecker(fn func(claims.Principal, []string) bool) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.CheckRoles = fn
	}
}

// WithScopeChecker overrides the default scope checking behavior with a custom function.
func WithScopeChecker(fn func(claims.Principal, []string) bool) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.CheckScopes = fn
	}
}

// WithPermissionChecker overrides the default permission checking behavior with a custom function.
func WithPermissionChecker(fn func(claims.Principal, []string) bool) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.CheckPermissions = fn
	}
}

// ---- Authorization ----

// AuthorizationOptions defines configuration for the authorization middleware,
// including global roles, scopes, permissions, and custom checkers.
type AuthorizationOptions struct {
	Roles            []string
	Scopes           []string
	Permissions      []string
	CheckRoles       func(claims.Principal, []string) bool
	CheckScopes      func(claims.Principal, []string) bool
	CheckPermissions func(claims.Principal, []string) bool
}

// UseAuthorization registers the authorization middleware on the provided router
// with the supplied options. Options are collected and applied once at registration time.
func UseAuthorization(rtr *router.Router, opts ...AuthZOption) {
	options := &AuthorizationOptions{}
	for _, opt := range opts {
		opt(options)
	}
	// Use the public API to register middleware instead of touching unexported fields.
	rtr.Use(&authorizationMiddleware{options: options})
}

type authorizationMiddleware struct {
	options *AuthorizationOptions
}

func (m *authorizationMiddleware) Invoke(c routing.RouteContext, next router.HandlerFunc) {
	// Be defensive: if middleware was constructed without options, treat as no-op config
	if !m.checkRoles(c) {
		c.Forbidden(forbiddenMessage)
		return
	}
	if !m.checkScopes(c) {
		c.Forbidden(forbiddenMessage)
		return
	}
	if !m.checkPermission(c) {
		c.Forbidden(forbiddenMessage)
		return
	}
	next(c)
}

func (m *authorizationMiddleware) checkRoles(c routing.RouteContext) bool {
	var checker func(claims.Principal, []string) bool = func(p claims.Principal, vals []string) bool {
		if m.options != nil && m.options.CheckRoles != nil {
			return m.options.CheckRoles(p, vals)
		}
		return matchAny(vals, p.Roles())
	}
	if m.options != nil && !m.checkStringList(m.options.Roles, c.User(), checker) {
		return false
	}
	if opts := c.Options(); opts != nil {
		return m.checkStringList(opts.Roles, c.User(), checker)
	}
	return true
}

func (m *authorizationMiddleware) checkScopes(c routing.RouteContext) bool {
	var checker func(claims.Principal, []string) bool = func(p claims.Principal, vals []string) bool {
		if m.options != nil && m.options.CheckScopes != nil {
			return m.options.CheckScopes(p, vals)
		}
		return matchAny(vals, p.Scopes())
	}
	if m.options != nil && !m.checkStringList(m.options.Scopes, c.User(), checker) {
		return false
	}
	if opts := c.Options(); opts != nil {
		return m.checkStringList(opts.Scopes, c.User(), checker)
	}
	return true
}

// checkStringList centralizes the logic used by checkRoles and checkScopes.
//   - required: values required by the route
//   - user: the principal (may be nil)
//   - opts: middleware options (may be nil)
//   - checker: custom checker to evaluate when options are present; it should
//     return true to allow access.
func (m *authorizationMiddleware) checkStringList(required []string, user claims.Principal, checker func(claims.Principal, []string) bool) bool {
	if len(required) == 0 {
		return true
	}
	if user == nil {
		return false
	}
	if checker == nil {
		return false
	}
	return checker(user, required)
}

// matchAny returns true if any element in required exists in userVals. If required is empty
// it returns true. userVals may be nil. The implementation builds a set from the smaller
// slice to reduce allocations and work when slices differ in size.
func matchAny(required []string, userVals []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(userVals) == 0 {
		return false
	}
	// Build set from the smaller slice to reduce allocations.
	if len(userVals) < len(required) {
		set := make(map[string]struct{}, len(userVals))
		for _, v := range userVals {
			set[v] = struct{}{}
		}
		for _, r := range required {
			if _, ok := set[r]; ok {
				return true
			}
		}
		return false
	}
	set := make(map[string]struct{}, len(required))
	for _, r := range required {
		set[r] = struct{}{}
	}
	for _, v := range userVals {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

func (m *authorizationMiddleware) checkPermission(c routing.RouteContext) bool {
	// Gather permission sources in a stable order: first middleware-level, then route-level.
	// Preallocate for the common case (0-2 sources) to avoid small allocations.
	sources := make([][]string, 0, 2)
	if m.options != nil {
		sources = append(sources, m.options.Permissions)
	}
	if opts := c.Options(); opts != nil {
		sources = append(sources, opts.Permissions)
	}

	// If there are no configured permissions globally or for the route, allow.
	total := 0
	for _, s := range sources {
		total += len(s)
	}
	if total == 0 {
		return true
	}

	// Prefer a non-allocating interpolation path using the Params slice.
	// Fall back to the map-based helper only if there is no params slice.
	if ps := c.ParamsSlice(); ps != nil {
		perms := interpolatePermissions(ps, sources...)
		if m.options != nil && m.options.CheckPermissions != nil {
			return m.options.CheckPermissions(c.User(), perms)
		}
		// Permissions required but no checker configured - log warning and deny
		slog.WarnContext(c, "permissions required but no CheckPermissions function configured", "permissions", perms)
		return false
	}

	// Legacy: no params slice available, use map-based interpolation (allocates).
	// There are no params to copy since ps is nil; call the map-based helper with
	// an empty map to keep behavior consistent.
	perms := interpolatePermissions(nil, sources...)

	if m.options != nil && m.options.CheckPermissions != nil {
		return m.options.CheckPermissions(c.User(), perms)
	}
	// Permissions required but no checker configured - log warning and deny
	slog.WarnContext(c, "permissions required but no CheckPermissions function configured", "permissions", perms)
	return false
}

// ---- Helpers ----

// Deprecated map-based interpolation removed in favor of slice-based helpers.
// The slice-based versions are defined below and are used throughout the codebase.

// interpolatePermissionsFromSlice behaves like interpolatePermissions but looks
// up replacements from the provided Params slice without allocating a map.
func interpolatePermissions(ps *routing.Params, permissions ...[]string) []string {
	var result []string
	for _, slice := range permissions {
		for _, item := range slice {
			val := interpolatePermission(ps, item)
			// small N dedupe via linear scan avoids allocating a map
			found := false
			for _, r := range result {
				if r == val {
					found = true
					break
				}
			}
			if !found {
				result = append(result, val)
			}
		}
	}
	return result
}

// interpolatePermissionFromSlice performs placeholder interpolation using
// a Params slice for lookups. It scans the slice for matching keys.
func interpolatePermission(ps *routing.Params, permission string) string {
	// Use a pooled buffer to avoid intermediate allocations from creating new
	// buffers/builders for each interpolation. Final string allocation still occurs
	// when calling `buf.String()` but intermediate buffer reuse reduces pressure.
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	start := -1
	for i := 0; i < len(permission); i++ {
		ch := permission[i]
		if ch == '{' {
			start = i + 1
			continue
		}
		if ch == '}' && start != -1 {
			placeholder := permission[start:i]
			replaced := placeholder
			if ps != nil {
				for j := 0; j < ps.Len(); j++ {
					p := (*ps)[j]
					if strings.EqualFold(p.Key, placeholder) {
						replaced = p.Value
						break
					}
				}
			}
			buf.WriteString(replaced)
			start = -1
			continue
		}
		if start == -1 {
			buf.WriteByte(ch)
		}
	}
	return buf.String()
}

var bufferPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// note: map-based interpolation removed
