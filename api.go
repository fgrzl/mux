package mux

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fgrzl/claims"
	internalbinder "github.com/fgrzl/mux/internal/binder"
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
	Params() *ParamAccessor
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
func (c *routeContext) Params() *ParamAccessor           { return &ParamAccessor{ctx: c} }
func (c *routeContext) Query() *QueryAccessor            { return &QueryAccessor{ctx: c} }
func (c *routeContext) Form() *FormAccessor              { return &FormAccessor{ctx: c} }
func (c *routeContext) Headers() *HeaderAccessor         { return &HeaderAccessor{ctx: c} }
func (c *routeContext) Cookies() *CookieAccessor         { return &CookieAccessor{ctx: c} }
func (c *routeContext) JSON(status int, model any)       { c.inner.JSON(status, model) }
func (c *routeContext) Plain(status int, data []byte)    { c.inner.Plain(status, data) }
func (c *routeContext) HTML(status int, html string)     { c.inner.HTML(status, html) }
func (c *routeContext) OK(model any)                     { c.inner.OK(model) }
func (c *routeContext) Created(model any)                { c.inner.Created(model) }
func (c *routeContext) Accepted(model any)               { c.inner.Accept(model) }
func (c *routeContext) NoContent()                       { c.inner.NoContent() }
func (c *routeContext) NotFound()                        { c.inner.NotFound() }
func (c *routeContext) BadRequest(title, detail string)  { c.inner.BadRequest(title, detail) }
func (c *routeContext) Unauthorized()                    { c.inner.Unauthorized() }
func (c *routeContext) Forbidden(message string)         { c.inner.Forbidden(message) }
func (c *routeContext) Conflict(title, detail string)    { c.inner.Conflict(title, detail) }
func (c *routeContext) ServerError(title, detail string) { c.inner.ServerError(title, detail) }
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

type ParamAccessor struct{ ctx *routeContext }

func (p *ParamAccessor) String(name string) (string, bool)  { return p.ctx.inner.Param(name) }
func (p *ParamAccessor) UUID(name string) (uuid.UUID, bool) { return p.ctx.inner.ParamUUID(name) }
func (p *ParamAccessor) Int(name string) (int, bool)        { return p.ctx.inner.ParamInt(name) }
func (p *ParamAccessor) Int16(name string) (int16, bool)    { return p.ctx.inner.ParamInt16(name) }
func (p *ParamAccessor) Int32(name string) (int32, bool)    { return p.ctx.inner.ParamInt32(name) }
func (p *ParamAccessor) Int64(name string) (int64, bool)    { return p.ctx.inner.ParamInt64(name) }

type QueryAccessor struct{ ctx *routeContext }

