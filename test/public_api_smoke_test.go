package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fgrzl/claims"
	"github.com/fgrzl/mux"
	"github.com/stretchr/testify/require"
)

type smokeContextKey string

const (
	smokeRequestContextKey smokeContextKey = "request"
	smokeRouteContextKey   smokeContextKey = "route"
)

type smokeHeaderWriter struct {
	http.ResponseWriter
}

func (w *smokeHeaderWriter) WriteHeader(statusCode int) {
	w.Header().Set("X-Smoke", "wrapped")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *smokeHeaderWriter) Write(data []byte) (int, error) {
	if w.Header().Get("X-Smoke") == "" {
		w.Header().Set("X-Smoke", "wrapped")
	}
	return w.ResponseWriter.Write(data)
}

type smokeMiddleware struct{}

func (smokeMiddleware) Invoke(c mux.MutableRouteContext, next mux.HandlerFunc) {
	ctx := context.WithValue(c.Request().Context(), smokeRequestContextKey, "request-context")
	c.SetRequest(c.Request().WithContext(ctx))
	c.SetContextValue(smokeRouteContextKey, "route-context")
	c.SetResponse(&smokeHeaderWriter{ResponseWriter: c.Response()})
	next(c)
}

type smokePrincipal struct{ subject string }

func (p smokePrincipal) Subject() string                      { return p.subject }
func (p smokePrincipal) Issuer() string                       { return "" }
func (p smokePrincipal) Audience() []string                   { return nil }
func (p smokePrincipal) ExpirationTime() int64                { return 0 }
func (p smokePrincipal) NotBefore() int64                     { return 0 }
func (p smokePrincipal) IssuedAt() int64                      { return 0 }
func (p smokePrincipal) JWTI() string                         { return "" }
func (p smokePrincipal) Scopes() []string                     { return nil }
func (p smokePrincipal) Roles() []string                      { return nil }
func (p smokePrincipal) Email() string                        { return "" }
func (p smokePrincipal) Username() string                     { return "" }
func (p smokePrincipal) CustomClaim(name string) claims.Claim { return nil }
func (p smokePrincipal) CustomClaimValue(name string) string  { return "" }
func (p smokePrincipal) Claims() *claims.ClaimSet             { return nil }

