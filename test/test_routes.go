package test

import (
	"strconv"

	"github.com/fgrzl/mux"
)

func ConfigureRoutes(r *mux.Router) {

	// route groups help to organize your routes
	rg := r.NewRouteGroup("/api/v1")

	// Resource routes
	rg.GET("/resources/", func(c *mux.RouteContext) {
		// List all resources
		resources := service.ListResources(0)
		if len(resources) == 0 {
			c.NotFound() // return 404 if no resources are found
			return
		}
		c.OK(resources) // return 200 with the list of resources
	}).AllowAnonymous()

	rg.HEAD("/resources/{resourceId}", func(c *mux.RouteContext) {
		resourceIdStr, found := c.Param("resourceId") // Get resourceId as string
		if !found {
			c.NotFound() // return 404 if resourceId is not found
			return
		}
		// Convert to int32
		resourceId, err := strconv.ParseInt(resourceIdStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid ResourceID", "Failed to parse resource id.") // return 400 if conversion fails
			return
		}
		// Check if resource exists by ID
		_, found = service.GetResource(int32(resourceId))
		if !found {
			c.NotFound() // return 404 if resource is not found
			return
		}
		c.NoContent() // return 204 if resource exists
	})

	rg.GET("/resources/{resourceId}", func(c *mux.RouteContext) {
		resourceIdStr, found := c.Param("resourceId") // Get resourceId as string
		if !found {
			c.NotFound() // return 404 if resourceId is not found
			return
		}
		// Convert to int32
		resourceId, err := strconv.ParseInt(resourceIdStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid ResourceID", "Failed to parse resource id.") // return 400 if conversion fails
			return
		}
		// Get resource by ID
		resource, found := service.GetResource(int32(resourceId))
		if !found {
			c.NotFound() // return 404 if resource not found
			return
		}
		c.OK(resource) // return 200 with the resource data
	})

	// Tenant routes
	rg.GET("/tenants/", func(c *mux.RouteContext) {
		// List all tenants
		tenants := service.ListTenants()
		if len(tenants) == 0 {
			c.NotFound() // return 404 if no tenants are found
			return
		}
		c.OK(tenants) // return 200 with the list of tenants
	})

	rg.POST("/tenants/", func(c *mux.RouteContext) {
		var tenant Tenant
		if err := c.Bind(&tenant); err != nil {
			c.BadRequest("Bad Request", err.Error()) // return 400 if input is invalid
			return
		}
		if tenant.TenantID == 0 || tenant.Name == "" {
			c.BadRequest("Invalid Input", "tenantID and Name are required.") // return 400 if input is invalid
			return
		}
		// Create the tenant
		createdTenant := service.PutTenant(&tenant)
		c.Created(createdTenant) // return 201 with the created tenant
	})

	rg.PUT("/tenants/{tenantID}", func(c *mux.RouteContext) {
		var tenant Tenant
		if err := c.Bind(&tenant); err != nil {
			c.BadRequest("Bad Request", err.Error())
			return
		}

		result := service.PutTenant(&tenant)
		c.OK(result)
	})

	rg.HEAD("/tenants/{tenantID}", func(c *mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID") // Get tenantID as string
		if !found {
			c.NotFound() // return 404 if tenantID is not found
			return
		}
		// Convert to int32
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.") // return 400 if conversion fails
			return
		}
		// Check if tenant exists by ID
		_, found = service.GetTenant(int32(tenantID))
		if !found {
			c.NotFound() // return 404 if tenant not found
			return
		}
		c.NoContent() // return 204 if tenant exists
	})

	rg.GET("/tenants/{tenantID}", func(c *mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID") // Get tenantID as string
		if !found {
			c.NotFound() // return 404 if tenantID is not found
			return
		}
		// Convert to int32
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.") // return 400 if conversion fails
			return
		}
		// Get tenant by ID
		tenant, found := service.GetTenant(int32(tenantID))
		if !found {
			c.NotFound() // return 404 if tenant not found
			return
		}
		c.OK(tenant) // return 200 with the tenant data
	})

	rg.DELETE("/tenants/{tenantID}", func(c *mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID") // Get tenantID as string
		if !found {
			c.NotFound() // return 404 if tenantID is not found
			return
		}
		// Convert to int32
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.") // return 400 if conversion fails
			return
		}
		// Get tenant by ID
		_, found = service.GetTenant(int32(tenantID))
		if !found {
			c.NotFound() // return 404 if tenant not found
			return
		}
		c.NoContent()
	})

	// Tenant-Resource routes
	rg.GET("/tenants/{tenantID}/resources", func(c *mux.RouteContext) {
		tenantIDStr, found := c.Param("tenantID") // Get tenantID as string
		if !found {
			c.NotFound() // return 404 if tenantID is not found
			return
		}
		// Convert to int32
		tenantID, err := strconv.ParseInt(tenantIDStr, 10, 32)
		if err != nil {
			c.BadRequest("Invalid tenantID", "Failed to parse tenant id.") // return 400 if conversion fails
			return
		}
		// List resources for a tenant
		resources := service.ListResources(int32(tenantID))
		if len(resources) == 0 {
			c.NotFound() // return 404 if no resources are found for the tenant
			return
		}
		c.OK(resources) // return 200 with the list of resources
	})

	rg.POST("/tenants/{tenantID}/resources", func(c *mux.RouteContext) {
		var resource Resource
		if err := c.Bind(&resource); err != nil {
			c.BadRequest("Bad Request", err.Error()) // return 400 if input is invalid
			return
		}
		if resource.TenantID == 0 || resource.ResourceID == 0 || resource.Name == "" {
			c.BadRequest("Invalid Input", "tenantID, ResourceID and Name are required.") // return 400 if input is invalid
			return
		}
		// Create resource for tenant
		createdResource := service.PutResource(&resource)
		c.Created(createdResource) // return 201 with the created resource
	})

	rg.StaticFallback("/**", "static", "static/index.html") // Serve static files from the "static" directory
}
