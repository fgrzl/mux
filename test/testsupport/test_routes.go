package testsupport

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/fgrzl/mux"
	"github.com/fgrzl/mux/internal/common"
	"github.com/google/uuid"
)

// Constants to avoid repeating string literals throughout the test routes.
const (
	APIPrefix       = "/api/v1"
	WSPath          = "/ws"
	ResourcesPath   = "/resources"
	ResourcesBulk   = "/resources/bulk"
	ResourcesSearch = "/resources/search"
	ResourceMeta    = "/resources/{resourceId}/metadata"
	ResourceIDPath  = "/resources/{resourceId}"
	ItemsUUIDPath   = "/items/{itemId}/uuid"
	HeadersEchoPath = "/headers/echo"
	FilterPath      = "/filter"
	FormSubmitPath  = "/form/submit"
	TenantsPath     = "/tenants"
	TenantIDPath    = "/tenants/{tenantID}"
	TenantResources = "/tenants/{tenantID}/resources"

	ParamResourceID = "resourceId"
	ParamTenantID   = "tenantID"
	ParamItemID     = "itemId"

	TagResources = "Resources"
	TagTenants   = "Tenants"
	TagMisc      = "Misc"
	TagItems     = "Items"
)

const (
	ErrInvalidResourceID = "Invalid ResourceID"
	ErrParseResourceID   = "Failed to parse resource id."
	ErrBadRequest        = "Bad Request"
	ErrInvalidTenantID   = "Invalid TenantID"
	ErrTenantMissing     = "tenantID missing or invalid"
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
	r.GET(WSPath, wsHandler)

	rg := r.Group(APIPrefix)

	// Basic resources list
	rg.GET(ResourcesPath+"/", listResourcesHandler).AllowAnonymous().
		WithOperationID("listResources").
		WithSummary("List all resources").
		WithOKResponse(ResourcePage{}).
		WithResponse(http.StatusNotFound, mux.ProblemDetails{}).
		WithTags(TagResources)

	// HEAD to check existence
	rg.HEAD(ResourceIDPath, headResourceHandler).
		WithOperationID("checkResourceExists").
		WithPathParam(ParamResourceID, "", int(0)).
		WithNoContentResponse().
		WithResponse(http.StatusNotFound, mux.ProblemDetails{}).
		WithTags(TagResources)

	// GET a single resource
	rg.GET(ResourceIDPath, getResourceHandler).
		WithOperationID("getResource").
		WithPathParam(ParamResourceID, "", int(0)).
		WithOKResponse(Resource{}).
		WithResponse(http.StatusNotFound, mux.ProblemDetails{}).
		WithTags(TagResources)

	// Search by query params (single and repeated)
	rg.GET(ResourcesSearch, searchResourcesHandler).
		WithOperationID("searchResources").
		WithQueryParam("name", "", "").
		WithQueryParam("type", "", "").
		WithOKResponse([]Resource{}).
		WithTags(TagResources)

	// Bulk create resources via JSON array
	rg.POST(ResourcesBulk, createResourcesBulkHandler).
		WithOperationID("createResourcesBulk").
		WithJsonBody([]Resource{}).
		WithCreatedResponse([]Resource{}).
		WithResponse(http.StatusBadRequest, mux.ProblemDetails{}).
		WithTags(TagResources)

	// Update resource metadata to exercise map/object JSON bodies.
	rg.PUT(ResourceMeta, updateResourceMetadataHandler).
		WithOperationID("updateResourceMetadata").
		WithPathParam(ParamResourceID, "", int(0)).
		WithJsonBody(&struct {
			Metadata map[string]string `json:"metadata"`
		}{}).
		WithOKResponse(struct {
			Resource Resource          `json:"resource"`
			Metadata map[string]string `json:"metadata"`
		}{}).
		WithTags(TagResources)

	// UUID path parameter example
	rg.GET(ItemsUUIDPath, getItemByUUIDHandler).
		WithOperationID("getItemByUUID").
		WithPathParam(ParamItemID, "", uuid.Nil).
		WithOKResponse(map[string]uuid.UUID{}).
		WithTags(TagItems)

	// Simple header echo to exercise header lookup
	rg.GET(HeadersEchoPath, headersEchoHandler).
		WithOperationID("echoHeader").
		WithOKResponse("").
		WithTags(TagMisc)

	// Query multiple ints and UUIDs to exercise QueryInts/QueryUUIDs
	rg.GET(FilterPath, filterHandler).
		WithOperationID("filter").
		WithQueryParam("ids", "", "").
		WithQueryParam("uuids", "", uuid.Nil).
		WithOKResponse(struct {
			IDs   []int       `json:"ids"`
			UUIDs []uuid.UUID `json:"uuids"`
		}{}).
		WithTags(TagMisc)

	// Form submission example (Bind should support form-encoded bodies)
	rg.POST(FormSubmitPath, formSubmitHandler).
		WithOperationID("submitForm").
		WithOKResponse(map[string]string{}).
		WithTags(TagMisc)

	rg.StaticFallback("/**", "static", "static/index.html")

	// Tenant management routes (list, get, create, update, delete)
	rg.GET(TenantsPath+"/", listTenantsHandler).
		WithOperationID("listTenants").
		WithSummary("List all tenants").
		WithOKResponse([]Tenant{}).
		WithTags(TagTenants)

	rg.POST(TenantsPath+"/", createTenantHandler).
		WithOperationID("createTenant").
		WithJsonBody(Tenant{}).
		WithCreatedResponse(Tenant{}).
		WithTags(TagTenants)

	rg.GET(TenantIDPath, getTenantHandler).
		WithOperationID("getTenant").
		WithPathParam(ParamTenantID, "", int(0)).
		WithOKResponse(Tenant{}).
		WithResponse(http.StatusNotFound, mux.ProblemDetails{}).
		WithTags(TagTenants)

	rg.PUT(TenantIDPath, updateTenantHandler).
		WithOperationID("updateTenant").
		WithPathParam(ParamTenantID, "", int(0)).
		WithJsonBody(Tenant{}).
		WithOKResponse(Tenant{}).
		WithTags(TagTenants)

	rg.DELETE(TenantIDPath, deleteTenantHandler).
		WithOperationID("deleteTenant").
		WithPathParam(ParamTenantID, "", int(0)).
		WithNoContentResponse().
		WithResponse(http.StatusNotFound, mux.ProblemDetails{}).
		WithTags(TagTenants)

	// Tenant resources
	rg.GET(TenantResources, listTenantResourcesHandler).
		WithOperationID("listTenantResources").
		WithPathParam(ParamTenantID, "", int(0)).
		WithOKResponse([]Resource{}).
		WithResponse(http.StatusNotFound, mux.ProblemDetails{}).
		WithTags(TagTenants, TagResources)

	rg.POST(TenantResources, createTenantResourceHandler).
		WithOperationID("createTenantResource").
		WithPathParam(ParamTenantID, "", int(0)).
		WithJsonBody(Resource{}).
		WithCreatedResponse(Resource{}).
		WithTags(TagTenants, TagResources)
}

