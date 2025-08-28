package test

// import (
// 	"testing"

// 	"github.com/fgrzl/mux/pkg/openapi"
// 	routing "github.com/fgrzl/mux/internal/routing"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestWithPathPrefixFiltersRoutesWithLeadingSlash(t *testing.T) {
// 	router := NewRouter(
// 		WithTitle("Test API"),
// 		WithVersion("1.0.0"),
// 	)

// 	rg := router.NewRouteGroup("/api/v1")
// 	rg.GET("/hello", func(c routing.RouteContext) { c.OK("ok") }).WithOperationID("helloOp")

// 	// Another route outside the prefix
// 	router.GET("/other/ping", func(c routing.RouteContext) { c.OK("pong") }).WithOperationID("pingOp")

// 	gen := openapi.NewGenerator(openapi.WithPathPrefix("/api/v1"))
// 	spec, err := openapi.GenerateSpecWithGenerator(gen, router)
// 	require.NoError(t, err)
// 	require.NotNil(t, spec)

// 	// Only /api/v1/hello should be present
// 	assert.Contains(t, spec.Paths, "/api/v1/hello")
// 	assert.NotContains(t, spec.Paths, "/other/ping")
// }

// func TestWithPathPrefixNormalizesPrefixWithoutLeadingSlash(t *testing.T) {
// 	router := NewRouter(
// 		WithTitle("Test API"),
// 		WithVersion("1.0.0"),
// 	)

// 	rg := router.NewRouteGroup("/api/v2")
// 	rg.POST("/submit", func(c routing.RouteContext) { c.OK("ok") }).WithOperationID("submitOp")

// 	gen := openapi.NewGenerator(openapi.WithPathPrefix("api/v2")) // no leading slash
// 	spec, err := openapi.GenerateSpecWithGenerator(gen, router)
// 	require.NoError(t, err)
// 	require.NotNil(t, spec)

// 	assert.Contains(t, spec.Paths, "/api/v2/submit")
// }

// func TestWithPathPrefixAccumulatesMultiplePrefixes(t *testing.T) {
// 	router := NewRouter(
// 		WithTitle("Test API"),
// 		WithVersion("1.0.0"),
// 	)

// 	rg1 := router.NewRouteGroup("/svc1")
// 	rg1.GET("/a", func(c routing.RouteContext) { c.OK("a") }).WithOperationID("aOp")

// 	rg2 := router.NewRouteGroup("/svc2")
// 	rg2.GET("/b", func(c routing.RouteContext) { c.OK("b") }).WithOperationID("bOp")

// 	// Accumulate prefixes via multiple options
// 	gen := openapi.NewGenerator(openapi.WithPathPrefix("svc1"), openapi.WithPathPrefix("/svc2"))
// 	spec, err := openapi.GenerateSpecWithGenerator(gen, router)
// 	require.NoError(t, err)
// 	require.NotNil(t, spec)

// 	assert.Contains(t, spec.Paths, "/svc1/a")
// 	assert.Contains(t, spec.Paths, "/svc2/b")
// }
