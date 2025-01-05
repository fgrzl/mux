package mux

import "strings"

// AuthorizationOptions allows configuration of the authorization check
type AuthorizationOptions struct {
	Roles            []string
	Scopes           []string
	Permissions      []string
	CheckRoles       func(user ClaimsPrincipal, roles []string) bool
	CheckScopes      func(user ClaimsPrincipal, scopes []string) bool
	CheckPermissions func(user ClaimsPrincipal, permissions []string) bool
}

func (rtr *Router) UseAuthorization(options *AuthorizationOptions) {
	rtr.middleware = append(rtr.middleware, &authorizationMiddleware{options: options})
}

// authorizationMiddleware holds the options for authorization checking
type authorizationMiddleware struct {
	options *AuthorizationOptions
}

// Invoke is called for every request, checks the authorization and proceeds accordingly
func (m *authorizationMiddleware) Invoke(c *RouteContext, next HandlerFunc) {
	// Check if the user has the necessary roles
	if !m.checkRoles(c) {
		// User is not authorized, respond with 403 Forbidden
		c.Forbidden("You do not have the necessary permissions to access this resource.")
		return
	}

	// Check if the user has the necessary roles
	if !m.checkScopes(c) {
		// User is not authorized, respond with 403 Forbidden
		c.Forbidden("You do not have the necessary permissions to access this resource.")
		return
	}

	// Check if the user has the necessary roles
	if !m.checkPermission(c) {
		// User is not authorized, respond with 403 Forbidden
		c.Forbidden("You do not have the necessary permissions to access this resource.")
		return
	}

	// If authorized, proceed to the next handler
	next(c)
}

// CheckAuthorization checks if the user has the necessary roles to access the resource
func (m *authorizationMiddleware) checkRoles(c *RouteContext) bool {
	validRoles := c.Options.Roles
	if m.options.CheckRoles != nil {
		return m.options.CheckRoles(c.User, validRoles)
	}

	if validRoles == nil || len(validRoles) == 0 {
		// no Roles to check
		return true
	}

	userRoles := c.User.Roles()
	if len(userRoles) == 0 {
		// No user Roles, deny access
		return false
	}

	for _, requiredRole := range validRoles {
		for _, userRole := range userRoles {
			if requiredRole == userRole {
				return true
			}
		}
	}

	// No matching Roles
	return false
}

func (m *authorizationMiddleware) checkScopes(c *RouteContext) bool {
	validScopes := c.Options.Scopes

	if m.options.CheckScopes != nil {
		return m.options.CheckRoles(c.User, validScopes)
	}

	if validScopes == nil || len(validScopes) == 0 {
		// no scopes to check
		return true
	}

	userScopes := c.User.Scopes()
	if len(userScopes) == 0 {
		// No user Scopes, deny access
		return false
	}

	for _, requiredScope := range validScopes {
		for _, userScope := range userScopes {
			if requiredScope == userScope {
				return true
			}
		}
	}

	// No matching Scopes
	return false
}

func (m *authorizationMiddleware) checkPermission(c *RouteContext) bool {
	var permission []string
	if m.options.Permissions != nil {
		permission = append(permission, m.options.Permissions...)
	}

	if c.Options.Permissions != nil {
		permission = append(permission, c.Options.Permissions...)
	}
	if permission == nil || len(permission) == 0 {
		// no permissions to check
		return true
	}

	permissions := interpolatePermissions(c.Params, m.options.Permissions, c.Options.Permissions)
	return m.options.CheckPermissions(c.User, permissions)
}

// interpolatePermissions interpolates placeholders in permissions and ensures uniqueness.
func interpolatePermissions(replacements map[string]string, permissions ...[]string) []string {
	uniqueMap := make(map[string]struct{})
	var result []string

	// Iterate over slices of permissions
	for _, slice := range permissions {
		for _, item := range slice {
			val := interpolatePermission(replacements, item) // Interpolate the permission string
			if _, exists := uniqueMap[val]; !exists {
				uniqueMap[val] = struct{}{} // Mark as seen
				result = append(result, val)
			}
		}
	}
	return result
}

// interpolatePermission replaces placeholders in a permission string using a replacements map.
func interpolatePermission(replacements map[string]string, permission string) string {
	var result strings.Builder
	var placeholderStart int
	inPlaceholder := false

	// Iterate over the string and manually detect placeholders
	for i, ch := range permission {
		if ch == '{' {
			// Mark the start of a placeholder
			inPlaceholder = true
			placeholderStart = i + 1
		} else if ch == '}' && inPlaceholder {
			// End of a placeholder
			inPlaceholder = false
			placeholder := permission[placeholderStart:i] // Extract the placeholder

			// Look for the case-insensitive match in the map
			replaced := placeholder // Default to the placeholder itself if not replaced
			for key, value := range replacements {
				if strings.EqualFold(key, placeholder) {
					replaced = value
					break
				}
			}
			result.WriteString(replaced) // Append the replacement or the placeholder
		} else if !inPlaceholder {
			// Append normal characters outside of placeholders
			result.WriteRune(ch)
		}
	}

	return result.String()
}
