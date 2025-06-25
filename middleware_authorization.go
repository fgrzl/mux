package mux

import (
	"strings"

	"github.com/fgrzl/claims"
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

func (rtr *Router) UseAuthorization(opts ...AuthZOption) {
	options := &AuthorizationOptions{}
	for _, opt := range opts {
		opt(options)
	}
	rtr.middleware = append(rtr.middleware, &authorizationMiddleware{options: options})
}

type authorizationMiddleware struct {
	options *AuthorizationOptions
}

func (m *authorizationMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
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

func (m *authorizationMiddleware) checkRoles(c *RouteContext) bool {
	valid := c.Options.Roles
	if m.options.CheckRoles != nil {
		return m.options.CheckRoles(c.User, valid)
	}
	if len(valid) == 0 {
		return true
	}
	user := c.User.Roles()
	for _, r := range valid {
		for _, u := range user {
			if r == u {
				return true
			}
		}
	}
	return false
}

func (m *authorizationMiddleware) checkScopes(c *RouteContext) bool {
	valid := c.Options.Scopes
	if m.options.CheckScopes != nil {
		return m.options.CheckScopes(c.User, valid)
	}
	if len(valid) == 0 {
		return true
	}
	user := c.User.Scopes()
	for _, s := range valid {
		for _, u := range user {
			if s == u {
				return true
			}
		}
	}
	return false
}

func (m *authorizationMiddleware) checkPermission(c *RouteContext) bool {
	var merged []string
	merged = append(merged, m.options.Permissions...)
	merged = append(merged, c.Options.Permissions...)
	if len(merged) == 0 {
		return true
	}
	perms := interpolatePermissions(c.Params, m.options.Permissions, c.Options.Permissions)
	return m.options.CheckPermissions(c.User, perms)
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
