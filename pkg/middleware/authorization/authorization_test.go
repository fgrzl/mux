package authorization

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/stretchr/testify/assert"
)

// mockPrincipalForAuth implements claims.Principal for authorization testing
type mockPrincipalForAuth struct {
	roles       []string
	scopes      []string
	permissions []string
}

func (m *mockPrincipalForAuth) Subject() string                      { return "test-user" }
func (m *mockPrincipalForAuth) Issuer() string                       { return "test" }
func (m *mockPrincipalForAuth) Audience() []string                   { return []string{"test"} }
func (m *mockPrincipalForAuth) ExpirationTime() int64                { return 0 }
func (m *mockPrincipalForAuth) NotBefore() int64                     { return 0 }
func (m *mockPrincipalForAuth) IssuedAt() int64                      { return 0 }
func (m *mockPrincipalForAuth) JWTI() string                         { return "test-jwt" }
func (m *mockPrincipalForAuth) Scopes() []string                     { return m.scopes }
func (m *mockPrincipalForAuth) Roles() []string                      { return m.roles }
func (m *mockPrincipalForAuth) Email() string                        { return "test@example.com" }
func (m *mockPrincipalForAuth) Username() string                     { return "testuser" }
func (m *mockPrincipalForAuth) CustomClaim(name string) claims.Claim { return nil }
func (m *mockPrincipalForAuth) CustomClaimValue(name string) string  { return "" }
func (m *mockPrincipalForAuth) Claims() *claims.ClaimSet             { return nil }

func TestShouldCreateAuthZOptionsWithRoles(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	roles := []string{"admin", "user"}

	// Act
	opt := WithRoles(roles...)
	opt(options)

	// Assert
	assert.Equal(t, roles, options.Roles)
}

func TestShouldCreateAuthZOptionsWithScopes(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	scopes := []string{"read", "write"}

	// Act
	opt := WithScopes(scopes...)
	opt(options)

	// Assert
	assert.Equal(t, scopes, options.Scopes)
}

func TestShouldCreateAuthZOptionsWithPermissions(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	permissions := []string{"user:read", "user:write"}

	// Act
	opt := WithPermissions(permissions...)
	opt(options)

	// Assert
	assert.Equal(t, permissions, options.Permissions)
}

func TestShouldCreateAuthZOptionsWithRoleChecker(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	checker := func(claims.Principal, []string) bool { return true }

	// Act
	opt := WithRoleChecker(checker)
	opt(options)

	// Assert
	assert.NotNil(t, options.CheckRoles)
}

func TestShouldCreateAuthZOptionsWithScopeChecker(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	checker := func(claims.Principal, []string) bool { return true }

	// Act
	opt := WithScopeChecker(checker)
	opt(options)

	// Assert
	assert.NotNil(t, options.CheckScopes)
}

func TestShouldCreateAuthZOptionsWithPermissionChecker(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	checker := func(claims.Principal, []string) bool { return true }

	// Act
	opt := WithPermissionChecker(checker)
	opt(options)

	// Assert
	assert.NotNil(t, options.CheckPermissions)
}

func TestShouldAddAuthorizationMiddlewareToRouter(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	// Act
	UseAuthorization(rtr, WithRoles("admin"))

	// There's no exported way to inspect middleware slice; instead ensure ServeHTTP runs without panic
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()
	rtr.ServeHTTP(rec, req)
}

func TestShouldAllowAccessWhenUserHasRequiredRole(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{roles: []string{"admin", "user"}})
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{"admin"}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestShouldDenyAccessWhenUserLacksRequiredRole(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{roles: []string{"user"}})
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{"admin"}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, 403, rec.Code) // Forbidden
}

func TestShouldAllowAccessWhenUserHasRequiredScope(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{scopes: []string{"api:read", "api:write"}})
	ctx.SetOptions(&routing.RouteOptions{Scopes: []string{"api:read"}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestShouldDenyAccessWhenUserLacksRequiredScope(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{scopes: []string{"api:write"}})
	ctx.SetOptions(&routing.RouteOptions{Scopes: []string{"api:admin"}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, 403, rec.Code)
}

func TestShouldUseCustomRoleChecker(t *testing.T) {
	// Arrange
	customCheckerCalled := false
	customChecker := func(principal claims.Principal, roles []string) bool {
		customCheckerCalled = true
		return true // Always allow
	}

	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{
			CheckRoles: customChecker,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{roles: []string{"user"}})
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{"admin"}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, customCheckerCalled)
	assert.True(t, nextCalled) // Should pass because custom checker returns true
}

func TestShouldUseCustomScopeChecker(t *testing.T) {
	// Arrange
	customCheckerCalled := false
	customChecker := func(principal claims.Principal, scopes []string) bool {
		customCheckerCalled = true
		return false // Always deny
	}

	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{
			CheckScopes: customChecker,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{scopes: []string{"api:read"}})
	ctx.SetOptions(&routing.RouteOptions{Scopes: []string{"api:read"}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, customCheckerCalled)
	assert.False(t, nextCalled) // Should be denied because custom checker returns false
}

func TestShouldAllowAccessWhenNoRolesRequired(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{})
	ctx.SetOptions(&routing.RouteOptions{}) // No roles required

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestShouldAllowAccessWhenNoScopesRequired(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{})
	ctx.SetOptions(&routing.RouteOptions{}) // No scopes required

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}

func TestInterpolatePermissionShouldReplaceParameters(t *testing.T) {
	// Arrange
	replacements := map[string]string{
		"userId":   "123",
		"tenantId": "tenant-456",
	}
	permission := "user:{userId}:read:tenant:{tenantId}"

	// Act
	result := interpolatePermission(replacements, permission)

	// Assert
	assert.Equal(t, "user:123:read:tenant:tenant-456", result)
}

func TestInterpolatePermissionShouldBeCaseInsensitive(t *testing.T) {
	// Arrange
	replacements := map[string]string{
		"UserId": "123",
	}
	permission := "user:{userid}:read"

	// Act
	result := interpolatePermission(replacements, permission)

	// Assert
	assert.Equal(t, "user:123:read", result)
}

func TestInterpolatePermissionShouldHandleNonexistentPlaceholders(t *testing.T) {
	// Arrange
	replacements := map[string]string{
		"userId": "123",
	}
	permission := "user:{nonexistent}:read"

	// Act
	result := interpolatePermission(replacements, permission)

	// Assert
	assert.Equal(t, "user:nonexistent:read", result) // Should keep the placeholder name
}

func TestInterpolatePermissionsShouldRemoveDuplicates(t *testing.T) {
	// Arrange
	replacements := map[string]string{
		"id": "123",
	}
	permissions1 := []string{"user:{id}:read", "user:123:write"}
	permissions2 := []string{"user:123:read", "admin:{id}:delete"}

	// Act
	result := interpolatePermissions(replacements, permissions1, permissions2)

	// Assert
	expected := []string{"user:123:read", "user:123:write", "admin:123:delete"}
	assert.ElementsMatch(t, expected, result)
}

func TestShouldHandlePermissionCheckingWithCustomChecker(t *testing.T) {
	// Arrange
	customPermissionChecker := func(principal claims.Principal, permissions []string) bool {
		return len(permissions) > 0 && permissions[0] == "allowed:permission"
	}

	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{
			CheckPermissions: customPermissionChecker,
			Permissions:      []string{"allowed:permission"},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{})
	ctx.SetOptions(&routing.RouteOptions{})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled)
}
