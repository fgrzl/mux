package mux

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fgrzl/claims"
	"github.com/google/uuid"
)

// MockRouteContext implements RouteContextInterface for testing
type MockRouteContext struct {
	context.Context
	user     claims.Principal
	services map[ServiceKey]any
	bindings map[string]any
	response []byte
	status   int
	request  *http.Request
}

func NewMockRouteContext() *MockRouteContext {
	return &MockRouteContext{
		Context:  context.Background(),
		services: make(map[ServiceKey]any),
		bindings: make(map[string]any),
		status:   200,
		request:  &http.Request{},
	}
}

// Core context methods
func (m *MockRouteContext) SetUser(user claims.Principal) {
	m.user = user
}

func (m *MockRouteContext) User() claims.Principal {
	return m.user
}

func (m *MockRouteContext) SetService(key ServiceKey, service any) {
	m.services[key] = service
}

func (m *MockRouteContext) GetService(key ServiceKey) (any, bool) {
	svc, ok := m.services[key]
	return svc, ok
}

// Request binding
func (m *MockRouteContext) Bind(target any) error {
	return nil // Mock implementation
}

// Response methods - Basic HTTP responses
func (m *MockRouteContext) OK(model any) {
	m.status = 200
	m.response = []byte(fmt.Sprintf("%v", model))
}

func (m *MockRouteContext) JSON(status int, model any) {
	m.status = status
	m.response = []byte(fmt.Sprintf(`{"data": "%v"}`, model))
}

func (m *MockRouteContext) Plain(status int, data []byte) {
	m.status = status
	m.response = data
}

func (m *MockRouteContext) HTML(status int, html string) {
	m.status = status
	m.response = []byte(html)
}

func (m *MockRouteContext) NoContent() {
	m.status = 204
	m.response = nil
}

func (m *MockRouteContext) NotFound() {
	m.status = 404
	m.response = []byte("Not Found")
}

func (m *MockRouteContext) Created(model any) {
	m.status = 201
	m.response = []byte(fmt.Sprintf(`{"data": "%v"}`, model))
}

func (m *MockRouteContext) Accept(model any) {
	m.status = 202
	m.response = []byte(fmt.Sprintf(`{"data": "%v"}`, model))
}

// Response methods - Error responses
func (m *MockRouteContext) BadRequest(title, detail string) {
	m.status = 400
	m.response = []byte(fmt.Sprintf(`{"title": "%s", "detail": "%s"}`, title, detail))
}

func (m *MockRouteContext) Unauthorized() {
	m.status = 401
	m.response = []byte("Unauthorized")
}

func (m *MockRouteContext) Forbidden(message string) {
	m.status = 403
	m.response = []byte(message)
}

func (m *MockRouteContext) Conflict(title, detail string) {
	m.status = 409
	m.response = []byte(fmt.Sprintf(`{"title": "%s", "detail": "%s"}`, title, detail))
}

func (m *MockRouteContext) ServerError(title, detail string) {
	m.status = 500
	m.response = []byte(fmt.Sprintf(`{"title": "%s", "detail": "%s"}`, title, detail))
}

func (m *MockRouteContext) Problem(detail *ProblemDetails) {
	m.status = detail.Status
	m.response = []byte(fmt.Sprintf(`{"title": "%s", "detail": "%s", "status": %d}`,
		detail.Title, detail.Detail, detail.Status))
}

// Response methods - File and redirects (simplified for demo)
func (m *MockRouteContext) File(filePath string) {
	m.status = 200
	m.response = []byte(fmt.Sprintf("File: %s", filePath))
}

func (m *MockRouteContext) Download(filePath string, filename string) {
	m.status = 200
	m.response = []byte(fmt.Sprintf("Download: %s as %s", filePath, filename))
}

func (m *MockRouteContext) Redirect(status int, url string) {
	m.status = status
	m.response = []byte(fmt.Sprintf("Redirect to: %s", url))
}

func (m *MockRouteContext) TemporaryRedirect(url string) {
	m.status = 307
	m.response = []byte(fmt.Sprintf("Temporary redirect to: %s", url))
}

func (m *MockRouteContext) PermanentRedirect(url string) {
	m.status = 301
	m.response = []byte(fmt.Sprintf("Permanent redirect to: %s", url))
}

