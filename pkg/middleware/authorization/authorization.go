package authorization

import (
	"strings"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
)

// ---- Functional Options ----

type AuthZOption func(*AuthorizationOptions)

func WithRoles(roles ...string) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.Roles = append(o.Roles, roles...)
	}
}

func WithScopes(scopes ...string) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.Scopes = append(o.Scopes, scopes...)
	}
}

func WithPermissions(perms ...string) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.Permissions = append(o.Permissions, perms...)
	}
}

func WithRoleChecker(fn func(claims.Principal, []string) bool) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.CheckRoles = fn
	}
}

func WithScopeChecker(fn func(claims.Principal, []string) bool) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.CheckScopes = fn
	}
}

func WithPermissionChecker(fn func(claims.Principal, []string) bool) AuthZOption {
	return func(o *AuthorizationOptions) {
		o.CheckPermissions = fn
	}
}

// ---- Authorization ----

type AuthorizationOptions struct {
	Roles            []string
	Scopes           []string
	Permissions      []string
	CheckRoles       func(claims.Principal, []string) bool
	CheckScopes      func(claims.Principal, []string) bool
	CheckPermissions func(claims.Principal, []string) bool
}

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
		c.Forbidden("You do not have the necessary permissions to access this resource.")
		return
	}
	if !m.checkScopes(c) {
		c.Forbidden("You do not have the necessary permissions to access this resource.")
		return
	}
	if !m.checkPermission(c) {
		c.Forbidden("You do not have the necessary permissions to access this resource.")
		return
	}
	next(c)
}

func (m *authorizationMiddleware) checkRoles(c routing.RouteContext) bool {
	opts := c.Options()
	var valid []string
	if opts != nil {
		valid = opts.Roles
	}
	// If middleware options are nil, there are no global role requirements or custom checker.
	if m.options == nil {
		if len(valid) == 0 {
			return true
		}
		if c.User() == nil {
			return false
		}
		return matchAny(valid, c.User().Roles())
	}
	if m.options.CheckRoles != nil {
		return m.options.CheckRoles(c.User(), valid)
	}
	if len(valid) == 0 {
		return true
	}
	if c.User() == nil {
		return false
	}
	return matchAny(valid, c.User().Roles())
}

func (m *authorizationMiddleware) checkScopes(c routing.RouteContext) bool {
	opts := c.Options()
	var valid []string
	if opts != nil {
		valid = opts.Scopes
	}
	if m.options == nil {
		if len(valid) == 0 {
			return true
		}
		if c.User() == nil {
			return false
		}
		return matchAny(valid, c.User().Scopes())
	}
	if m.options.CheckScopes != nil {
		return m.options.CheckScopes(c.User(), valid)
	}
	if len(valid) == 0 {
		return true
	}
	if c.User() == nil {
		return false
	}
	return matchAny(valid, c.User().Scopes())
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
	var sources [][]string
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

	perms := interpolatePermissions(c.Params(), sources...)

	if m.options != nil && m.options.CheckPermissions != nil {
		return m.options.CheckPermissions(c.User(), perms)
	}
	// If permissions are required but there's no checker, deny by default.
	return false
}

// ---- Helpers ----

func interpolatePermissions(replacements map[string]string, permissions ...[]string) []string {
	uniqueMap := make(map[string]struct{})
	var result []string
	for _, slice := range permissions {
		for _, item := range slice {
			val := interpolatePermission(replacements, item)
			if _, exists := uniqueMap[val]; !exists {
				uniqueMap[val] = struct{}{}
				result = append(result, val)
			}
		}
	}
	return result
}

func interpolatePermission(replacements map[string]string, permission string) string {
	var result strings.Builder
	var start int
	inPlaceholder := false

	for i, ch := range permission {
		if ch == '{' {
			inPlaceholder = true
			start = i + 1
		} else if ch == '}' && inPlaceholder {
			inPlaceholder = false
			placeholder := permission[start:i]
			replaced := placeholder
			for k, v := range replacements {
				if strings.EqualFold(k, placeholder) {
					replaced = v
					break
				}
			}
			result.WriteString(replaced)
		} else if !inPlaceholder {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
