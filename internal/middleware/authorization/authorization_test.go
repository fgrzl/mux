package authorization

import (
	"context"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux/internal/router"
	"github.com/fgrzl/mux/internal/routing"
	"github.com/stretchr/testify/assert"
)

const (
	// Common permissions used across tests
	PermUser123Read = "user:123:read"

	// Test identity values for mock principal were removed; not used in the
	// current test suite.

	// Common request paths / URLs used in tests
	testURL  = "http://example.com/test"
	pathTest = "/test"

	// Roles and scopes
	roleAdmin  = "admin"
	roleUser   = "user"
	scopeRead  = "read"
	scopeWrite = "write"

	// Permission templates and examples
	permUserRead           = "user:read"
	permUserWrite          = "user:write"
	permUserIDReadTemplate = "user:{id}:read"
	user123Write           = "user:123:write"
	adminIDDeleteTemplate  = "admin:{id}:delete"
	admin123Delete         = "admin:123:delete"
	permAllowedPermission  = "allowed:permission"

	// Placeholder-heavy permission used in interpolation tests
	permUserTenantTemplate = "user:{userId}:read:tenant:{tenantId}"
	permUseridTemplate     = "user:{userid}:read"
	permNonexistent        = "user:{nonexistent}:read"
)

// mockPrincipalForAuth implements claims.Principal for authorization testing
type mockPrincipalForAuth struct {
	roles  []string
	scopes []string
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
	roles := []string{roleAdmin, roleUser}

	// Act
	opt := WithRoles(roles...)
	opt(options)

	// Assert
	assert.Equal(t, roles, options.Roles)
}

func TestShouldCreateAuthZOptionsWithScopes(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	scopes := []string{scopeRead, scopeWrite}

	// Act
	opt := WithScopes(scopes...)
	opt(options)

	// Assert
	assert.Equal(t, scopes, options.Scopes)
}

func TestShouldCreateAuthZOptionsWithPermissions(t *testing.T) {
	// Arrange
	options := &AuthorizationOptions{}
	permissions := []string{permUserRead, permUserWrite}

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
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/not-found", nil)
	rec := httptest.NewRecorder()
	rtr.ServeHTTP(rec, req)
}

func TestShouldPreserveMethodNotAllowedGivenAuthorizationMiddleware(t *testing.T) {
	// Arrange
	rtr := router.NewRouter()
	UseAuthorization(rtr, WithRoles(roleAdmin))
	rtr.GET(pathTest, func(c routing.RouteContext) {
		c.OK("success")
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, pathTest, nil)
	rec := httptest.NewRecorder()

	// Act
	rtr.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Contains(t, rec.Header().Get("Allow"), http.MethodGet)
}

func TestShouldAllowAccessWhenUserHasRequiredRole(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{roles: []string{roleAdmin, roleUser}})
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{roleAdmin}})

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

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{roles: []string{roleUser}})
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{roleAdmin}})

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

func TestShouldDenyAccessWhenMiddlewareRequiresRoleAndUserLacksIt(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{Roles: []string{roleAdmin}},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, pathTest, nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{roles: []string{roleUser}})
	ctx.SetOptions(&routing.RouteOptions{})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestShouldRequireBothMiddlewareAndRouteRoles(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{Roles: []string{roleAdmin}},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, pathTest, nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{roles: []string{roleAdmin}})
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{roleUser}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestShouldAllowAccessWhenUserHasRequiredScope(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, pathTest, nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{scopes: []string{ScopeAPIRead, ScopeAPIWrite}})
	ctx.SetOptions(&routing.RouteOptions{Scopes: []string{ScopeAPIRead}})

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

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, pathTest, nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{scopes: []string{ScopeAPIWrite}})
	ctx.SetOptions(&routing.RouteOptions{Scopes: []string{ScopeAPIAdmin}})

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

