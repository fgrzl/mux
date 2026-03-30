# Best Practices Guide

This guide outlines recommended patterns and conventions for building production-ready APIs with Mux.

## Project Structure

### Recommended Directory Layout

```
my-api/
├── main.go              # Application entry point
├── handlers/            # HTTP handlers
│   ├── users.go
│   ├── auth.go
│   └── health.go
├── middleware/          # Custom middleware
│   ├── auth.go
│   └── cors.go
├── models/              # Data models
│   ├── user.go
│   └── response.go
├── services/            # Business logic
│   ├── user_service.go
│   └── auth_service.go
├── config/              # Configuration
│   └── config.go
├── docs/                # API documentation
│   └── openapi.yaml
└── go.mod
```

### Package Organization

```go
// handlers/users.go
package handlers

import (
    "github.com/fgrzl/mux"
    "my-api/models"
    "my-api/services"
)

type UserHandler struct {
    userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) CreateUser(c mux.RouteContext) {
    // Handler implementation
}
```

## Router Configuration

### Router Setup Best Practices

```go
func NewRouter() (*mux.Router, error) {
    router := mux.NewRouter(
        mux.WithTitle("My API"),
        mux.WithVersion("1.0.0"),
        mux.WithDescription("Production API built with Mux"),
        mux.WithContact("API Support", "https://example.com/support", "support@example.com"),
        mux.WithLicense("MIT", "https://opensource.org/licenses/MIT"),
    ).Safe()

    // Register shared services before middleware/routes use them.
    setupServices(router)

    // Add middleware in correct order
    setupMiddleware(router)

    // Setup route groups
    setupRoutes(router)

    if err := router.Err(); err != nil {
        return nil, err
    }

    return router, nil
}

func setupServices(router *mux.Router) {
    router.Services().
        Register("auditWriter", auditWriter).
        Register("clock", systemClock)
}

func setupMiddleware(router *mux.Router) {
    // 1. Infrastructure middleware
    mux.UseForwardedHeaders(router)
    mux.UseLogging(router)

    // 2. Security middleware
    if os.Getenv("ENFORCE_HTTPS") == "true" {
        mux.UseEnforceHTTPS(router)
    }

    // 3. Application middleware
    mux.UseCompression(router)

    if os.Getenv("ENABLE_TRACING") == "true" {
        mux.UseOpenTelemetry(router)
    }

    // 4. Authentication (if needed globally)
    if authRequired() {
        mux.UseAuthentication(router,
            mux.WithValidator(validateToken),
            mux.WithTokenCreator(createToken),
        )
    }
}
```

Prefer explicit constructor injection for core domain services. Use `Services()` when middleware and handlers both need the same collaborator, or when a router, group, or route needs a scoped override.

## Route Organization

### Use Route Groups Effectively

```go
func setupRoutes(router *mux.Router) {
    // Health check (no auth required)
    router.GET("/health", healthHandler)
    
    // API v1
    v1 := router.NewRouteGroup("/api/v1")
    v1.WithTags("API v1")
    
    // Public endpoints
    public := v1.NewRouteGroup("/public")
    setupPublicRoutes(public)
    
    // Protected endpoints
    protected := v1.NewRouteGroup("/protected")
    protected.RequireRoles("user") // All routes require authentication
    setupProtectedRoutes(protected)
    
    // Admin endpoints
    admin := v1.NewRouteGroup("/admin")
    admin.RequireRoles("admin")
    setupAdminRoutes(admin)
}

func setupUserRoutes(group *mux.RouteGroup) {
    users := group.NewRouteGroup("/users")
    users.WithTags("Users")
    
    users.GET("/", listUsers).
        WithSummary("List all users").
        WithParam("page", "query", "Page number", 1, false).
        WithParam("limit", "query", "Page size", 10, false).
        WithOKResponse([]User{})
        
    users.POST("/", createUser).
        WithSummary("Create a new user").
        WithJsonBody(User{}).
        WithCreatedResponse(User{}).
        WithRateLimit(10, time.Minute) // Rate limit user creation
        
    users.GET("/{id}", getUser).
        WithSummary("Get user by ID").
        WithPathParam("id", "The unique user identifier", uuid.Nil).
        WithOKResponse(User{}).
        WithNotFoundResponse()
        
    users.PUT("/{id}", updateUser).
        WithSummary("Update user").
        WithPathParam("id", "The unique user identifier", uuid.Nil).
        WithJsonBody(User{}).
        WithOKResponse(User{}).
        RequirePermission("write") // Additional permission check
        
    users.DELETE("/{id}", deleteUser).
        WithSummary("Delete user").
        WithPathParam("id", "The unique user identifier", uuid.Nil).
        WithNoContentResponse().
        RequirePermission("delete")
}
```

