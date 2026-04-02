package mux

import (
	"context"
	"net/http"
	"time"

	"github.com/fgrzl/claims"
	internalcommon "github.com/fgrzl/mux/internal/common"
	internalcookiekit "github.com/fgrzl/mux/internal/cookiekit"
	internalauthentication "github.com/fgrzl/mux/internal/middleware/authentication"
	internalrouting "github.com/fgrzl/mux/internal/routing"
	internaltokenizer "github.com/fgrzl/mux/internal/tokenizer"
	"github.com/google/uuid"
)

type ServiceKey string

type HandlerFunc func(RouteContext)

type MutableRouteContext interface {
	RouteContext
	SetRequest(*http.Request)
	SetResponse(http.ResponseWriter)
	SetUser(claims.Principal)
	SetContextValue(key, value any)
}

type MiddlewareFunc func(MutableRouteContext, HandlerFunc)

func (f MiddlewareFunc) Invoke(c MutableRouteContext, next HandlerFunc) {
	f(c, next)
}

type Middleware interface {
	Invoke(MutableRouteContext, HandlerFunc)
}

type TokenProvider interface {
	CreateToken(ctx context.Context, principal claims.Principal) (string, error)
	ValidateToken(ctx context.Context, token string) (claims.Principal, error)
	GetTTL() time.Duration
	CanCreateTokens() bool
}

type ProblemDetails struct {
	Type     string  `json:"type"`
	Title    string  `json:"title"`
	Status   int     `json:"status"`
	Detail   string  `json:"detail"`
	Instance *string `json:"instance,omitempty"`
}

var DefaultProblem = &ProblemDetails{}

const ServiceKeyTokenProvider = ServiceKey(internaltokenizer.ServiceKeyTokenProvider)

const (
	MimeJSON        = internalcommon.MimeJSON
	MimeOpenAPI     = internalcommon.MimeOpenAPI
	MimeYAML        = internalcommon.MimeYAML
	MimeProblemJSON = internalcommon.MimeProblemJSON
)

const (
	HeaderAccept        = internalcommon.HeaderAccept
	HeaderAuthorization = internalcommon.HeaderAuthorization
	HeaderContentType   = internalcommon.HeaderContentType
	HeaderLocation      = internalcommon.HeaderLocation
	HeaderRetryAfter    = internalcommon.HeaderRetryAfter
)

type CookieOption struct {
	apply internalcookiekit.CookieOption
}

func WithCookieMaxAge(maxAge int) CookieOption {
	return CookieOption{apply: internalcookiekit.WithMaxAge(maxAge)}
}

func WithCookiePath(path string) CookieOption {
	return CookieOption{apply: internalcookiekit.WithPath(path)}
}

func WithCookieDomain(domain string) CookieOption {
	return CookieOption{apply: internalcookiekit.WithDomain(domain)}
}

func WithCookieSecure(secure bool) CookieOption {
	return CookieOption{apply: internalcookiekit.WithSecure(secure)}
}

func WithCookieHTTPOnly(httpOnly bool) CookieOption {
	return CookieOption{apply: internalcookiekit.WithHttpOnly(httpOnly)}
}

func WithCookieSameSite(sameSite http.SameSite) CookieOption {
	return CookieOption{apply: internalcookiekit.WithSameSite(sameSite)}
}

func toInternalCookieOptions(opts []CookieOption) []internalcookiekit.CookieOption {
	if len(opts) == 0 {
		return nil
	}
	internal := make([]internalcookiekit.CookieOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internal = append(internal, opt.apply)
		}
	}
	return internal
}

type RouteContext interface {
	context.Context
	Request() *http.Request
	Response() http.ResponseWriter
	Bind(any) error
	User() claims.Principal
	Services() *ServiceRegistry
	Param(name string) (string, bool)
	ParamUUID(name string) (uuid.UUID, bool)
	ParamInt(name string) (int, bool)
	ParamInt16(name string) (int16, bool)
	ParamInt32(name string) (int32, bool)
	ParamInt64(name string) (int64, bool)
	Query() *QueryAccessor
	Form() *FormAccessor
	Headers() *HeaderAccessor
	Cookies() *CookieAccessor
	JSON(status int, model any)
	Plain(status int, data []byte)
	HTML(status int, html string)
	OK(model any)
	Created(model any)
	Accepted(model any)
	NoContent()
	NotFound()
	BadRequest(title, detail string)
	Unauthorized()
	Forbidden(message string)
	Conflict(title, detail string)
	ServerError(title, detail string)
	Problem(detail *ProblemDetails)
	File(path string)
	Download(path, filename string)
	Redirect(status int, url string)
	MovedPermanently(url string)
	Found(url string)
	SeeOther(url string)
	TemporaryRedirect(url string)
	PermanentRedirect(url string)
}

