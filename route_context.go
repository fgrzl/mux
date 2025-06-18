package mux

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/fgrzl/claims"
)

type RouteContext struct {
	context.Context
	Response http.ResponseWriter
	Request  *http.Request
	User     claims.Principal
	Options  *RouteOptions
	Params   RouteParams
}

func NewRouteContext(w http.ResponseWriter, r *http.Request) *RouteContext {
	return &RouteContext{
		Context:  r.Context(),
		Response: w,
		Request:  r,
	}
}

type ProblemDetails struct {
	Type     string  `json:"type"`               // A URI reference to the error type
	Title    string  `json:"title"`              // A brief summary of the error
	Status   int     `json:"status"`             // HTTP status code
	Detail   string  `json:"detail"`             // Specific details about the error
	Instance *string `json:"instance,omitempty"` // A URI identifying the error instance, nullable
}

type RouteParams map[string]string

type RouteOptions struct {
	Method         string
	Pattern        string
	Handler        HandlerFunc
	AllowAnonymous bool
	Roles          []string
	Scopes         []string
	Permissions    []string
	RateLimit      int
	RateInterval   time.Duration

	// Documentation fields
	OperationID   string
	Description   string
	Summary       string
	Parameters    []ParameterDescriptor
	RequestBodies []RequestBodyDescriptor
	Responses     []ResponseDescriptor

	// Dependencies
	AuthProvider AuthProvider
}

type ParameterDescriptor struct {
	Description string
	Source      string // "query", "header", "path", "cookie"
	Model       any
}

type RequestDescriptor struct {
	Description string
	ContentType string
	Model       any
}

type RequestBodyDescriptor struct {
	Description string
	ContentType string
	Model       any
}

type ResponseDescriptor struct {
	StatusCode  int
	Description string
	ContentType string
	Model       any
}

func (c *RouteContext) Bind(model any) error {
	staging := make(map[string]any)

	switch c.Request.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		for key, values := range c.Request.URL.Query() {
			addToStaging(staging, key, values)
		}
	case http.MethodPut, http.MethodPost:
		c.Request.Body = http.MaxBytesReader(c.Response, c.Request.Body, 1<<20) // 1MB max
		ct := c.Request.Header.Get("Content-Type")
		if ct == "application/x-www-form-urlencoded" {
			if err := c.Request.ParseForm(); err != nil {
				return err
			}
			for key, values := range c.Request.Form {
				addToStaging(staging, key, values)
			}
		} else if strings.HasPrefix(ct, "application/json") {
			bodyMap := make(map[string]any)
			decoder := json.NewDecoder(c.Request.Body)
			if err := decoder.Decode(&bodyMap); err != nil {
				return err
			}
			for key, val := range bodyMap {
				staging[key] = val
			}
		} else {
			return errors.New("unsupported content type")
		}
	}

	for key, headerValues := range c.Request.Header {
		addToStaging(staging, key, headerValues)
	}
	for key, paramValue := range c.Params {
		staging[key] = paramValue
	}

	marshaledData, err := json.Marshal(staging)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshaledData, model)
}

func (c *RouteContext) Param(name string) (string, bool) {
	val, ok := c.Params[name]
	return val, ok
}

func (c *RouteContext) QueryValue(name string) (string, bool) {
	vals, ok := c.Request.URL.Query()[name]
	if ok && len(vals) > 0 {
		return vals[0], true
	}
	return "", false
}

func (c *RouteContext) QueryValues(name string) ([]string, bool) {
	vals, ok := c.Request.URL.Query()[name]
	return vals, ok
}

func (c *RouteContext) ServerError(title, detail string) {
	if title == "" {
		title = http.StatusText(http.StatusInternalServerError)
	}
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusInternalServerError,
		Type:     "about:blank",
		Instance: getInstanceURI(c.Request),
	})
}

func (c *RouteContext) BadRequest(title, detail string) {
	if title == "" {
		title = http.StatusText(http.StatusBadRequest)
	}
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusBadRequest,
		Type:     "about:blank",
		Instance: getInstanceURI(c.Request),
	})
}

func (c *RouteContext) Conflict(title, detail string) {
	c.Problem(&ProblemDetails{
		Title:    title,
		Detail:   detail,
		Status:   http.StatusConflict,
		Type:     "about:blank",
		Instance: getInstanceURI(c.Request),
	})
}