func TestPublicAPIGoldenPathSmoke(t *testing.T) {
	t.Run("quick start", func(t *testing.T) {
		router := mux.NewRouter()

		err := router.Configure(func(router *mux.Router) {
			router.GET("/hello", func(c mux.RouteContext) {
				c.OK(map[string]string{"message": "hello"})
			})
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/hello", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Equal(t, "hello", payload["message"])
	})

	t.Run("grouped routes", func(t *testing.T) {
		type createUserRequest struct {
			Name string `json:"name"`
		}
		type createUserResponse struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			Verbose bool   `json:"verbose"`
		}

		router := mux.NewRouter(
			mux.WithTitle("Smoke API"),
			mux.WithVersion("1.0.0"),
			mux.WithDescription("Golden path smoke test"),
		)

		api := router.Group("/api")
		api.WithTags("api")

		api.POST("/users/{id}", func(c mux.RouteContext) {
			id, ok := c.Params().Int("id")
			if !ok {
				c.BadRequest("Missing ID", "id path parameter is required")
				return
			}

			verbose, _ := c.Query().Bool("verbose")

			var body createUserRequest
			if err := c.Bind(&body); err != nil {
				c.BadRequest("Invalid body", err.Error())
				return
			}

			c.Created(createUserResponse{ID: id, Name: body.Name, Verbose: verbose})
		}).
			WithOperationID("createUser").
			WithSummary("Create user").
			WithDescription("Creates a user from grouped routes").
			WithTags("users").
			WithPathParam("id", "User identifier", 1).
			WithJsonBody(createUserRequest{}).
			WithCreatedResponse(createUserResponse{}).
			WithResponse(http.StatusBadRequest, mux.ProblemDetails{})

		req := httptest.NewRequest(
			http.MethodPost,
			"/api/users/41?verbose=true",
			strings.NewReader(`{"name":"Ada"}`),
		)
		req.Header.Set(mux.HeaderContentType, mux.MimeJSON)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusCreated, rec.Code)

		var payload createUserResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Equal(t, createUserResponse{ID: 41, Name: "Ada", Verbose: true}, payload)
	})

	t.Run("grouped request accessors", func(t *testing.T) {
		router := mux.NewRouter()

		router.POST("/access/{id}", func(c mux.RouteContext) {
			params := c.Params()
			query := c.Query()
			form := c.Form()
			headers := c.Headers()
			cookies := c.Cookies()

			id, ok := params.Int("id")
			if !ok {
				c.BadRequest("Missing ID", "id path parameter is required")
				return
			}

			page, _ := query.Int("page")
			verbose, _ := query.Bool("verbose")
			redirectURL := query.GetRedirectURL("/fallback")

			name, ok := form.String("name")
			if !ok {
				c.BadRequest("Missing name", "name form field is required")
				return
			}

			traceID, _ := headers.String("X-Trace-ID")
			session, err := cookies.Get("session")
			if err != nil {
				c.BadRequest("Missing session", err.Error())
				return
			}

			c.OK(map[string]any{
				"id":       id,
				"page":     page,
				"redirect": redirectURL,
				"verbose":  verbose,
				"name":     name,
				"trace":    traceID,
				"session":  session,
			})
		})

		req := httptest.NewRequest(
			http.MethodPost,
			"/access/41?page=2&verbose=true&return_to=dashboard",
			strings.NewReader("name=Ada"),
		)
		req.Header.Set(mux.HeaderContentType, "application/x-www-form-urlencoded")
		req.Header.Set("X-Trace-ID", "trace-123")
		req.AddCookie(&http.Cookie{Name: "session", Value: "session-abc"})

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)

		var payload map[string]any
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Equal(t, float64(41), payload["id"])
		require.Equal(t, float64(2), payload["page"])
		require.Equal(t, "/dashboard", payload["redirect"])
		require.Equal(t, true, payload["verbose"])
		require.Equal(t, "Ada", payload["name"])
		require.Equal(t, "trace-123", payload["trace"])
		require.Equal(t, "session-abc", payload["session"])
	})

	t.Run("middleware", func(t *testing.T) {
		router := mux.NewRouter()
		mux.UseLogging(router)
		mux.UseCORS(router,
			mux.WithCORSAllowedOrigins("https://example.com"),
			mux.WithCORSAllowedMethods(http.MethodGet),
			mux.WithCORSAllowedHeaders(mux.HeaderContentType),
		)
		router.Use(smokeMiddleware{})

		router.GET("/middleware", func(c mux.RouteContext) {
			requestValue, _ := c.Request().Context().Value(smokeRequestContextKey).(string)
			routeValue, _ := c.Value(smokeRouteContextKey).(string)
			c.OK(map[string]string{
				"request": requestValue,
				"route":   routeValue,
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/middleware", nil)
		req.Header.Set("Origin", "https://example.com")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "wrapped", rec.Header().Get("X-Smoke"))
		require.Equal(t, "https://example.com", rec.Header().Get("Access-Control-Allow-Origin"))

		var payload map[string]string
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&payload))
		require.Equal(t, "request-context", payload["request"])
		require.Equal(t, "route-context", payload["route"])
	})

	t.Run("incremental adoption", func(t *testing.T) {
		router := mux.NewRouter()
		router.Services().Register(mux.ServiceKey("db"), "database")

		router.Handle(http.MethodGet, "/legacy", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))

		router.HandleFunc(http.MethodGet, "/bridge", func(w http.ResponseWriter, r *http.Request) {
			routeCtx, ok := mux.RouteContextFromRequest(r)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			svc, ok := routeCtx.Services().Get(mux.ServiceKey("db"))
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			require.NoError(t, json.NewEncoder(w).Encode(map[string]string{"service": svc.(string)}))
		})

		legacyReq := httptest.NewRequest(http.MethodGet, "/legacy", nil)
		legacyRec := httptest.NewRecorder()
		router.ServeHTTP(legacyRec, legacyReq)
		require.Equal(t, http.StatusNoContent, legacyRec.Code)

		bridgeReq := httptest.NewRequest(http.MethodGet, "/bridge", nil)
		bridgeRec := httptest.NewRecorder()
		router.ServeHTTP(bridgeRec, bridgeReq)

		require.Equal(t, http.StatusOK, bridgeRec.Code)

		var payload map[string]string
		require.NoError(t, json.NewDecoder(bridgeRec.Body).Decode(&payload))
		require.Equal(t, "database", payload["service"])
	})

	t.Run("openapi generation", func(t *testing.T) {
		router := mux.NewRouter(
			mux.WithTitle("Smoke API"),
			mux.WithVersion("1.0.0"),
			mux.WithDescription("OpenAPI generation smoke test"),
		)

		router.GET("/health", func(c mux.RouteContext) {
			c.NoContent()
		}).
			WithOperationID("healthCheck").
			WithSummary("Health check").
			WithNoContentResponse()

		spec, err := mux.GenerateSpecWithGenerator(mux.NewGenerator(), router)
		require.NoError(t, err)
		require.NotNil(t, spec)
		require.NoError(t, spec.Validate())

		var doc map[string]any
		require.NoError(t, json.Unmarshal(specJSONBytes(t, spec), &doc))

		info := requireMap(t, doc["info"])
		require.Equal(t, "Smoke API", info["title"])
		paths := requireMap(t, doc["paths"])
		healthPath := requireMap(t, paths["/health"])
		getOp := requireMap(t, healthPath["get"])
		require.Equal(t, "Health check", getOp["summary"])
	})

	t.Run("auth cookie naming and csrf", func(t *testing.T) {
		router := mux.NewRouter()
		mux.UseAuthentication(router,
			mux.WithAuthValidator(func(token string) (claims.Principal, error) {
				if !strings.HasPrefix(token, "signed:") {
					return nil, fmt.Errorf("invalid token")
				}
				return smokePrincipal{subject: strings.TrimPrefix(token, "signed:")}, nil
			}),
			mux.WithAuthTokenCreator(func(principal claims.Principal, _ time.Duration) (string, error) {
				return "signed:" + principal.Subject(), nil
			}),
			mux.WithAuthCSRFProtection(),
			mux.WithAuthAppSessionCookieName("session_token"),
			mux.WithAuthTwoFactorCookieName("step_up_token"),
			mux.WithAuthIDPSessionCookieName("idp_token"),
		)

		router.POST("/login", func(c mux.RouteContext) {
			if _, err := c.Cookies().CSRFTokenErr(); err != nil {
				c.ServerError("csrf", err.Error())
				return
			}
			c.Cookies().SignIn(smokePrincipal{subject: "ada"}, "/dashboard")
		}).AllowAnonymous()

		router.POST("/protected", func(c mux.RouteContext) {
			user := c.User()
			require.NotNil(t, user)
			c.OK(map[string]string{"subject": user.Subject()})
		}).WithOKResponse(map[string]string{"subject": "ada"})

		loginReq := httptest.NewRequest(http.MethodPost, "/login", nil)
		loginRec := httptest.NewRecorder()
		router.ServeHTTP(loginRec, loginReq)

		require.Equal(t, http.StatusTemporaryRedirect, loginRec.Code)

		result := loginRec.Result()
		cookies := result.Cookies()
		require.NotEmpty(t, cookies)

		var csrfToken string
		var sawSessionCookie bool
		for _, cookie := range cookies {
			switch cookie.Name {
			case "session_token":
				sawSessionCookie = true
			case "csrf_token":
				csrfToken = cookie.Value
			case "app_token":
				t.Fatalf("unexpected default app session cookie in response")
			}
		}
		require.True(t, sawSessionCookie)
		require.NotEmpty(t, csrfToken)

		protectedReq := httptest.NewRequest(http.MethodPost, "/protected", nil)
		for _, cookie := range cookies {
			protectedReq.AddCookie(cookie)
		}
		protectedReq.Header.Set("X-CSRF-Token", csrfToken)

		protectedRec := httptest.NewRecorder()
		router.ServeHTTP(protectedRec, protectedReq)

		require.Equal(t, http.StatusOK, protectedRec.Code)

		var payload map[string]string
		require.NoError(t, json.NewDecoder(protectedRec.Body).Decode(&payload))
		require.Equal(t, "ada", payload["subject"])
	})

	t.Run("new route context helper", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/unit", nil)
		rec := httptest.NewRecorder()
		ctx := mux.NewRouteContext(rec, req)

		smokeMiddleware{}.Invoke(ctx, func(c mux.RouteContext) {
			require.Equal(t, "request-context", c.Request().Context().Value(smokeRequestContextKey))
			require.Equal(t, "route-context", c.Value(smokeRouteContextKey))
			c.NoContent()
		})

		require.Equal(t, http.StatusNoContent, rec.Code)
		require.Equal(t, "wrapped", rec.Header().Get("X-Smoke"))
	})

	t.Run("web server bootstrap", func(t *testing.T) {
		router := mux.NewRouter()
		require.NoError(t, router.Configure(func(router *mux.Router) {
			router.GET("/health", func(c mux.RouteContext) {
				c.NoContent()
			})
		}))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		time.AfterFunc(50*time.Millisecond, cancel)

		server := mux.NewServer("127.0.0.1:0", router)
		require.NoError(t, server.Listen(ctx))
	})
}