func TestShouldDenyAccessWhenMiddlewareRequiresScopeAndUserLacksIt(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{Scopes: []string{ScopeAPIAdmin}},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, pathTest, nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{scopes: []string{ScopeAPIRead}})
	ctx.SetOptions(&routing.RouteOptions{})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestShouldRequireBothMiddlewareAndRouteScopes(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{Scopes: []string{ScopeAPIAdmin}},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, pathTest, nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	ctx.SetUser(&mockPrincipalForAuth{scopes: []string{ScopeAPIAdmin}})
	ctx.SetOptions(&routing.RouteOptions{Scopes: []string{ScopeAPIRead}})

	nextCalled := false
	next := func(c routing.RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestShouldUseCustomPrincipalCheckers(t *testing.T) {
	cases := []struct {
		name          string
		useRoles      bool
		checkerReturn bool
		wantNext      bool
		routeRoles    []string
		routeScopes   []string
		newPrincipal  func() *mockPrincipalForAuth
	}{
		{
			name:          "roles allow via CheckRoles",
			useRoles:      true,
			checkerReturn: true,
			wantNext:      true,
			routeRoles:    []string{roleAdmin},
			newPrincipal: func() *mockPrincipalForAuth {
				return &mockPrincipalForAuth{roles: []string{roleUser}}
			},
		},
		{
			name:          "scopes deny via CheckScopes",
			useRoles:      false,
			checkerReturn: false,
			wantNext:      false,
			routeScopes:   []string{ScopeAPIRead},
			newPrincipal: func() *mockPrincipalForAuth {
				return &mockPrincipalForAuth{scopes: []string{ScopeAPIRead}}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			customCheckerCalled := false
			var opts *AuthorizationOptions
			if tc.useRoles {
				opts = &AuthorizationOptions{
					CheckRoles: func(principal claims.Principal, roles []string) bool {
						customCheckerCalled = true
						return tc.checkerReturn
					},
				}
			} else {
				opts = &AuthorizationOptions{
					CheckScopes: func(principal claims.Principal, scopes []string) bool {
						customCheckerCalled = true
						return tc.checkerReturn
					},
				}
			}

			middleware := &authorizationMiddleware{options: opts}
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			ctx := routing.NewRouteContext(rec, req)
			ctx.SetUser(tc.newPrincipal())
			ctx.SetOptions(&routing.RouteOptions{Roles: tc.routeRoles, Scopes: tc.routeScopes})

			nextCalled := false
			next := func(c routing.RouteContext) {
				nextCalled = true
			}

			middleware.Invoke(ctx, next)

			assert.True(t, customCheckerCalled)
			assert.Equal(t, tc.wantNext, nextCalled)
		})
	}
}

func TestShouldAllowAccessWhenNoRolesOrScopesRequired(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{
		options: &AuthorizationOptions{},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, pathTest, nil)
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
	// Arrange - use Params slice
	ps := routing.AcquireParams()
	defer routing.ReleaseParams(ps)
	ps.Set("userId", "123")
	ps.Set("tenantId", "tenant-456")
	permission := permUserTenantTemplate

	// Act
	result := interpolatePermission(ps, permission)

	// Assert
	assert.Equal(t, "user:123:read:tenant:tenant-456", result)
}

func TestInterpolatePermissionShouldBeCaseInsensitive(t *testing.T) {
	// Arrange - use Params slice
	ps := routing.AcquireParams()
	defer routing.ReleaseParams(ps)
	ps.Set("UserId", "123")
	permission := permUseridTemplate

	// Act
	result := interpolatePermission(ps, permission)

	// Assert
	assert.Equal(t, "user:123:read", result)
}

func TestInterpolatePermissionShouldHandleNonexistentPlaceholders(t *testing.T) {
	// Arrange - use Params slice
	ps := routing.AcquireParams()
	defer routing.ReleaseParams(ps)
	ps.Set("userId", "123")
	permission := permNonexistent

	// Act
	result := interpolatePermission(ps, permission)

	// Assert
	assert.Equal(t, "user:nonexistent:read", result) // Should keep the placeholder name
}

func TestInterpolatePermissionsShouldRemoveDuplicates(t *testing.T) {
	// Arrange - use the optimized Params slice for replacements
	ps := routing.AcquireParams()
	defer routing.ReleaseParams(ps)
	ps.Set("id", "123")
	permissions1 := []string{permUserIDReadTemplate, user123Write}
	permissions2 := []string{PermUser123Read, adminIDDeleteTemplate}

	// Act
	result := interpolatePermissions(ps, permissions1, permissions2)

	// Assert
	expected := []string{PermUser123Read, user123Write, admin123Delete}
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
			Permissions:      []string{permAllowedPermission},
		},
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
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

func TestShouldDenyWhenNoUserAndRolesRequired(t *testing.T) {
	// Arrange
	middleware := &authorizationMiddleware{options: &AuthorizationOptions{}}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	// Intentionally do NOT set a user
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{roleAdmin}})

	nextCalled := false
	next := func(c routing.RouteContext) { nextCalled = true }

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestShouldDenyWhenNoUserAndRolesRequiredNilOptions(t *testing.T) {
	// Arrange: middleware constructed without explicit options
	middleware := &authorizationMiddleware{}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	// Route requires a role; with no user set this should deny, not panic
	ctx.SetOptions(&routing.RouteOptions{Roles: []string{roleAdmin}})

	nextCalled := false
	next := func(c routing.RouteContext) { nextCalled = true }

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.False(t, nextCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

// newAuthzCtx creates a DefaultRouteContext and attaches the provided user and options.
// It centralizes request + recorder creation used throughout tests and benchmarks.
func newAuthzCtx(user *mockPrincipalForAuth, opts *routing.RouteOptions) *routing.DefaultRouteContext {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, testURL, nil)
	rec := httptest.NewRecorder()
	ctx := routing.NewRouteContext(rec, req)
	if user != nil {
		ctx.SetUser(user)
	}
	if opts != nil {
		ctx.SetOptions(opts)
	}
	return ctx
}

// newDefaultAuthzUser returns a mock principal commonly used in benchmarks/tests.
func newDefaultAuthzUser() *mockPrincipalForAuth {
	return &mockPrincipalForAuth{roles: []string{roleAdmin, roleUser}, scopes: []string{scopeRead, scopeWrite}}
}
