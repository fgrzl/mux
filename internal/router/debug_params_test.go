package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/internal/routing"
)

func TestDebugParamsNonPooled(t *testing.T) {
	// Arrange
	rtr := NewRouter()
	rtr.GET("/users/{id}", func(c routing.RouteContext) {
		id, _ := c.Param("id")
		t.Logf("params at handler: id=%s", id)
		c.OK("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rr := httptest.NewRecorder()

	// Act
	rtr.ServeHTTP(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body:%s", rr.Code, rr.Body.String())
	}
}
