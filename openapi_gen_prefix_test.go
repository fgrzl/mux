package mux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithPathPrefixFiltersRoutesWithLeadingSlash(t *testing.T) {
	router := NewRouter(
		WithTitle("Test API"),
		WithVersion("1.0.0"),
	)

	rg := router.NewRouteGroup("/api/v1")
	rg.GET("/hello", func(c RouteContext) { c.OK("ok") }).WithOperationID("helloOp")

	// Another route outside the prefix
	router.GET("/other/ping", func(c RouteContext) { c.OK("pong") }).WithOperationID("pingOp")

	gen := NewGenerator(WithPathPrefix("/api/v1"))
	spec, err := gen.GenerateSpec(router)
	require.NoError(t, err)
	require.NotNil(t, spec)

	// Only /api/v1/hello should be present
	assert.Contains(t, spec.Paths, "/api/v1/hello")
	assert.NotContains(t, spec.Paths, "/other/ping")
}

func TestWithPathPrefixNormalizesPrefixWithoutLeadingSlash(t *testing.T) {
	router := NewRouter(
		WithTitle("Test API"),
		WithVersion("1.0.0"),
	)

	rg := router.NewRouteGroup("/api/v2")
	rg.POST("/submit", func(c RouteContext) { c.OK("ok") }).WithOperationID("submitOp")

	gen := NewGenerator(WithPathPrefix("api/v2")) // no leading slash
	spec, err := gen.GenerateSpec(router)
	require.NoError(t, err)
	require.NotNil(t, spec)

	assert.Contains(t, spec.Paths, "/api/v2/submit")
}

func TestWithPathPrefixAccumulatesMultiplePrefixes(t *testing.T) {
	router := NewRouter(
		WithTitle("Test API"),
		WithVersion("1.0.0"),
	)

	rg1 := router.NewRouteGroup("/svc1")
	rg1.GET("/a", func(c RouteContext) { c.OK("a") }).WithOperationID("aOp")

	rg2 := router.NewRouteGroup("/svc2")
	rg2.GET("/b", func(c RouteContext) { c.OK("b") }).WithOperationID("bOp")

	// Accumulate prefixes via multiple options
	gen := NewGenerator(WithPathPrefix("svc1"), WithPathPrefix("/svc2"))
	spec, err := gen.GenerateSpec(router)
	require.NoError(t, err)
	require.NotNil(t, spec)

	assert.Contains(t, spec.Paths, "/svc1/a")
	assert.Contains(t, spec.Paths, "/svc2/b")
}