func (c *RouteContext) Problem(problem *ProblemDetails) {
	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}
	slog.Warn("problem response",
		slog.Int("status", problem.Status),
		slog.String("title", problem.Title),
		slog.String("detail", problem.Detail),
	)
	r := c.Response
	r.Header().Set("Content-Type", "application/problem+json")
	r.WriteHeader(problem.Status)
	json.NewEncoder(r).Encode(problem)
}

func (c *RouteContext) OK(model any) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(http.StatusOK)
	json.NewEncoder(c.Response).Encode(model)
}

// Plain writes a text/plain response
func (c *RouteContext) Plain(status int, data []byte) {
	c.Response.Header().Set("Content-Type", "text/plain")
	c.Response.WriteHeader(status)
	if len(data) > 0 {
		c.Response.Write(data)
	}
}

func (c *RouteContext) Created(model any) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(http.StatusCreated)
	if model != nil {
		json.NewEncoder(c.Response).Encode(model)
	}
}

func (c *RouteContext) Accept(model any) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(http.StatusAccepted)
	if model != nil {
		json.NewEncoder(c.Response).Encode(model)
	}
}

func (c *RouteContext) NoContent() {
	c.Response.WriteHeader(http.StatusNoContent)
}

func (c *RouteContext) NotFound() {
	http.NotFound(c.Response, c.Request)
}

func (c *RouteContext) Unauthorized() {
	http.Error(c.Response, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (c *RouteContext) Forbidden(message string) {
	http.Error(c.Response, message, http.StatusForbidden)
}

func (c *RouteContext) GetRedirectScheme() string {
	if scheme := c.Request.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if c.Request.URL.Scheme != "" {
		return c.Request.URL.Scheme
	}
	return "http"
}

func (c *RouteContext) TemporaryRedirect(url string) {
	target := ensureAbsoluteURL(c, url)
	http.Redirect(c.Response, c.Request, target, http.StatusTemporaryRedirect)
}

func (c *RouteContext) PermanentRedirect(url string) {
	target := ensureAbsoluteURL(c, url)
	http.Redirect(c.Response, c.Request, target, http.StatusPermanentRedirect)
}

func ensureAbsoluteURL(c *RouteContext, url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	scheme := c.GetRedirectScheme()
	return fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, url)
}

func (c *RouteContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	http.SetCookie(c.Response, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}

func (c *RouteContext) GetCookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (c *RouteContext) ClearCookie(name string) {
	http.SetCookie(c.Response, &http.Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
	})
}

func (c *RouteContext) Authenticate(cookieName string, user claims.Principal) {

	if c.Options.AuthProvider == nil {
		panic("a signer is required if using authentication")
	}

	token, err := c.Options.AuthProvider.CreateToken(c, user)
	if err != nil {
		slog.ErrorContext(c, "failed to create token")
		return
	}

	c.SetCookie(cookieName, token, 0, "/", "", true, true)
}

func (c *RouteContext) SignIn(user claims.Principal, redirectUrl string) {
	c.Authenticate(GetUserCookieName(), user)
	if redirectUrl == "" {
		redirectUrl = "/"
	}
	c.TemporaryRedirect(redirectUrl)
}

func (c *RouteContext) SignOut() {
	c.ClearCookie(GetUserCookieName())
	c.ClearCookie(GetTwoFactorCookieName())
	c.ClearCookie(GetIdpSessionCookieName())
	c.TemporaryRedirect("/logout")
}

func (c *RouteContext) GetRedirectURL(defaultRedirect string) string {
	candidates := []string{
		"redirect_uri", // most common, OAuth standard
		"redirect_url",
		"return_url",
		"returnUrl",
		"return_to",
		"redirect",
		"return",
	}

	for _, key := range candidates {
		if url, ok := c.QueryValue(key); ok {
			return url
		}
	}

	return defaultRedirect
}

func addToStaging(staging map[string]any, key string, values []string) {
	if len(values) == 1 {
		staging[key] = values[0]
	} else {
		staging[key] = values
	}
}

func getInstanceURI(r *http.Request) *string {
	instanceURI := r.RequestURI
	return &instanceURI
}
