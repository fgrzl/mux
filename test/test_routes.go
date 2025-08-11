package test

import (
	"strconv"

	"github.com/fgrzl/mux"
	"github.com/google/uuid"
)

func ConfigureRoutes(r *mux.Router) {
	rg := r.NewRouteGroup("/api/v1")

	rg.GET("/resources/", func(c mux.RouteContext) {
		resources := service.ListResources(0)
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
		_, found = service.GetResource(int32(resourceId))
		if !found {
			c.NotFound()
			return
		}
		c.NoContent()
	}).WithOperationID("checkResourceExists").
		WithParam("resourceId", "path", uuid.Nil, true).
		WithResponse(204, nil).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Resources")

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
		resource, found := service.GetResource(int32(resourceId))
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

	rg.GET("/tenants/", func(c mux.RouteContext) {
		tenants := service.ListTenants()
		if len(tenants) == 0 {
			c.NotFound()
			return
		}
		c.OK(tenants)
	}).WithOperationID("listTenants").
		WithSummary("List all tenants").
		WithResponse(200, []Tenant{}).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Tenants")

	rg.POST("/tenants/", func(c mux.RouteContext) {
		var tenant Tenant
		if err := c.Bind(&tenant); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		if tenant.TenantID == 0 || tenant.Name == "" {
			c.BadRequest("Invalid Input", "tenantID and Name are required.")
			return
		}
		createdTenant := service.PutTenant(&tenant)
		c.Created(createdTenant)
	}).WithOperationID("createTenant").
		WithJsonBody(Tenant{}).
		WithResponse(201, Tenant{}).
		WithResponse(400, mux.ProblemDetails{}).
		WithTags("Tenants")

	rg.PUT("/tenants/{tenantID}", func(c mux.RouteContext) {
		var tenant Tenant
		if err := c.Bind(&tenant); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		result := service.PutTenant(&tenant)
		c.OK(result)
	}).WithOperationID("updateTenant").
		WithParam("tenantID", "path", int(0), true).
		WithJsonBody(Tenant{}).
		WithResponse(200, Tenant{}).
		WithResponse(400, mux.ProblemDetails{}).
		WithTags("Tenants")

	rg.HEAD("/tenants/{tenantID}", func(c mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID")
		if !found {
			c.NotFound()
			return
		}
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.")
			return
		}
		_, found = service.GetTenant(int32(tenantID))
		if !found {
			c.NotFound()
			return
		}
		c.NoContent()
	}).WithOperationID("checkTenantExists").
		WithParam("tenantID", "path", int(0), true).
		WithResponse(204, nil).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Tenants")

	rg.GET("/tenants/{tenantID}", func(c mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID")
		if !found {
			c.NotFound()
			return
		}
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.")
			return
		}
		tenant, found := service.GetTenant(int32(tenantID))
		if !found {
			c.NotFound()
			return
		}
		c.OK(tenant)
	}).WithOperationID("getTenant").
		WithParam("tenantID", "path", int(0), true).
		WithResponse(200, Tenant{}).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Tenants")

	rg.DELETE("/tenants/{tenantID}", func(c mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID")
		if !found {
			c.NotFound()
			return
		}
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.")
			return
		}
		_, found = service.GetTenant(int32(tenantID))
		if !found {
			c.NotFound()
			return
		}
		c.NoContent()
	}).WithOperationID("deleteTenant").
		WithParam("tenantID", "path", int(0), true).
		WithResponse(204, nil).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Tenants")

	rg.GET("/tenants/{tenantID}/resources", func(c mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID")
		if !found {
			c.NotFound()
			return
		}
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.")
			return
		}
		resources := service.ListResources(int32(tenantID))
		if len(resources) == 0 {
			c.NotFound()
			return
		}
		c.OK(resources)
	}).WithOperationID("listTenantResources").
		WithParam("tenantID", "path", int(0), true).
		WithResponse(200, []Resource{}).
		WithResponse(404, mux.ProblemDetails{}).
		WithTags("Tenants", "Resources")

	rg.POST("/tenants/{tenantID}/resources", func(c mux.RouteContext) {
		var resource Resource
		if err := c.Bind(&resource); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}
		if resource.TenantID == 0 || resource.ResourceID == 0 || resource.Name == "" {
			c.BadRequest("Invalid Input", "tenantID, ResourceID and Name are required.")
			return
		}
		createdResource := service.PutResource(&resource)
		c.Created(createdResource)
	}).WithOperationID("createTenantResource").
		WithParam("tenantID", "path", int(0), true).
		WithJsonBody(Resource{}).
		WithResponse(201, Resource{}).
		WithResponse(400, mux.ProblemDetails{}).
		WithTags("Tenants", "Resources")

	rg.StaticFallback("/**", "static", "static/index.html")
}