type routeContext struct {
	inner internalrouting.RouteContext
}

func wrapRouteContext(inner internalrouting.RouteContext) *routeContext {
	if inner == nil {
		return nil
	}
	return &routeContext{inner: inner}
}

func unwrapRouteContext(c RouteContext) internalrouting.RouteContext {
	if c == nil {
		return nil
	}
	if wrapped, ok := c.(*routeContext); ok {
		return wrapped.inner
	}
	return nil
}

func NewRouteContext(w http.ResponseWriter, r *http.Request) MutableRouteContext {
	return wrapRouteContext(internalrouting.NewRouteContext(w, r))
}

func RouteContextFromRequest(r *http.Request) (RouteContext, bool) {
	inner, ok := internalrouting.RouteContextFromRequest(r)
	if !ok {
		return nil, false
	}
	return wrapRouteContext(inner), true
}

func ClearCookieWithOptions(c RouteContext, name string, opts ...CookieOption) {
	internalrouting.ClearCookieWithOptions(unwrapRouteContext(c), name, toInternalCookieOptions(opts)...)
}

func SignOutWithOptions(c RouteContext, redirectURL string, opts ...CookieOption) {
	internalrouting.SignOutWithOptions(unwrapRouteContext(c), redirectURL, toInternalCookieOptions(opts)...)
}

func (c *routeContext) Deadline() (time.Time, bool)       { return c.inner.Deadline() }
func (c *routeContext) Done() <-chan struct{}             { return c.inner.Done() }
func (c *routeContext) Err() error                        { return c.inner.Err() }
func (c *routeContext) Value(key any) any                 { return c.inner.Value(key) }
func (c *routeContext) Request() *http.Request            { return c.inner.Request() }
func (c *routeContext) SetRequest(r *http.Request)        { c.inner.SetRequest(r) }
func (c *routeContext) Response() http.ResponseWriter     { return c.inner.Response() }
func (c *routeContext) SetResponse(w http.ResponseWriter) { c.inner.SetResponse(w) }
func (c *routeContext) Bind(target any) error             { return c.inner.Bind(target) }
func (c *routeContext) User() claims.Principal            { return c.inner.User() }
func (c *routeContext) SetUser(user claims.Principal)     { c.inner.SetUser(user) }
func (c *routeContext) SetContextValue(key, value any)    { c.inner.SetContextValue(key, value) }
func (c *routeContext) Services() *ServiceRegistry {
	return newServiceRegistry(
		func(key ServiceKey, svc any) {
			c.inner.SetService(internalcommon.ServiceKey(key), svc)
		},
		func(key ServiceKey) (any, bool) {
			return c.inner.GetService(internalcommon.ServiceKey(key))
		},
	)
}
func (c *routeContext) Param(name string) (string, bool) { return c.inner.Param(name) }
func (c *routeContext) ParamUUID(name string) (uuid.UUID, bool) {
	return c.inner.ParamUUID(name)
}
func (c *routeContext) ParamInt(name string) (int, bool)     { return c.inner.ParamInt(name) }
func (c *routeContext) ParamInt16(name string) (int16, bool) { return c.inner.ParamInt16(name) }
func (c *routeContext) ParamInt32(name string) (int32, bool) { return c.inner.ParamInt32(name) }
func (c *routeContext) ParamInt64(name string) (int64, bool) { return c.inner.ParamInt64(name) }
func (c *routeContext) Query() *QueryAccessor                { return &QueryAccessor{ctx: c} }
func (c *routeContext) Form() *FormAccessor                  { return &FormAccessor{ctx: c} }
func (c *routeContext) Headers() *HeaderAccessor             { return &HeaderAccessor{ctx: c} }
func (c *routeContext) Cookies() *CookieAccessor             { return &CookieAccessor{ctx: c} }
func (c *routeContext) JSON(status int, model any)           { c.inner.JSON(status, model) }
func (c *routeContext) Plain(status int, data []byte)        { c.inner.Plain(status, data) }
func (c *routeContext) HTML(status int, html string)         { c.inner.HTML(status, html) }
func (c *routeContext) OK(model any)                         { c.inner.OK(model) }
func (c *routeContext) Created(model any)                    { c.inner.Created(model) }
func (c *routeContext) Accepted(model any)                   { c.inner.Accept(model) }
func (c *routeContext) NoContent()                           { c.inner.NoContent() }
func (c *routeContext) NotFound()                            { c.inner.NotFound() }
func (c *routeContext) BadRequest(title, detail string)      { c.inner.BadRequest(title, detail) }
func (c *routeContext) Unauthorized()                        { c.inner.Unauthorized() }
func (c *routeContext) Forbidden(message string)             { c.inner.Forbidden(message) }
func (c *routeContext) Conflict(title, detail string)        { c.inner.Conflict(title, detail) }
func (c *routeContext) ServerError(title, detail string)     { c.inner.ServerError(title, detail) }
func (c *routeContext) Problem(detail *ProblemDetails) {
	if detail == nil {
		c.inner.Problem(nil)
		return
	}
	c.inner.Problem(&internalrouting.ProblemDetails{
		Type:     detail.Type,
		Title:    detail.Title,
		Status:   detail.Status,
		Detail:   detail.Detail,
		Instance: detail.Instance,
	})
}
func (c *routeContext) File(path string)                { c.inner.File(path) }
func (c *routeContext) Download(path, filename string)  { c.inner.Download(path, filename) }
func (c *routeContext) Redirect(status int, url string) { c.inner.Redirect(status, url) }
func (c *routeContext) MovedPermanently(url string)     { c.inner.MovedPermanently(url) }
func (c *routeContext) Found(url string)                { c.inner.Found(url) }
func (c *routeContext) SeeOther(url string)             { c.inner.SeeOther(url) }
func (c *routeContext) TemporaryRedirect(url string)    { c.inner.TemporaryRedirect(url) }
func (c *routeContext) PermanentRedirect(url string)    { c.inner.PermanentRedirect(url) }

