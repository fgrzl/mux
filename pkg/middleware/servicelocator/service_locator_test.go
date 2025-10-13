package servicelocator

import (
	"net/http"
	"testing"

	"github.com/fgrzl/mux/pkg/router"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/fgrzl/mux/test/testhelpers"
	"github.com/stretchr/testify/assert"
)

type MockService struct {
	Name string
}

func TestShouldSetSingleServiceOnRouteContext(t *testing.T) {
	// Arrange
	service := &MockService{Name: "test-service"}
	serviceKey := routing.ServiceKey("testService")

	middleware := &serviceSetterMiddleware{
		options: &ServiceSetterOptions{
			Services: map[routing.ServiceKey]any{
				serviceKey: service,
			},
		},
	}

	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/test", nil)

	nextCalled := false
	next := func(c routing.RouteContext) {
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
	serviceKey1 := routing.ServiceKey("service1")
	serviceKey2 := routing.ServiceKey("service2")

	middleware := &serviceSetterMiddleware{
		options: &ServiceSetterOptions{
			Services: map[routing.ServiceKey]any{
				serviceKey1: service1,
				serviceKey2: service2,
			},
		},
	}

	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/test", nil)

	nextCalled := false
	next := func(c routing.RouteContext) {
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

	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/test", nil)

	nextCalled := false
	next := func(c routing.RouteContext) {
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

	ctx, _ := testhelpers.NewRouteContext(http.MethodGet, "/test", nil)

	nextCalled := false
	next := func(c routing.RouteContext) {
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
	serviceKey := routing.ServiceKey("test")
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
	serviceKey := routing.ServiceKey("test")
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
	rtr := router.NewRouter()
	service := &MockService{Name: "test"}

	// Act - register middleware
	UseServices(rtr, WithService("test", service))

	// Register a route and ensure the middleware injects the service
	rtr.GET("/test", func(c routing.RouteContext) {
		s, ok := c.GetService(routing.ServiceKey("test"))
		if ok {
			if ms, ok := s.(*MockService); ok {
				c.Response().Write([]byte(ms.Name))
				return
			}
		}
		c.Response().Write([]byte("no-service"))
	})

	req, rec := testhelpers.NewRequestRecorder(http.MethodGet, "/test", nil)
	rtr.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())
}
