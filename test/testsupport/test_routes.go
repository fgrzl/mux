package testsupport

import (
	"strconv"

	"github.com/fgrzl/mux"
	"github.com/google/uuid"
)

// ConfigureRoutes registers a broad set of routes used by unit and integration
// tests. These routes exercise many RouteContext features and edge-cases so
// the test-suite can validate behavior and OpenAPI generation.
func ConfigureRoutes(r *mux.Router) {
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
		resourceId, ok := c.ParamInt("resourceId")
		if !ok {
			c.BadRequest("Invalid ResourceID", "resourceId missing or invalid")
			return
		}
		rsrc, found := Service.GetResource(int32(resourceId))
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
}
