package testsupport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TestClient encapsulates the client and base URL
type TestClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewTestClient creates a new HTTP client instance
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Helper methods
func (c *TestClient) doRequest(req *http.Request, result interface{}) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return resp, nil
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// Resource Routes
func (c *TestClient) ListResources(ctx context.Context) ([]Resource, *http.Response, error) {
	url := c.BaseURL + "/api/v1/resources/"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	var resources []Resource
	resp, err := c.doRequest(req, &resources)
	return resources, resp, err
}

func (c *TestClient) GetResource(ctx context.Context, resourceId int32) (*Resource, *http.Response, error) {
	url := c.BaseURL + "/api/v1/resources/" + fmt.Sprintf("%d", resourceId)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	var resource Resource
	resp, err := c.doRequest(req, &resource)
	return &resource, resp, err
}

func (c *TestClient) HeadResource(ctx context.Context, resourceId int32) (*http.Response, error) {
	url := c.BaseURL + "/api/v1/resources/" + fmt.Sprintf("%d", resourceId)
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)

	resp, err := c.doRequest(req, nil)
	return resp, err
}

// Tenant Routes
func (c *TestClient) ListTenants(ctx context.Context) ([]Tenant, *http.Response, error) {
	url := c.BaseURL + "/api/v1/tenants/"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	var tenants []Tenant
	resp, err := c.doRequest(req, &tenants)
	return tenants, resp, err
}

func (c *TestClient) CreateTenant(ctx context.Context, tenant *Tenant) (*http.Response, error) {
	url := c.BaseURL + "/api/v1/tenants/"

	data, _ := json.Marshal(tenant)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req, nil)
	return resp, err
}

func (c *TestClient) GetTenant(ctx context.Context, tenantID int32) (*Tenant, *http.Response, error) {
	url := c.BaseURL + "/api/v1/tenants/" + fmt.Sprintf("%d", tenantID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	var tenant Tenant
	resp, err := c.doRequest(req, &tenant)
	return &tenant, resp, err
}

func (c *TestClient) HeadTenant(ctx context.Context, tenantID int32) (*http.Response, error) {
	url := c.BaseURL + "/api/v1/tenants/" + fmt.Sprintf("%d", tenantID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)

	resp, err := c.doRequest(req, nil)
	return resp, err
}

func (c *TestClient) UpdateTenant(ctx context.Context, tenantID int32, tenant *Tenant) (*http.Response, error) {
	url := c.BaseURL + "/api/v1/tenants/" + fmt.Sprintf("%d", tenantID)

	data, _ := json.Marshal(tenant)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req, nil)
	return resp, err
}

func (c *TestClient) DeleteTenant(ctx context.Context, tenantID int32) (*http.Response, error) {
	url := c.BaseURL + "/api/v1/tenants/" + fmt.Sprintf("%d", tenantID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)

	resp, err := c.doRequest(req, nil)
	return resp, err
}

// Tenant Resource Routes
func (c *TestClient) ListTenantResources(ctx context.Context, tenantID int32) ([]Resource, *http.Response, error) {
	url := c.BaseURL + fmt.Sprintf("/api/v1/tenants/%d/resources", tenantID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	var resources []Resource
	resp, err := c.doRequest(req, &resources)
	return resources, resp, err
}

func (c *TestClient) CreateTenantResource(ctx context.Context, tenantID int32, resource *Resource) (*http.Response, error) {
	url := c.BaseURL + fmt.Sprintf("/api/v1/tenants/%d/resources", tenantID)

	data, _ := json.Marshal(resource)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req, nil)
	return resp, err
}

func (c *TestClient) GetTenantResource(ctx context.Context, tenantID, resourceId int32) (*Resource, *http.Response, error) {
	url := c.BaseURL + fmt.Sprintf("/api/v1/tenants/%d/resources/%d", tenantID, resourceId)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	var resource Resource
	resp, err := c.doRequest(req, &resource)
	return &resource, resp, err
}

func (c *TestClient) HeadTenantResource(ctx context.Context, tenantID, resourceId int32) (*http.Response, error) {
	url := c.BaseURL + fmt.Sprintf("/api/v1/tenants/%d/resources/%d", tenantID, resourceId)
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)

	resp, err := c.doRequest(req, nil)
	return resp, err
}

func (c *TestClient) UpdateTenantResource(ctx context.Context, tenantID, resourceId int32, resource *Resource) (*http.Response, error) {
	url := c.BaseURL + fmt.Sprintf("/api/v1/tenants/%d/resources/%d", tenantID, resourceId)

	data, _ := json.Marshal(resource)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req, nil)
	return resp, err
}

func (c *TestClient) DeleteTenantResource(ctx context.Context, tenantID, resourceId int32) (*http.Response, error) {
	url := c.BaseURL + fmt.Sprintf("/api/v1/tenants/%d/resources/%d", tenantID, resourceId)
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)

	resp, err := c.doRequest(req, nil)
	return resp, err
}