// Handlers extracted to reduce cognitive complexity of ConfigureRoutes.

func wsHandler(c mux.RouteContext) {
	if !strings.EqualFold(c.Request().Header.Get("Upgrade"), "websocket") {
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
}

func listResourcesHandler(c mux.RouteContext) {
	resources := Service.ListResources(0)
	if len(resources) == 0 {
		c.NotFound()
		return
	}
	c.OK(resources)
}

func headResourceHandler(c mux.RouteContext) {
	resourceIDStr, found := c.Param(ParamResourceID)
	if !found {
		c.NotFound()
		return
	}
	id64, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		c.BadRequest(ErrInvalidResourceID, ErrParseResourceID)
		return
	}
	if id64 < math.MinInt32 || id64 > math.MaxInt32 {
		c.BadRequest(ErrInvalidResourceID, ErrParseResourceID)
		return
	}
	resourceID := int32(id64)
	_, found = Service.GetResource(resourceID)
	if !found {
		c.NotFound()
		return
	}
	c.NoContent()
}

func getResourceHandler(c mux.RouteContext) {
	resourceIDStr, found := c.Param(ParamResourceID)
	if !found {
		c.NotFound()
		return
	}
	id64, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		c.BadRequest(ErrInvalidResourceID, ErrParseResourceID)
		return
	}
	if id64 < math.MinInt32 || id64 > math.MaxInt32 {
		c.BadRequest(ErrInvalidResourceID, ErrParseResourceID)
		return
	}
	resourceID := int32(id64)
	resource, found := Service.GetResource(resourceID)
	if !found {
		c.NotFound()
		return
	}
	c.OK(resource)
}

func typesMatch(rsrcType string, types []string) bool {
	if len(types) == 0 {
		return true
	}
	for _, t := range types {
		if rsrcType == t {
			return true
		}
	}
	return false
}

