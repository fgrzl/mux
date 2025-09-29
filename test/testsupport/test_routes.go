package testsupport

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/fgrzl/mux"
	"github.com/google/uuid"
)

func computeAccept(key string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + magic))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ConfigureRoutes registers a broad set of routes used by unit and integration
// tests. These routes exercise many RouteContext features and edge-cases so
// the test-suite can validate behavior and OpenAPI generation.
func ConfigureRoutes(r *mux.Router) {

	// Minimal websocket upgrade route at top-level so tests can exercise
	// Upgrade/Hijack behavior. This route lives at `/ws` (root) instead of
	// under /api/v1 so test clients can connect directly to /ws.
	r.GET("/ws", func(c mux.RouteContext) {
		if strings.ToLower(c.Request().Header.Get("Upgrade")) != "websocket" {
			c.NotFound()
			return
		}
		key := c.Request().Header.Get("Sec-WebSocket-Key")
		if key == "" {
			c.BadRequest("Missing Sec-WebSocket-Key", "Sec-WebSocket-Key header is required")
			return
		}
		accept := computeAccept(key)
		rw := c.Response()
		hj, ok := rw.(http.Hijacker)
		if !ok {
			c.ServerError("Hijack unsupported", "ResponseWriter does not support hijack")
			return
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			c.ServerError("Hijack failed", err.Error())
			return
		}
		// Write raw upgrade response and close the connection.
		fmt.Fprintf(conn, "HTTP/1.1 101 Switching Protocols\r\n")
		fmt.Fprintf(conn, "Upgrade: websocket\r\n")
		fmt.Fprintf(conn, "Connection: Upgrade\r\n")
		fmt.Fprintf(conn, "Sec-WebSocket-Accept: %s\r\n", accept)
		fmt.Fprintf(conn, "\r\n")
		_ = conn.Close()
	})

	rg := r.NewRouteGroup("/api/v1")

	// Basic resources list
	rg.GET("/resources/", func(c mux.RouteContext) {
		resources := Service.ListResources(0)
		if len(resources) == 0 {
			c.NotFound()
			return
		}
		c.OK(resources)
	}).AllowAnonymous().
		WithOperationID("listResources").
		WithSummary("List all resources").
		WithResponse(200, ResourcePage{}).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Resources")

	// HEAD to check existence
	rg.HEAD("/resources/{resourceId}", func(c mux.RouteContext) {
		resourceIdStr, found := c.Param("resourceId")
		if !found {
			c.NotFound()
			return
		}
		resourceId, err := strconv.ParseInt(resourceIdStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid ResourceID", "Failed to parse resource id.")
			return
		}
		_, found = Service.GetResource(int32(resourceId))
		if !found {
			c.NotFound()
			return
		}
		c.NoContent()
	}).WithOperationID("checkResourceExists").
		WithParam("resourceId", "path", int(0), true).
		WithResponse(204, nil).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Resources")

	// GET a single resource
	rg.GET("/resources/{resourceId}", func(c mux.RouteContext) {
		resourceIdStr, found := c.Param("resourceId")
		if !found {
			c.NotFound()
			return
		}
		resourceId, err := strconv.ParseInt(resourceIdStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid ResourceID", "Failed to parse resource id.")
			return
		}
		resource, found := Service.GetResource(int32(resourceId))
		if !found {
			c.NotFound()
			return
		}
		c.OK(resource)
	}).WithOperationID("getResource").
		WithParam("resourceId", "path", int(0), true).
		WithResponse(200, Resource{}).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Resources")

	// Search by query params (single and repeated)
	rg.GET("/resources/search", func(c mux.RouteContext) {
		name, _ := c.QueryValue("name")
		types, _ := c.QueryValues("type")
		var out []Resource
		for _, rsrc := range Service.ListResources(0) {
			if name != "" && rsrc.Name != name {
				continue
			}
			if len(types) > 0 {
				matched := false
				for _, t := range types {
					if rsrc.Type == t {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
			out = append(out, *rsrc)
		}
		if len(out) == 0 {
			c.NotFound()
			return
		}
		c.OK(out)
	}).WithOperationID("searchResources").
		WithParam("name", "query", "", false).
		WithParam("type", "query", "", false).
		WithResponse(200, []Resource{}).
		WithTags("Resources")

	// Bulk create resources via JSON array
	rg.POST("/resources/bulk", func(c mux.RouteContext) {
		var resources []Resource
		if err := c.Bind(&resources); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		// simple validation
		if len(resources) == 0 {
			c.BadRequest("Invalid Input", "no resources provided")
			return
		}
		// pretend to store them
		var created []Resource
		for i := range resources {
			r := Service.PutResource(&resources[i])
			created = append(created, *r)
		}
		c.Created(created)
	}).WithOperationID("createResourcesBulk").
		WithJsonBody([]Resource{}).
		WithResponse(201, []Resource{}).
		WithResponse(400, mux.ProblemDetails{}).
		WithTags("Resources")

	// Update resource metadata — exercise map/object JSON bodies
	rg.PUT("/resources/{resourceId}/metadata", func(c mux.RouteContext) {
		var body struct {
			Metadata map[string]string `json:"metadata"`
		}
		if err := c.Bind(&body); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		resourceId, ok := c.ParamInt32("resourceId")
		if !ok {
			c.BadRequest("Invalid ResourceID", "resourceId missing or invalid")
			return
		}
		rsrc, found := Service.GetResource(resourceId)
		if !found {
			c.NotFound()
			return
		}
		// For tests we simply echo back the received metadata along with the id
		c.OK(map[string]any{"resource": rsrc, "metadata": body.Metadata})
	}).WithOperationID("updateResourceMetadata").
		WithParam("resourceId", "path", int(0), true).
		WithJsonBody(&struct {
			Metadata map[string]string `json:"metadata"`
		}{}).
		WithResponse(200, struct {
			Resource Resource          `json:"resource"`
			Metadata map[string]string `json:"metadata"`
		}{}).
		WithTags("Resources")

	// UUID path parameter example
	rg.GET("/items/{itemId}/uuid", func(c mux.RouteContext) {
		id, ok := c.ParamUUID("itemId")
		if !ok {
			c.BadRequest("Invalid UUID", "itemId is required and must be a UUID")
			return
		}
		c.OK(map[string]uuid.UUID{"id": id})
	}).WithOperationID("getItemByUUID").
		WithParam("itemId", "path", uuid.Nil, true).
		WithResponse(200, map[string]uuid.UUID{}).
		WithTags("Items")

	// Simple header echo to exercise Header lookup
	rg.GET("/headers/echo", func(c mux.RouteContext) {
		if val, ok := c.Header("X-Echo"); ok {
			c.Plain(200, []byte(val))
			return
		}
		c.NotFound()
	}).WithOperationID("echoHeader").
		WithResponse(200, "").
		WithTags("Misc")

	// Query multiple ints and UUIDs to exercise QueryInts/QueryUUIDs
	rg.GET("/filter", func(c mux.RouteContext) {
		ints, _ := c.QueryInts("ids")
		uuids, _ := c.QueryUUIDs("uuids")
		c.OK(map[string]any{"ids": ints, "uuids": uuids})
	}).WithOperationID("filter").
		WithParam("ids", "query", "", false).
		WithParam("uuids", "query", uuid.Nil, false).
		WithResponse(200, struct {
			IDs   []int       `json:"ids"`
			UUIDs []uuid.UUID `json:"uuids"`
		}{}).
		WithTags("Misc")

	// Form submission example (Bind should support form-encoded bodies)
	rg.POST("/form/submit", func(c mux.RouteContext) {
		var body struct {
			Field string `json:"field"`
		}
		if err := c.Bind(&body); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		c.OK(body)
	}).WithOperationID("submitForm").
		WithResponse(200, map[string]string{}).
		WithTags("Misc")

	rg.StaticFallback("/**", "static", "static/index.html")

	// Tenant management routes (list, get, create, update, delete)
	rg.GET("/tenants/", func(c mux.RouteContext) {
		tenants := Service.ListTenants()
		if len(tenants) == 0 {
			c.NotFound()
			return
		}
		c.OK(tenants)
	}).WithOperationID("listTenants").WithSummary("List all tenants").WithResponse(200, []Tenant{}).WithTags("Tenants")

	rg.POST("/tenants/", func(c mux.RouteContext) {
		var t Tenant
		if err := c.Bind(&t); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		created := Service.PutTenant(&t)
		c.Created(created)
	}).WithOperationID("createTenant").WithJsonBody(Tenant{}).WithResponse(201, Tenant{}).WithTags("Tenants")

	rg.GET("/tenants/{tenantID}", func(c mux.RouteContext) {
		id, ok := c.ParamInt("tenantID")
		if !ok {
			c.NotFound()
			return
		}
		tenant, found := Service.GetTenant(int32(id))
		if !found {
			c.NotFound()
			return
		}
		c.OK(tenant)
	}).WithOperationID("getTenant").WithParam("tenantID", "path", int(0), true).WithResponse(200, Tenant{}).WithResponse(404, mux.ProblemDetails{}).WithTags("Tenants")

	rg.PUT("/tenants/{tenantID}", func(c mux.RouteContext) {
		id, ok := c.ParamInt("tenantID")
		if !ok {
			c.BadRequest("Invalid TenantID", "tenantID missing or invalid")
			return
		}
		var t Tenant
		if err := c.Bind(&t); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		t.TenantID = int32(id)
		updated := Service.PutTenant(&t)
		c.OK(updated)
	}).WithOperationID("updateTenant").WithParam("tenantID", "path", int(0), true).WithJsonBody(Tenant{}).WithResponse(200, Tenant{}).WithTags("Tenants")

	rg.DELETE("/tenants/{tenantID}", func(c mux.RouteContext) {
		id, ok := c.ParamInt("tenantID")
		if !ok {
			c.BadRequest("Invalid TenantID", "tenantID missing or invalid")
			return
		}
		deleted := Service.DeleteTenant(int32(id))
		if !deleted {
			c.NotFound()
			return
		}
		c.NoContent()
	}).WithOperationID("deleteTenant").WithParam("tenantID", "path", int(0), true).WithResponse(204, nil).WithResponse(404, mux.ProblemDetails{}).WithTags("Tenants")

	// Tenant resources
	rg.GET("/tenants/{tenantID}/resources", func(c mux.RouteContext) {
		id, ok := c.ParamInt("tenantID")
		if !ok {
			c.BadRequest("Invalid TenantID", "tenantID missing or invalid")
			return
		}
		resources := Service.ListResources(int32(id))
		if len(resources) == 0 {
			c.NotFound()
			return
		}
		c.OK(resources)
	}).WithOperationID("listTenantResources").WithParam("tenantID", "path", int(0), true).WithResponse(200, []Resource{}).WithResponse(404, mux.ProblemDetails{}).WithTags("Tenants", "Resources")

	rg.POST("/tenants/{tenantID}/resources", func(c mux.RouteContext) {
		id, ok := c.ParamInt("tenantID")
		if !ok {
			c.BadRequest("Invalid TenantID", "tenantID missing or invalid")
			return
		}
		var res Resource
		if err := c.Bind(&res); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		res.TenantID = int32(id)
		created := Service.PutResource(&res)
		c.Created(created)
	}).WithOperationID("createTenantResource").WithParam("tenantID", "path", int(0), true).WithJsonBody(Resource{}).WithResponse(201, Resource{}).WithTags("Tenants", "Resources")
}