func (q *QueryAccessor) String(name string) (string, bool) { return queryValue(q.ctx.Request(), name) }
func (q *QueryAccessor) Strings(name string) ([]string, bool) {
	return queryValues(q.ctx.Request(), name)
}
func (q *QueryAccessor) UUID(name string) (uuid.UUID, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return uuid.Nil, false
	}
	return internalbinder.ParseUUIDVal(value)
}
func (q *QueryAccessor) UUIDs(name string) ([]uuid.UUID, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseUUIDSlice(values)
}
func (q *QueryAccessor) Int(name string) (int, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseIntVal(value)
}
func (q *QueryAccessor) Ints(name string) ([]int, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseIntSlice(values)
}
func (q *QueryAccessor) Int16(name string) (int16, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseInt16Val(value)
}
func (q *QueryAccessor) Int16s(name string) ([]int16, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseInt16Slice(values)
}
func (q *QueryAccessor) Int32(name string) (int32, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseInt32Val(value)
}
func (q *QueryAccessor) Int32s(name string) ([]int32, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseInt32Slice(values)
}
func (q *QueryAccessor) Int64(name string) (int64, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseInt64Val(value)
}
func (q *QueryAccessor) Int64s(name string) ([]int64, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseInt64Slice(values)
}
func (q *QueryAccessor) Bool(name string) (bool, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return false, false
	}
	return internalbinder.ParseBoolVal(value)
}
func (q *QueryAccessor) Bools(name string) ([]bool, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseBoolSlice(values)
}
func (q *QueryAccessor) Float32(name string) (float32, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseFloat32Val(value)
}
func (q *QueryAccessor) Float32s(name string) ([]float32, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseFloat32Slice(values)
}
func (q *QueryAccessor) Float64(name string) (float64, bool) {
	value, ok := queryValue(q.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseFloat64Val(value)
}
func (q *QueryAccessor) Float64s(name string) ([]float64, bool) {
	values, ok := queryValues(q.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseFloat64Slice(values)
}

func (q *QueryAccessor) GetRedirectURL(defaultRedirect string) string {
	for _, key := range []string{"redirect_uri", "redirect_url", "return_url", "returnUrl", "return_to", "redirect", "return"} {
		value, ok := queryValue(q.ctx.Request(), key)
		if !ok {
			continue
		}
		if redirectURL, ok := safeRedirectTarget(value, q.ctx.Request(), routeClientURL(q.ctx.inner)); ok {
			return redirectURL
		}
	}
	return defaultRedirect
}

func queryValue(r *http.Request, name string) (string, bool) {
	values, ok := queryValues(r, name)
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func queryValues(r *http.Request, name string) ([]string, bool) {
	if r == nil || r.URL == nil {
		return nil, false
	}
	values, ok := r.URL.Query()[name]
	return values, ok
}

func safeRedirectTarget(raw string, r *http.Request, clientURL *url.URL) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if strings.HasPrefix(raw, "//") || strings.HasPrefix(raw, `\\`) {
		return "", false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	if parsed.Scheme != "" || parsed.Host != "" {
		if !isSameOriginRedirect(parsed, r, clientURL) {
			return "", false
		}
		return redirectPathFromURL(parsed), true
	}

	if !strings.HasPrefix(raw, "/") {
		raw = "/" + raw
	}
	if len(raw) > 1 && (raw[1] == '/' || raw[1] == '\\') {
		return "", false
	}
	return raw, true
}

func isSameOriginRedirect(target *url.URL, r *http.Request, clientURL *url.URL) bool {
	if target == nil {
		return false
	}
	expectedHost := ""
	if clientURL != nil {
		expectedHost = clientURL.Host
	}
	if expectedHost == "" && r != nil {
		expectedHost = r.Host
	}
	return expectedHost != "" && strings.EqualFold(target.Host, expectedHost)
}

func redirectPathFromURL(target *url.URL) string {
	path := target.EscapedPath()
	if path == "" {
		path = "/"
	} else if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if target.RawQuery != "" {
		path += "?" + target.RawQuery
	}
	if target.Fragment != "" {
		path += "#" + target.Fragment
	}
	return path
}

func routeClientURL(inner internalrouting.RouteContext) *url.URL {
	if inner == nil {
		return nil
	}
	provider, ok := any(inner).(interface{ ClientURL() *url.URL })
	if !ok {
		return nil
	}
	return provider.ClientURL()
}

type FormAccessor struct{ ctx *routeContext }

func (f *FormAccessor) String(name string) (string, bool) { return formValue(f.ctx.Request(), name) }
func (f *FormAccessor) Strings(name string) ([]string, bool) {
	return formValues(f.ctx.Request(), name)
}
func (f *FormAccessor) UUID(name string) (uuid.UUID, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return uuid.Nil, false
	}
	return internalbinder.ParseUUIDVal(value)
}
func (f *FormAccessor) UUIDs(name string) ([]uuid.UUID, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseUUIDSlice(values)
}
func (f *FormAccessor) Int(name string) (int, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseIntVal(value)
}
func (f *FormAccessor) Ints(name string) ([]int, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseIntSlice(values)
}
func (f *FormAccessor) Int16(name string) (int16, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseInt16Val(value)
}
func (f *FormAccessor) Int16s(name string) ([]int16, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseInt16Slice(values)
}
func (f *FormAccessor) Int32(name string) (int32, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseInt32Val(value)
}
func (f *FormAccessor) Int32s(name string) ([]int32, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseInt32Slice(values)
}
func (f *FormAccessor) Int64(name string) (int64, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseInt64Val(value)
}
func (f *FormAccessor) Int64s(name string) ([]int64, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseInt64Slice(values)
}
func (f *FormAccessor) Bool(name string) (bool, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return false, false
	}
	return internalbinder.ParseBoolVal(value)
}
func (f *FormAccessor) Bools(name string) ([]bool, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseBoolSlice(values)
}
func (f *FormAccessor) Float32(name string) (float32, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseFloat32Val(value)
}
func (f *FormAccessor) Float32s(name string) ([]float32, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseFloat32Slice(values)
}
func (f *FormAccessor) Float64(name string) (float64, bool) {
	value, ok := formValue(f.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseFloat64Val(value)
}
func (f *FormAccessor) Float64s(name string) ([]float64, bool) {
	values, ok := formValues(f.ctx.Request(), name)
	if !ok {
		return nil, false
	}
	return internalbinder.ParseFloat64Slice(values)
}

func formValue(r *http.Request, name string) (string, bool) {
	values, ok := formValues(r, name)
	if !ok || len(values) == 0 {
		return "", false
	}
	return values[0], true
}

func formValues(r *http.Request, name string) ([]string, bool) {
	if err := ensureFormsParsed(r); err != nil {
		return nil, false
	}
	if r == nil || r.Form == nil {
		return nil, false
	}
	values, ok := r.Form[name]
	return values, ok
}

func ensureFormsParsed(r *http.Request) error {
	if r == nil {
		return nil
	}
	contentType := r.Header.Get(HeaderContentType)
	if strings.HasPrefix(contentType, internalcommon.MimeMultipartFormData) {
		return r.ParseMultipartForm(32 << 20)
	}
	if strings.HasPrefix(contentType, internalcommon.MimeFormURLEncoded) {
		return r.ParseForm()
	}
	return nil
}

type HeaderAccessor struct{ ctx *routeContext }

func (h *HeaderAccessor) String(name string) (string, bool) {
	return headerValue(h.ctx.Request(), name)
}
func (h *HeaderAccessor) Int(name string) (int, bool) {
	value, ok := headerValue(h.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseIntVal(value)
}
func (h *HeaderAccessor) UUID(name string) (uuid.UUID, bool) {
	value, ok := headerValue(h.ctx.Request(), name)
	if !ok {
		return uuid.Nil, false
	}
	return internalbinder.ParseUUIDVal(value)
}
func (h *HeaderAccessor) Bool(name string) (bool, bool) {
	value, ok := headerValue(h.ctx.Request(), name)
	if !ok {
		return false, false
	}
	return internalbinder.ParseBoolVal(value)
}
func (h *HeaderAccessor) Float64(name string) (float64, bool) {
	value, ok := headerValue(h.ctx.Request(), name)
	if !ok {
		return 0, false
	}
	return internalbinder.ParseFloat64Val(value)
}

func headerValue(r *http.Request, name string) (string, bool) {
	if r == nil {
		return "", false
	}
	value := r.Header.Get(name)
	return value, value != ""
}

type CookieAccessor struct{ ctx *routeContext }

func (c *CookieAccessor) Get(name string) (string, error) {
	if c == nil || c.ctx == nil || c.ctx.Request() == nil {
		return "", http.ErrNoCookie
	}
	cookie, err := c.ctx.Request().Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
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
