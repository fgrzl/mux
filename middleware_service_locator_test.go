package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockService struct {
	Name string
}

func TestShouldSetSingleServiceOnRouteContext(t *testing.T) {
	// Arrange
	service := &MockService{Name: "test-service"}
	serviceKey := ServiceKey("testService")

	middleware := &serviceSetterMiddleware{
		options: &ServiceSetterOptions{
			Services: map[ServiceKey]any{
				serviceKey: service,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
		// Verify service was set
		retrievedService, ok := c.GetService(serviceKey)
		assert.True(t, ok)
		assert.Equal(t, service, retrievedService)
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")
}

func TestShouldSetMultipleServicesOnRouteContext(t *testing.T) {
	// Arrange
	service1 := &MockService{Name: "service1"}
	service2 := &MockService{Name: "service2"}
	serviceKey1 := ServiceKey("service1")
	serviceKey2 := ServiceKey("service2")

	middleware := &serviceSetterMiddleware{
		options: &ServiceSetterOptions{
			Services: map[ServiceKey]any{
				serviceKey1: service1,
				serviceKey2: service2,
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
		// Verify both services were set
		retrievedService1, ok1 := c.GetService(serviceKey1)
		assert.True(t, ok1)
		assert.Equal(t, service1, retrievedService1)

		retrievedService2, ok2 := c.GetService(serviceKey2)
		assert.True(t, ok2)
		assert.Equal(t, service2, retrievedService2)
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called")
}

func TestShouldHandleNilOptionsGracefully(t *testing.T) {
	// Arrange
	middleware := &serviceSetterMiddleware{options: nil}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called even with nil options")
}

func TestShouldHandleEmptyServicesMapGracefully(t *testing.T) {
	// Arrange
	middleware := &serviceSetterMiddleware{
		options: &ServiceSetterOptions{Services: nil},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()
	ctx := NewRouteContext(recorder, req)

	nextCalled := false
	next := func(c RouteContext) {
		nextCalled = true
	}

	// Act
	middleware.Invoke(ctx, next)

	// Assert
	assert.True(t, nextCalled, "next handler should be called even with nil services map")
}

func TestWithServiceShouldCreateServiceOption(t *testing.T) {
	// Arrange
	service := &MockService{Name: "test"}
	serviceKey := ServiceKey("test")
	options := &ServiceSetterOptions{}

	// Act
	option := WithService(serviceKey, service)
	option(options)

	// Assert
	assert.NotNil(t, options.Services)
	assert.Equal(t, service, options.Services[serviceKey])
}

func TestWithServiceShouldInitializeServicesMap(t *testing.T) {
	// Arrange
	service := &MockService{Name: "test"}
	serviceKey := ServiceKey("test")
	options := &ServiceSetterOptions{} // Services map is nil

	// Act
	option := WithService(serviceKey, service)
	option(options)

	// Assert
	assert.NotNil(t, options.Services)
	assert.Equal(t, service, options.Services[serviceKey])
}

func TestShouldAddServiceMiddlewareToRouter(t *testing.T) {
	// Arrange
	router := NewRouter()
	initialMiddlewareCount := len(router.middleware)
	service := &MockService{Name: "test"}

	// Act
	UseServices(router, WithService("test", service))

	// Assert
	assert.Equal(t, initialMiddlewareCount+1, len(router.middleware))
	assert.IsType(t, &serviceSetterMiddleware{}, router.middleware[len(router.middleware)-1])
}