When route groups are built from dynamic inputs such as config files or generators, prefer the `Err` variants like `WithParamErr`, `WithPathParamErr`, and `WithRequiredQueryParamErr` so startup validation can return actionable errors instead of panicking immediately.

## Handler Implementation

### Handler Best Practices

```go
// Use dependency injection
type UserHandler struct {
    userService UserService
    logger      *slog.Logger
}

func (h *UserHandler) CreateUser(c mux.RouteContext) {
    // 1. Input validation
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        h.logger.WarnContext(c, "invalid request body", "error", err)
        c.BadRequest("Invalid request", "Request body must be valid JSON")
        return
    }
    
    // 2. Business logic validation
    if err := req.Validate(); err != nil {
        c.BadRequest("Validation failed", err.Error())
        return
    }
    
    // 3. Authorization (if not handled by middleware)
    principal := c.User()
    if principal == nil {
        c.Unauthorized()
        return
    }
    
    // 4. Business logic execution
    user, err := h.userService.CreateUser(c, req.ToModel())
    if err != nil {
        h.logger.ErrorContext(c, "failed to create user", "error", err)
        
        // Handle different error types
        switch {
        case errors.Is(err, ErrUserExists):
            c.Conflict("User exists", "A user with this email already exists")
        case errors.Is(err, ErrInvalidInput):
            c.BadRequest("Invalid input", err.Error())
        default:
            c.ServerError("Failed to create user", "")
        }
        return
    }
    
    // 5. Success response
    h.logger.InfoContext(c, "user created", "user_id", user.ID, "created_by", principal.Subject())
    c.Created(user.ToResponse())
}

// Request/Response models
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=100"`
    Email string `json:"email" validate:"required,email"`
}

func (r *CreateUserRequest) Validate() error {
    if r.Name == "" {
        return errors.New("name is required")
    }
    if r.Email == "" {
        return errors.New("email is required")
    }
    if !strings.Contains(r.Email, "@") {
        return errors.New("invalid email format")
    }
    return nil
}

func (r *CreateUserRequest) ToModel() *User {
    return &User{
        Name:  r.Name,
        Email: r.Email,
    }
}
```

### Error Handling Patterns

```go
// Define application-specific errors
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserExists   = errors.New("user already exists")
    ErrInvalidInput = errors.New("invalid input")
)

// Centralized error handling
func handleServiceError(c mux.RouteContext, err error, operation string) {
    logger := getLoggerFromContext(c)
    logger.ErrorContext(c, operation+" failed", "error", err)
    
    switch {
    case errors.Is(err, ErrUserNotFound):
        c.NotFound()
    case errors.Is(err, ErrUserExists):
        c.Conflict("Resource exists", "The resource already exists")
    case errors.Is(err, ErrInvalidInput):
        c.BadRequest("Invalid input", err.Error())
    case errors.Is(err, context.DeadlineExceeded):
        c.Problem(&mux.ProblemDetails{
            Title:  "Request timeout",
            Detail: "The request took too long to process",
            Status: http.StatusRequestTimeout,
        })
    default:
        c.ServerError("Operation failed", "")
    }
}

// Usage in handlers
func (h *UserHandler) GetUser(c mux.RouteContext) {
    userID, ok := c.ParamUUID("id")
    if !ok {
        c.BadRequest("Invalid ID", "User ID must be a valid UUID")
        return
    }
    
    user, err := h.userService.GetUser(c, userID)
    if err != nil {
        handleServiceError(c, err, "get user")
        return
    }
    
    c.OK(user.ToResponse())
}
```

## Configuration Management

### Environment-based Configuration

```go
// config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
    Features FeatureFlags
}