// Simplified parameter/query/form/header/cookie methods for demo
func (m *MockRouteContext) Param(name string) (string, bool)             { return "", false }
func (m *MockRouteContext) ParamUUID(name string) (uuid.UUID, bool)      { return uuid.UUID{}, false }
func (m *MockRouteContext) ParamInt(name string) (int, bool)             { return 0, false }
func (m *MockRouteContext) ParamInt16(name string) (int16, bool)         { return 0, false }
func (m *MockRouteContext) ParamInt32(name string) (int32, bool)         { return 0, false }
func (m *MockRouteContext) ParamInt64(name string) (int64, bool)         { return 0, false }
func (m *MockRouteContext) QueryValue(name string) (string, bool)        { return "", false }
func (m *MockRouteContext) QueryValues(name string) ([]string, bool)     { return nil, false }
func (m *MockRouteContext) QueryUUID(name string) (uuid.UUID, bool)      { return uuid.UUID{}, false }
func (m *MockRouteContext) QueryUUIDs(name string) ([]uuid.UUID, bool)   { return nil, false }
func (m *MockRouteContext) QueryInt(name string) (int, bool)             { return 0, false }
func (m *MockRouteContext) QueryInts(name string) ([]int, bool)          { return nil, false }
func (m *MockRouteContext) QueryInt16(name string) (int16, bool)         { return 0, false }
func (m *MockRouteContext) QueryInt16s(name string) ([]int16, bool)      { return nil, false }
func (m *MockRouteContext) QueryInt32(name string) (int32, bool)         { return 0, false }
func (m *MockRouteContext) QueryInt32s(name string) ([]int32, bool)      { return nil, false }
func (m *MockRouteContext) QueryInt64(name string) (int64, bool)         { return 0, false }
func (m *MockRouteContext) QueryInt64s(name string) ([]int64, bool)      { return nil, false }
func (m *MockRouteContext) QueryBool(name string) (bool, bool)           { return false, false }
func (m *MockRouteContext) QueryBools(name string) ([]bool, bool)        { return nil, false }
func (m *MockRouteContext) QueryFloat32(name string) (float32, bool)     { return 0, false }
func (m *MockRouteContext) QueryFloat32s(name string) ([]float32, bool)  { return nil, false }
func (m *MockRouteContext) QueryFloat64(name string) (float64, bool)     { return 0, false }
func (m *MockRouteContext) QueryFloat64s(name string) ([]float64, bool)  { return nil, false }
func (m *MockRouteContext) GetRedirectURL(defaultRedirect string) string { return defaultRedirect }
func (m *MockRouteContext) FormValue(name string) (string, bool)         { return "", false }
func (m *MockRouteContext) FormValues(name string) ([]string, bool)      { return nil, false }
func (m *MockRouteContext) FormUUID(name string) (uuid.UUID, bool)       { return uuid.UUID{}, false }
func (m *MockRouteContext) FormUUIDs(name string) ([]uuid.UUID, bool)    { return nil, false }
func (m *MockRouteContext) FormInt(name string) (int, bool)              { return 0, false }
func (m *MockRouteContext) FormInts(name string) ([]int, bool)           { return nil, false }
func (m *MockRouteContext) FormInt16(name string) (int16, bool)          { return 0, false }
func (m *MockRouteContext) FormInt16s(name string) ([]int16, bool)       { return nil, false }
func (m *MockRouteContext) FormInt32(name string) (int32, bool)          { return 0, false }
func (m *MockRouteContext) FormInt32s(name string) ([]int32, bool)       { return nil, false }
func (m *MockRouteContext) FormInt64(name string) (int64, bool)          { return 0, false }
func (m *MockRouteContext) FormInt64s(name string) ([]int64, bool)       { return nil, false }
func (m *MockRouteContext) FormBool(name string) (bool, bool)            { return false, false }
func (m *MockRouteContext) FormBools(name string) ([]bool, bool)         { return nil, false }
func (m *MockRouteContext) FormFloat32(name string) (float32, bool)      { return 0, false }
func (m *MockRouteContext) FormFloat32s(name string) ([]float32, bool)   { return nil, false }
func (m *MockRouteContext) FormFloat64(name string) (float64, bool)      { return 0, false }
func (m *MockRouteContext) FormFloat64s(name string) ([]float64, bool)   { return nil, false }
func (m *MockRouteContext) Header(name string) (string, bool)            { return "", false }
func (m *MockRouteContext) HeaderInt(name string) (int, bool)            { return 0, false }
func (m *MockRouteContext) HeaderUUID(name string) (uuid.UUID, bool)     { return uuid.UUID{}, false }
func (m *MockRouteContext) HeaderBool(name string) (bool, bool)          { return false, false }
func (m *MockRouteContext) HeaderFloat64(name string) (float64, bool)    { return 0, false }
func (m *MockRouteContext) GetCookie(name string) (string, error)        { return "", http.ErrNoCookie }
func (m *MockRouteContext) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
}
func (m *MockRouteContext) ClearCookie(name string)               {}
func (m *MockRouteContext) SignOut()                              {}
func (m *MockRouteContext) Authenticate(string, claims.Principal) {}

// Request returns the underlying *http.Request for the mock context.
func (m *MockRouteContext) Request() *http.Request {
	return m.request
}

func (m *MockRouteContext) Response() http.ResponseWriter {
	return nil
}

func (m *MockRouteContext) SetResponse(http.ResponseWriter) {

}

func (m *MockRouteContext) Options() *RouteOptions {
	return nil
}

func (m *MockRouteContext) Params() RouteParams {
	return nil
}

func (m *MockRouteContext) SignIn(user claims.Principal, redirectUrl string) {}

// Test the handler with the mock
func TestHandlerWithMock(handler HandlerFunc) {
	mock := NewMockRouteContext()
	handler(mock)
	fmt.Printf("Status: %d, Response: %s\n", mock.status, string(mock.response))
}

func ExampleMockUsage() {
	fmt.Println("Testing RouteContextInterface with Mock Implementation")

	// Example handler that uses the interface
	exampleHandler := func(c RouteContext) {
		c.SetService("test", "test-service")
		service, ok := c.GetService("test")
		if ok {
			fmt.Printf("Service retrieved: %v\n", service)
		}

		c.JSON(200, map[string]string{"message": "Hello, World!"})
	}

	TestHandlerWithMock(exampleHandler)

	// Test error response
	errorHandler := func(c RouteContext) {
		c.BadRequest("Invalid Request", "The request data is invalid")
	}

	TestHandlerWithMock(errorHandler)

	fmt.Println("\nMock implementation works! RouteContext is now mockable.")
}
