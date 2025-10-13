package testsupport

// Common API paths used by integration tests. Centralizing them here reduces
// duplication and makes it easier to update paths in one place.
const (
	APIBase             = "/api/v1"
	APIResources        = APIBase + "/resources/"
	APIResourceByID     = APIBase + "/resources/%d"
	APIResourceMetadata = APIBase + "/resources/%d/metadata"
	APITenants          = APIBase + "/tenants/"
	APITenantByID       = APIBase + "/tenants/%d"
	APITenantResources  = APIBase + "/tenants/%d/resources"
	APIHeadersEcho      = APIBase + "/headers/echo"
	APIFormSubmit       = APIBase + "/form/submit"
	APIItemsUUID        = APIBase + "/items/%s/uuid"
)
