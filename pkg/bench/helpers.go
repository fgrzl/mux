package bench

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

// BuildDeepPath returns a path of the form /app/seg/seg/... with n segments.
func BuildDeepPath(n int) string {
	segs := make([]string, n)
	for i := range segs {
		segs[i] = "seg"
	}
	return "/app/" + strings.Join(segs, "/")
}

// NewRecorderRequest returns a recorder and request for the given method and path.
func NewRecorderRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	return rr, req
}

// BuildManyRoutePatterns constructs a slice of route patterns used in some benchmarks.
func BuildManyRoutePatterns(n int) []string {
	patterns := make([]string, 0, n*2+3)
	for i := 0; i < n; i++ {
		patterns = append(patterns, "/static/route/"+strconv.Itoa(i))
		patterns = append(patterns, "/items/{id}/"+strconv.Itoa(i))
	}
	patterns = append(patterns, "/users/{userId}", "/files/*", "/catch/**")
	return patterns
}
