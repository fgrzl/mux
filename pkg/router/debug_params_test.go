package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fgrzl/mux/pkg/routing"
)

func TestDebugParamsNonPooled(t *testing.T) {
	rtr := NewRouter()
	rtr.GET("/users/{id}", func(c routing.RouteContext) {
		t.Logf("params at handler: %#v", c.Params())
		c.OK("ok")
	})
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rr := httptest.NewRecorder()
	rtr.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body:%s", rr.Code, rr.Body.String())
	}
}