type ServiceRegistry struct {
	register func(ServiceKey, any)
	lookup   func(ServiceKey) (any, bool)
}

func newServiceRegistry(register func(ServiceKey, any), lookup func(ServiceKey) (any, bool)) *ServiceRegistry {
	return &ServiceRegistry{register: register, lookup: lookup}
}

func (r *ServiceRegistry) Register(key ServiceKey, svc any) *ServiceRegistry {
	if r == nil || r.register == nil {
		return r
	}
	r.register(key, svc)
	return r
}

func (r *ServiceRegistry) Get(key ServiceKey) (any, bool) {
	if r == nil || r.lookup == nil {
		return nil, false
	}
	return r.lookup(key)
}

type QueryAccessor struct{ ctx *routeContext }

func (q *QueryAccessor) String(name string) (string, bool)     { return q.ctx.inner.QueryValue(name) }
func (q *QueryAccessor) Strings(name string) ([]string, bool)  { return q.ctx.inner.QueryValues(name) }
func (q *QueryAccessor) UUID(name string) (uuid.UUID, bool)    { return q.ctx.inner.QueryUUID(name) }
func (q *QueryAccessor) UUIDs(name string) ([]uuid.UUID, bool) { return q.ctx.inner.QueryUUIDs(name) }
func (q *QueryAccessor) Int(name string) (int, bool)           { return q.ctx.inner.QueryInt(name) }
func (q *QueryAccessor) Ints(name string) ([]int, bool)        { return q.ctx.inner.QueryInts(name) }
func (q *QueryAccessor) Int16(name string) (int16, bool)       { return q.ctx.inner.QueryInt16(name) }
func (q *QueryAccessor) Int16s(name string) ([]int16, bool)    { return q.ctx.inner.QueryInt16s(name) }
func (q *QueryAccessor) Int32(name string) (int32, bool)       { return q.ctx.inner.QueryInt32(name) }
func (q *QueryAccessor) Int32s(name string) ([]int32, bool)    { return q.ctx.inner.QueryInt32s(name) }
func (q *QueryAccessor) Int64(name string) (int64, bool)       { return q.ctx.inner.QueryInt64(name) }
func (q *QueryAccessor) Int64s(name string) ([]int64, bool)    { return q.ctx.inner.QueryInt64s(name) }
func (q *QueryAccessor) Bool(name string) (bool, bool)         { return q.ctx.inner.QueryBool(name) }
func (q *QueryAccessor) Bools(name string) ([]bool, bool)      { return q.ctx.inner.QueryBools(name) }
func (q *QueryAccessor) Float32(name string) (float32, bool)   { return q.ctx.inner.QueryFloat32(name) }
func (q *QueryAccessor) Float32s(name string) ([]float32, bool) {
	return q.ctx.inner.QueryFloat32s(name)
}
func (q *QueryAccessor) Float64(name string) (float64, bool) { return q.ctx.inner.QueryFloat64(name) }
func (q *QueryAccessor) Float64s(name string) ([]float64, bool) {
	return q.ctx.inner.QueryFloat64s(name)
}

