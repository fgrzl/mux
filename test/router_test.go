package test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceRoutes(t *testing.T) {
	ctx, client := StartTestServer(t)
	t.Run("GET /api/v1/resources/", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resources, resp, err := client.ListResources(ctx)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resources, 10)
			// todo : validate etag
		})
	})

	t.Run("HEAD /api/v1/resources/{resourceId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resp, err := client.HeadResource(ctx, 8)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
			// todo : validate etag
		})
		t.Run("Resource Not Found", func(t *testing.T) {
			// Act
			resp, err := client.HeadResource(ctx, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	t.Run("GET /api/v1/resources/{resourceId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resource, resp, err := client.GetResource(ctx, 8)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotNil(t, resource)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			// todo : validate etag
		})
		t.Run("Resource Not Found", func(t *testing.T) {
			// Act
			_, resp, err := client.GetResource(ctx, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}

func TestTenantRoutes(t *testing.T) {
	ctx, client := StartTestServer(t)
	t.Run("GET /api/v1/tenants/", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			tenants, resp, err := client.ListTenants(ctx)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, tenants, 10)
			// todo : validate etag
		})
	})

	t.Run("POST /api/v1/tenants/", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			tenant := &Tenant{TenantId: 1, Name: "Tenant 1"}
			// Act
			resp, err := client.CreateTenant(ctx, tenant)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
		t.Run("Invalid Input", func(t *testing.T) {
			tenant := &Tenant{TenantId: 0, Name: ""}
			// Act
			resp, err := client.CreateTenant(ctx, tenant)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})

	t.Run("HEAD /api/v1/tenants/{tenantId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resp, err := client.HeadTenant(ctx, 8)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})
		t.Run("Tenant Not Found", func(t *testing.T) {
			// Act
			resp, err := client.HeadTenant(ctx, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	t.Run("GET /api/v1/tenants/{tenantId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			tenant, resp, err := client.GetTenant(ctx, 8)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotNil(t, tenant)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("Tenant Not Found", func(t *testing.T) {
			// Act
			_, resp, err := client.GetTenant(ctx, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	t.Run("PUT /api/v1/tenants/{tenantId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			tenant := &Tenant{TenantId: 8, Name: "Updated Tenant"}
			// Act
			resp, err := client.UpdateTenant(ctx, 8, tenant)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("Unauthorized", func(t *testing.T) {
			tenant := &Tenant{TenantId: 8, Name: "Unauthorized Update"}
			// Act
			resp, err := client.UpdateTenant(ctx, 8, tenant)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("DELETE /api/v1/tenants/{tenantId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resp, err := client.DeleteTenant(ctx, 8)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})
		t.Run("Tenant Not Found", func(t *testing.T) {
			// Act
			resp, err := client.DeleteTenant(ctx, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}

func TestTenantResourceRoutes(t *testing.T) {
	ctx, client := StartTestServer(t)
	t.Run("GET /api/v1/tenants/{tenantId}/resources", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resources, resp, err := client.ListTenantResources(ctx, 8)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resources, 10)
		})
	})

	t.Run("POST /api/v1/tenants/{tenantId}/resources", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			resource := &Resource{ResourceId: 1, Name: "Resource 1"}
			// Act
			resp, err := client.CreateTenantResource(ctx, 8, resource)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		})
		t.Run("Invalid Input", func(t *testing.T) {
			resource := &Resource{TenantId: 0, Name: ""}
			// Act
			resp, err := client.CreateTenantResource(ctx, 8, resource)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})

	t.Run("HEAD /api/v1/tenants/{tenantId}/resources/{resourceId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resp, err := client.HeadTenantResource(ctx, 8, 5)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		})
		t.Run("Resource Not Found", func(t *testing.T) {
			// Act
			resp, err := client.HeadTenantResource(ctx, 8, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	t.Run("GET /api/v1/tenants/{tenantId}/resources/{resourceId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resource, resp, err := client.GetTenantResource(ctx, 8, 5)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.NotNil(t, resource)
		})
		t.Run("Resource Not Found", func(t *testing.T) {
			// Act
			resource, resp, err := client.GetTenantResource(ctx, 8, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Nil(t, resource)
		})
	})

	t.Run("PUT /api/v1/tenants/{tenantId}/resources/{resourceId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			resource := &Resource{ResourceId: 5, Name: "Updated Resource"}
			// Act
			resp, err := client.UpdateTenantResource(ctx, 8, 5, resource)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("Unauthorized", func(t *testing.T) {
			resource := &Resource{ResourceId: 5, Name: "Unauthorized Update"}
			// Act
			resp, err := client.UpdateTenantResource(ctx, 8, 5, resource)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("DELETE /api/v1/tenants/{tenantId}/resources/{resourceId}", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// Act
			resp, err := client.DeleteTenantResource(ctx, 8, 5)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("Resource Not Found", func(t *testing.T) {
			// Act
			resp, err := client.DeleteTenantResource(ctx, 8, 99999)

			// Assert
			assert.Error(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}