type ServerConfig struct {
    Port            string        `env:"PORT" envDefault:"8080"`
    ReadTimeout     time.Duration `env:"READ_TIMEOUT" envDefault:"30s"`
    WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
    ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
}

type FeatureFlags struct {
    EnableHTTPS       bool `env:"ENABLE_HTTPS" envDefault:"false"`
    EnableCompression bool `env:"ENABLE_COMPRESSION" envDefault:"true"`
    EnableTracing     bool `env:"ENABLE_TRACING" envDefault:"false"`
    EnableRateLimit   bool `env:"ENABLE_RATE_LIMIT" envDefault:"true"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, fmt.Errorf("failed to parse config: %w", err)
    }
    return cfg, nil
}

// Usage in main.go
func main() {
    config, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    router := setupRouter(config)
    server := setupServer(router, config)
    
    // Graceful shutdown
    gracefulShutdown(server, config.Server.ShutdownTimeout)
}
```

## Testing Strategies

### Handler Testing

```go
func TestCreateUser(t *testing.T) {
    // Setup
    userService := &mockUserService{}
    handler := &UserHandler{userService: userService}
    
    router := mux.NewRouter()
    router.POST("/users", handler.CreateUser)
    
    tests := []struct {
        name           string
        body           string
        expectedStatus int
        expectedBody   string
        setupMock      func()
    }{
        {
            name:           "valid user creation",
            body:           `{"name":"John Doe","email":"john@example.com"}`,
            expectedStatus: http.StatusCreated,
            setupMock: func() {
                userService.On("CreateUser", mock.Anything, mock.Anything).
                    Return(&User{ID: uuid.New(), Name: "John Doe"}, nil)
            },
        },
        {
            name:           "invalid json",
            body:           `{"name":}`,
            expectedStatus: http.StatusBadRequest,
            setupMock:      func() {},
        },
        {
            name:           "duplicate user",
            body:           `{"name":"John Doe","email":"john@example.com"}`,
            expectedStatus: http.StatusConflict,
            setupMock: func() {
                userService.On("CreateUser", mock.Anything, mock.Anything).
                    Return(nil, ErrUserExists)
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setupMock()
            
            req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()
            
            rtr.ServeHTTP(rec, req)
            
            assert.Equal(t, tt.expectedStatus, rec.Code)
            if tt.expectedBody != "" {
                assert.Contains(t, rec.Body.String(), tt.expectedBody)
            }
        })
    }
}
```

### Integration Testing

```go
func TestUserAPIIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()
    
    // Setup services
    userService := services.NewUserService(db)
    
    // Setup router
    router := setupTestRouter(userService)
    
    // Test complete flow
    t.Run("create and retrieve user", func(t *testing.T) {
        // Create user
        createBody := `{"name":"Jane Doe","email":"jane@example.com"}`
        req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(createBody))
        req.Header.Set("Content-Type", "application/json")
        rec := httptest.NewRecorder()
        
        rtr.ServeHTTP(rec, req)
        
        require.Equal(t, http.StatusCreated, rec.Code)
        
        var user User
        err := json.Unmarshal(rec.Body.Bytes(), &user)
        require.NoError(t, err)
        
        // Retrieve user
        req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", user.ID), nil)
        rec = httptest.NewRecorder()
        
        rtr.ServeHTTP(rec, req)
        
        require.Equal(t, http.StatusOK, rec.Code)
        
        var retrievedUser User
        err = json.Unmarshal(rec.Body.Bytes(), &retrievedUser)
        require.NoError(t, err)
        
        assert.Equal(t, user.ID, retrievedUser.ID)
        assert.Equal(t, "Jane Doe", retrievedUser.Name)
    })
}
```

## Security Best Practices

### Authentication and Authorization

```go
// Implement proper token validation
func validateToken(tokenString string) (claims.Principal, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Validate signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(os.Getenv("JWT_SECRET")), nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }
    
    if !token.Valid {
        return nil, errors.New("token is not valid")
    }
    
    // Extract claims
    mapClaims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }
    
    // Validate required claims
    if exp, ok := mapClaims["exp"].(float64); ok {
        if time.Unix(int64(exp), 0).Before(time.Now()) {
            return nil, errors.New("token has expired")
        }
    }
    
    claimSet := claims.NewClaimsSet("")
    if sub, ok := mapClaims["sub"].(string); ok {
        claimSet.SetSubject(sub)
    }
    if roles, ok := mapClaims["roles"].([]interface{}); ok {
        roleStrings := make([]string, len(roles))
        for i, role := range roles {
            if roleStr, ok := role.(string); ok {
                roleStrings[i] = roleStr
            }
        }
        claimSet.SetRoles(roleStrings...)
    }

    return claims.NewPrincipal(claimSet), nil
}

// Implement input validation
func validateInput(c mux.RouteContext, req interface{}) error {
    validate := validator.New()
    if err := validate.Struct(req); err != nil {
        var validationErrors []string
        for _, err := range err.(validator.ValidationErrors) {
            validationErrors = append(validationErrors, 
                fmt.Sprintf("field '%s' failed validation: %s", err.Field(), err.Tag()))
        }
        return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, ", "))
    }
    return nil
}
```

### Rate Limiting Strategy

```go
// Apply rate limiting strategically
func setupRateLimiting(router *mux.Router) {
    // Strict limits for expensive operations
    router.POST("/api/v1/users", createUser).
        WithRateLimit(10, time.Minute)
        
    // Moderate limits for search/list operations
    router.GET("/api/v1/users", listUsers).
        WithRateLimit(100, time.Minute)
        
    // Generous limits for read operations
    router.GET("/api/v1/users/{id}", getUser).
        WithRateLimit(1000, time.Minute)
        
    // Very strict limits for auth operations
    router.POST("/api/v1/auth/login", login).
        WithRateLimit(5, time.Minute)
}
```

## Performance Optimization

### Database Integration

```go
// Use connection pooling
func setupDatabase() *sql.DB {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    db.SetConnMaxIdleTime(30 * time.Second)
    
    return db
}

// Use context for timeouts
func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (*User, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    query := "SELECT id, name, email, created_at FROM users WHERE id = $1"
    row := s.db.QueryRowContext(ctx, query, userID)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Created)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return &user, nil
}
```

### Caching Strategies

```go
// Implement caching for expensive operations
type CachedUserService struct {
    userService UserService
    cache       Cache
}

func (s *CachedUserService) GetUser(ctx context.Context, userID uuid.UUID) (*User, error) {
    // Check cache first
    cacheKey := fmt.Sprintf("user:%s", userID)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        var user User
        if err := json.Unmarshal(cached, &user); err == nil {
            return &user, nil
        }
    }
    
    // Get from database
    user, err := s.userService.GetUser(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    if data, err := json.Marshal(user); err == nil {
        s.cache.Set(ctx, cacheKey, data, 5*time.Minute)
    }
    
    return user, nil
}
```

## Production Deployment

### Graceful Shutdown

```go
func main() {
    router, err := NewRouter()
    if err != nil {
        log.Fatal(err)
    }

    server := mux.NewServer(
        ":8080",
        router,
        mux.WithReadTimeout(30*time.Second),
        mux.WithWriteTimeout(30*time.Second),
        mux.WithIdleTimeout(120*time.Second),
    )

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    log.Println("Server starting on :8080")
    if err := server.Listen(ctx); err != nil {
        log.Fatal(err)
    }
}
```

Prefer `mux.NewServer(...)` unless you need lower-level control than the framework's server wrapper exposes.

### Health Checks

```go
func setupHealthChecks(router *mux.Router, db *sql.DB) {
    router.GET("/health", func(c mux.RouteContext) {
        c.OK(map[string]string{
            "status":    "healthy",
            "timestamp": time.Now().Format(time.RFC3339),
            "version":   os.Getenv("APP_VERSION"),
        })
    })
    
    router.GET("/health/deep", func(c mux.RouteContext) {
        checks := make(map[string]interface{})
        
        // Database check
        if err := db.PingContext(c); err != nil {
            checks["database"] = map[string]interface{}{
                "status": "unhealthy",
                "error":  err.Error(),
            }
        } else {
            checks["database"] = map[string]string{"status": "healthy"}
        }
        
        // Determine overall health
        healthy := true
        for _, check := range checks {
            if checkMap, ok := check.(map[string]interface{}); ok {
                if status, exists := checkMap["status"]; exists && status != "healthy" {
                    healthy = false
                    break
                }
            }
        }
        
        response := map[string]interface{}{
            "status":    map[string]string{"healthy": "healthy", "unhealthy": "unhealthy"}[fmt.Sprintf("%t", healthy)],
            "timestamp": time.Now().Format(time.RFC3339),
            "checks":    checks,
        }
        
        if healthy {
            c.OK(response)
        } else {
            c.Problem(&mux.ProblemDetails{
                Title:  "Service unhealthy",
                Detail: "One or more health checks failed",
                Status: http.StatusServiceUnavailable,
            })
        }
    })
}
```

## Monitoring and Observability

### Structured Logging

```go
func setupLogging() *slog.Logger {
    level := slog.LevelInfo
    if os.Getenv("DEBUG") == "true" {
        level = slog.LevelDebug
    }
    
    opts := &slog.HandlerOptions{
        Level: level,
        ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
            // Remove time from logs in development
            if a.Key == slog.TimeKey && os.Getenv("ENV") == "development" {
                return slog.Attr{}
            }
            return a
        },
    }
    
    var handler slog.Handler
    if os.Getenv("ENV") == "production" {
        handler = slog.NewJSONHandler(os.Stdout, opts)
    } else {
        handler = slog.NewTextHandler(os.Stdout, opts)
    }
    
    return slog.New(handler)
}
```

### Metrics Collection

```go
// Add custom middleware for metrics
type MetricsMiddleware struct {
    requestDuration prometheus.HistogramVec
    requestCount    prometheus.CounterVec
}

func NewMetricsMiddleware() *MetricsMiddleware {
    return &MetricsMiddleware{
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "http_request_duration_seconds",
                Help: "HTTP request duration in seconds",
            },
            []string{"method", "path", "status"},
        ),
        requestCount: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "path", "status"},
        ),
    }
}

func (m *MetricsMiddleware) Invoke(c mux.RouteContext, next mux.HandlerFunc) {
    start := time.Now()
    rec := &statusRecorder{ResponseWriter: c.Response()}
    c.SetResponse(rec)
    
    next(c)
    
    duration := time.Since(start)
    status := fmt.Sprintf("%d", rec.Status)
    
    m.requestDuration.WithLabelValues(
        c.Request().Method,
        c.Request().URL.Path,
        status,
    ).Observe(duration.Seconds())
    
    m.requestCount.WithLabelValues(
        c.Request().Method,
        c.Request().URL.Path,
        status,
    ).Inc()
}
```

## Summary Checklist

### ✅ Development Best Practices
- [ ] Use structured project layout
- [ ] Implement proper error handling
- [ ] Use dependency injection
- [ ] Write comprehensive tests
- [ ] Use type-safe parameter extraction
- [ ] Implement request validation
- [ ] Use structured logging

### ✅ Security Best Practices
- [ ] Implement proper authentication
- [ ] Use role-based authorization
- [ ] Apply rate limiting strategically
- [ ] Validate all inputs
- [ ] Use HTTPS in production
- [ ] Implement proper token validation

### ✅ Performance Best Practices
- [ ] Use connection pooling
- [ ] Implement caching where appropriate
- [ ] Use context timeouts
- [ ] Add middleware judiciously
- [ ] Optimize database queries
- [ ] Use compression for large responses

### ✅ Production Best Practices
- [ ] Implement graceful shutdown
- [ ] Add comprehensive health checks
- [ ] Use environment-based configuration
- [ ] Implement proper logging
- [ ] Add metrics and monitoring
- [ ] Use proper deployment practices

Following these best practices will help you build robust, maintainable, and production-ready APIs with Mux.

## See Also

- [Getting Started](getting-started.md) - Comprehensive introduction
- [Router](router.md) - Routing fundamentals
- [Middleware](middleware.md) - Built-in middleware guide
- [WebServer](webserver.md) - Production server setup
- [Health Probes](health-probes.md) - Kubernetes-style health checks
- [Custom Middleware](custom-middleware.md) - Build your own middleware