type FormAccessor struct{ ctx *routeContext }

func (f *FormAccessor) String(name string) (string, bool)     { return f.ctx.inner.FormValue(name) }
func (f *FormAccessor) Strings(name string) ([]string, bool)  { return f.ctx.inner.FormValues(name) }
func (f *FormAccessor) UUID(name string) (uuid.UUID, bool)    { return f.ctx.inner.FormUUID(name) }
func (f *FormAccessor) UUIDs(name string) ([]uuid.UUID, bool) { return f.ctx.inner.FormUUIDs(name) }
func (f *FormAccessor) Int(name string) (int, bool)           { return f.ctx.inner.FormInt(name) }
func (f *FormAccessor) Ints(name string) ([]int, bool)        { return f.ctx.inner.FormInts(name) }
func (f *FormAccessor) Int16(name string) (int16, bool)       { return f.ctx.inner.FormInt16(name) }
func (f *FormAccessor) Int16s(name string) ([]int16, bool)    { return f.ctx.inner.FormInt16s(name) }
func (f *FormAccessor) Int32(name string) (int32, bool)       { return f.ctx.inner.FormInt32(name) }
func (f *FormAccessor) Int32s(name string) ([]int32, bool)    { return f.ctx.inner.FormInt32s(name) }
func (f *FormAccessor) Int64(name string) (int64, bool)       { return f.ctx.inner.FormInt64(name) }
func (f *FormAccessor) Int64s(name string) ([]int64, bool)    { return f.ctx.inner.FormInt64s(name) }
func (f *FormAccessor) Bool(name string) (bool, bool)         { return f.ctx.inner.FormBool(name) }
func (f *FormAccessor) Bools(name string) ([]bool, bool)      { return f.ctx.inner.FormBools(name) }
func (f *FormAccessor) Float32(name string) (float32, bool)   { return f.ctx.inner.FormFloat32(name) }
func (f *FormAccessor) Float32s(name string) ([]float32, bool) {
	return f.ctx.inner.FormFloat32s(name)
}
func (f *FormAccessor) Float64(name string) (float64, bool) { return f.ctx.inner.FormFloat64(name) }
func (f *FormAccessor) Float64s(name string) ([]float64, bool) {
	return f.ctx.inner.FormFloat64s(name)
}

type HeaderAccessor struct{ ctx *routeContext }

func (h *HeaderAccessor) String(name string) (string, bool)   { return h.ctx.inner.Header(name) }
func (h *HeaderAccessor) Int(name string) (int, bool)         { return h.ctx.inner.HeaderInt(name) }
func (h *HeaderAccessor) UUID(name string) (uuid.UUID, bool)  { return h.ctx.inner.HeaderUUID(name) }
func (h *HeaderAccessor) Bool(name string) (bool, bool)       { return h.ctx.inner.HeaderBool(name) }
func (h *HeaderAccessor) Float64(name string) (float64, bool) { return h.ctx.inner.HeaderFloat64(name) }

type CookieAccessor struct{ ctx *routeContext }

func (c *CookieAccessor) Get(name string) (string, error) {
	return c.ctx.inner.GetCookie(name)
}

func (c *CookieAccessor) Set(name, value string, maxAge int, path, domain string, secure, httpOnly bool, sameSite ...http.SameSite) {
	c.ctx.inner.SetCookie(name, value, maxAge, path, domain, secure, httpOnly, sameSite...)
}

func (c *CookieAccessor) Clear(name string) {
	c.ctx.inner.ClearCookie(name)
}

func (c *CookieAccessor) Authenticate(cookieName string, user claims.Principal, opts ...CookieOption) {
	c.ctx.inner.Authenticate(cookieName, user, toInternalCookieOptions(opts)...)
}

func (c *CookieAccessor) SignIn(user claims.Principal, redirectURL string, opts ...CookieOption) {
	c.ctx.inner.SignIn(user, redirectURL, toInternalCookieOptions(opts)...)
}

func (c *CookieAccessor) SignOut(redirectURL string) {
	c.ctx.inner.SignOut(redirectURL)
}

func (c *CookieAccessor) CSRFToken() string {
	return internalauthentication.GenerateCSRFToken(c.ctx.inner)
}

func (c *CookieAccessor) CSRFTokenErr() (string, error) {
	return internalauthentication.GenerateCSRFTokenErr(c.ctx.inner)
}