func searchResourcesHandler(c mux.RouteContext) {
	name, _ := c.Query().String("name")
	types, _ := c.Query().Strings("type")
	var out []Resource
	for _, rsrc := range Service.ListResources(0) {
		if name != "" && rsrc.Name != name {
			continue
		}
		if !typesMatch(rsrc.Type, types) {
			continue
		}
		out = append(out, *rsrc)
	}
	if len(out) == 0 {
		c.NotFound()
		return
	}
	c.OK(out)
}

func createResourcesBulkHandler(c mux.RouteContext) {
	var resources []Resource
	if err := c.Bind(&resources); err != nil {
		c.BadRequest(ErrBadRequest, err.Error())
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
}

func updateResourceMetadataHandler(c mux.RouteContext) {
	var body struct {
		Metadata map[string]string `json:"metadata"`
	}
	if err := c.Bind(&body); err != nil {
		c.BadRequest(ErrBadRequest, err.Error())
		return
	}
	resourceID, ok := c.ParamInt32(ParamResourceID)
	if !ok {
		c.BadRequest(ErrInvalidResourceID, ErrParseResourceID)
		return
	}
	rsrc, found := Service.GetResource(resourceID)
	if !found {
		c.NotFound()
		return
	}
	// For tests we simply echo back the received metadata along with the id.
	c.OK(map[string]any{"resource": rsrc, "metadata": body.Metadata})
}

func getItemByUUIDHandler(c mux.RouteContext) {
	id, ok := c.ParamUUID(ParamItemID)
	if !ok {
		c.BadRequest("Invalid UUID", "itemId is required and must be a UUID")
		return
	}
	c.OK(map[string]uuid.UUID{"id": id})
}

func headersEchoHandler(c mux.RouteContext) {
	if val, ok := c.Headers().String(common.HeaderXEcho); ok {
		c.Plain(http.StatusOK, []byte(val))
		return
	}
	c.NotFound()
}

func filterHandler(c mux.RouteContext) {
	ints, _ := c.Query().Ints("ids")
	uuids, _ := c.Query().UUIDs("uuids")
	c.OK(map[string]any{"ids": ints, "uuids": uuids})
}

func formSubmitHandler(c mux.RouteContext) {
	var body struct {
		Field string `json:"field"`
	}
	if err := c.Bind(&body); err != nil {
		c.BadRequest(ErrBadRequest, err.Error())
		return
	}
	c.OK(body)
}

func listTenantsHandler(c mux.RouteContext) {
	tenants := Service.ListTenants()
	if len(tenants) == 0 {
		c.NotFound()
		return
	}
	c.OK(tenants)
}

func createTenantHandler(c mux.RouteContext) {
	var t Tenant
	if err := c.Bind(&t); err != nil {
		c.BadRequest(ErrBadRequest, err.Error())
		return
	}
	created := Service.PutTenant(&t)
	c.Created(created)
}

func getTenantHandler(c mux.RouteContext) {
	id, ok := c.ParamInt(ParamTenantID)
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
}

func updateTenantHandler(c mux.RouteContext) {
	id, ok := c.ParamInt(ParamTenantID)
	if !ok {
		c.BadRequest(ErrInvalidTenantID, ErrTenantMissing)
		return
	}
	var t Tenant
	if err := c.Bind(&t); err != nil {
		c.BadRequest(ErrBadRequest, err.Error())
		return
	}
	t.TenantID = int32(id)
	updated := Service.PutTenant(&t)
	c.OK(updated)
}

func deleteTenantHandler(c mux.RouteContext) {
	id, ok := c.ParamInt(ParamTenantID)
	if !ok {
		c.BadRequest(ErrInvalidTenantID, ErrTenantMissing)
		return
	}
	deleted := Service.DeleteTenant(int32(id))
	if !deleted {
		c.NotFound()
		return
	}
	c.NoContent()
}

func listTenantResourcesHandler(c mux.RouteContext) {
	id, ok := c.ParamInt(ParamTenantID)
	if !ok {
		c.BadRequest(ErrInvalidTenantID, ErrTenantMissing)
		return
	}
	resources := Service.ListResources(int32(id))
	if len(resources) == 0 {
		c.NotFound()
		return
	}
	c.OK(resources)
}

func createTenantResourceHandler(c mux.RouteContext) {
	id, ok := c.ParamInt(ParamTenantID)
	if !ok {
		c.BadRequest(ErrInvalidTenantID, ErrTenantMissing)
		return
	}
	var res Resource
	if err := c.Bind(&res); err != nil {
		c.BadRequest(ErrBadRequest, err.Error())
		return
	}
	res.TenantID = int32(id)
	created := Service.PutResource(&res)
	c.Created(created)
}
