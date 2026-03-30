package builder

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fgrzl/mux/pkg/binder"
	"github.com/fgrzl/mux/pkg/common"
	"github.com/fgrzl/mux/pkg/openapi"
	"github.com/fgrzl/mux/pkg/routing"
	"github.com/google/uuid"
)

// RouteBuilder provides a fluent interface for configuring HTTP routes with OpenAPI documentation.
type RouteBuilder struct {
	Options *routing.RouteOptions
}

var opIDValidator = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Route bootstraps a new builder for the given HTTP method and path pattern.
func Route(method, pattern string) *RouteBuilder {
	return &RouteBuilder{
		Options: &routing.RouteOptions{
			Method:    strings.ToUpper(method),
			Pattern:   pattern,
			Operation: openapi.Operation{Responses: map[string]*openapi.ResponseObject{}},
		},
	}
}

// AllowAnonymous marks the route as accessible without authentication.
func (rb *RouteBuilder) AllowAnonymous() *RouteBuilder {
	rb.Options.AllowAnonymous = true
	return rb
}

// RequirePermission appends permission strings required to call this route.
func (rb *RouteBuilder) RequirePermission(perms ...string) *RouteBuilder {
	rb.Options.Permissions = append(rb.Options.Permissions, perms...)
	return rb
}

// RequireRoles appends role names required to call this route.
func (rb *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	rb.Options.Roles = append(rb.Options.Roles, roles...)
	return rb
}

// RequireScopes appends OAuth scopes required to call this route.
func (rb *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	rb.Options.Scopes = append(rb.Options.Scopes, scopes...)
	return rb
}

// WithRateLimit sets a simple token‑bucket limit on this route.
func (rb *RouteBuilder) WithRateLimit(limit int, interval time.Duration) *RouteBuilder {
	rb.Options.RateLimit = limit
	rb.Options.RateInterval = interval
	return rb
}

// WithOperationID sets/validates the OpenAPI OperationID.
func (rb *RouteBuilder) WithOperationID(id string) *RouteBuilder {
	if _, err := rb.WithOperationIDErr(id); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithOperationIDErr validates and sets the OpenAPI OperationID without panicking.
func (rb *RouteBuilder) WithOperationIDErr(id string) (*RouteBuilder, error) {
	if !opIDValidator.MatchString(id) {
		return rb, fmt.Errorf("invalid OperationID %q (only alnum + _ allowed)", id)
	}
	rb.Options.OperationID = id
	return rb, nil
}

// WithPathParam adds a required path parameter to this route.
//
// Parameters:
//   - name: The parameter name as it appears in the route pattern (e.g., "id" for "/users/{id}").
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type and for request binding.
//     For instance, pass uuid.Nil for UUID parameters, 0 for integers, or "" for strings.
//
// Path parameters are always marked as required in the OpenAPI spec.
func (rb *RouteBuilder) WithPathParam(name, description string, example any) *RouteBuilder {
	return rb.WithParam(name, "path", description, example, true)
}

// WithPathParamErr adds a required path parameter without panicking on validation failures.
func (rb *RouteBuilder) WithPathParamErr(name, description string, example any) (*RouteBuilder, error) {
	return rb.WithParamErr(name, "path", description, example, true)
}

// WithQueryParam adds an optional query parameter to this route.
//
// Parameters:
//   - name: The query parameter name (e.g., "limit" for "?limit=10").
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type and for request binding.
//     For instance, pass 10 for integer parameters, true for booleans, or "" for strings.
//
// Query parameters added via this method are marked as optional in the OpenAPI spec.
func (rb *RouteBuilder) WithQueryParam(name, description string, example any) *RouteBuilder {
	return rb.WithParam(name, "query", description, example, false)
}

// WithQueryParamErr adds an optional query parameter without panicking on validation failures.
func (rb *RouteBuilder) WithQueryParamErr(name, description string, example any) (*RouteBuilder, error) {
	return rb.WithParamErr(name, "query", description, example, false)
}

// WithRequiredQueryParam adds a required query parameter to this route.
//
// Parameters:
//   - name: The query parameter name (e.g., "apiKey" for "?apiKey=xyz").
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type and for request binding.
//     For instance, pass 10 for integer parameters, true for booleans, or "" for strings.
//
// Query parameters added via this method are marked as required in the OpenAPI spec.
func (rb *RouteBuilder) WithRequiredQueryParam(name, description string, example any) *RouteBuilder {
	return rb.WithParam(name, "query", description, example, true)
}

// WithRequiredQueryParamErr adds a required query parameter without panicking on validation failures.
func (rb *RouteBuilder) WithRequiredQueryParamErr(name, description string, example any) (*RouteBuilder, error) {
	return rb.WithParamErr(name, "query", description, example, true)
}

// WithHeaderParam adds a header parameter to this route.
//
// Parameters:
//   - name: The HTTP header name (e.g., "X-API-Version" or "Authorization").
//   - description: Human-readable explanation of the header for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type and for request binding.
//     For instance, pass "v1" for string headers or 1 for integer headers.
//   - required: If true, the header is marked as required in the OpenAPI spec;
//     if false, it's marked as optional.
func (rb *RouteBuilder) WithHeaderParam(name, description string, example any, required bool) *RouteBuilder {
	return rb.WithParam(name, "header", description, example, required)
}

// WithHeaderParamErr adds a header parameter without panicking on validation failures.
func (rb *RouteBuilder) WithHeaderParamErr(name, description string, example any, required bool) (*RouteBuilder, error) {
	return rb.WithParamErr(name, "header", description, example, required)
}

// WithCookieParam adds a cookie parameter to this route.
//
// Parameters:
//   - name: The cookie name (e.g., "sessionId" or "csrf_token").
//   - description: Human-readable explanation of the cookie for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type and for request binding.
//     For instance, pass "" for string cookies or 0 for integer cookies.
//   - required: If true, the cookie is marked as required in the OpenAPI spec;
//     if false, it's marked as optional.
func (rb *RouteBuilder) WithCookieParam(name, description string, example any, required bool) *RouteBuilder {
	return rb.WithParam(name, "cookie", description, example, required)
}

// WithCookieParamErr adds a cookie parameter without panicking on validation failures.
func (rb *RouteBuilder) WithCookieParamErr(name, description string, example any, required bool) (*RouteBuilder, error) {
	return rb.WithParamErr(name, "cookie", description, example, required)
}

// WithParam adds a parameter of any type/location to this route.
// This is a low-level method; prefer using WithPathParam, WithQueryParam, etc. for better clarity.
//
// Parameters:
//   - name: The parameter name (e.g., "id", "limit", "X-API-Key").
//   - in: The parameter location. Must be one of: "path", "query", "header", or "cookie".
//   - description: Human-readable explanation of the parameter for OpenAPI documentation.
//     Use an empty string ("") if no description is needed.
//   - example: Example value used to infer the OpenAPI schema type and for request binding.
//     The type of this value determines the schema (e.g., int → integer, uuid.UUID → string with format uuid).
//   - required: If true, the parameter is marked as required in the OpenAPI spec;
//     if false, it's marked as optional. Note: path parameters are always required regardless of this value.
//
// Panics if name or in is empty, or if in is not one of the valid parameter locations.
func (rb *RouteBuilder) WithParam(name, in, description string, example any, required bool) *RouteBuilder {
	if _, err := rb.WithParamErr(name, in, description, example, required); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithParamErr adds a parameter without panicking on validation failures.
func (rb *RouteBuilder) WithParamErr(name, in, description string, example any, required bool) (*RouteBuilder, error) {
	if name == "" || in == "" {
		return rb, fmt.Errorf("parameter name and 'in' cannot be empty")
	}
	if !isValidParameterIn(in) {
		return rb, fmt.Errorf("invalid parameter 'in': %q", in)
	}

	schema, err := QuickSchema(reflect.TypeOf(example))
	if err != nil {
		return rb, err
	}

	conv := binder.MakeConverter(reflect.TypeOf(example), schema)
	rb.Options.Parameters = append(rb.Options.Parameters, openapi.CloneParameterObject(&openapi.ParameterObject{
		Name:        name,
		In:          in,
		Description: description,
		Required:    required || in == "path",
		Schema:      schema,
		Example:     example,
		Converter:   conv,
	}))
	rb.Options.ParamIndex = routing.BuildParamIndex(rb.Options.Parameters)
	return rb, nil
}

// makeConverter is implemented in binding_convert.go

// WithResponse registers a response example and schema for the given HTTP code.
func (rb *RouteBuilder) WithResponse(code int, example any) *RouteBuilder {
	if _, err := rb.WithResponseErr(code, example); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithResponseErr registers a response example and schema without panicking.
func (rb *RouteBuilder) WithResponseErr(code int, example any) (*RouteBuilder, error) {
	if rb.Options.Responses == nil {
		rb.Options.Responses = map[string]*openapi.ResponseObject{}
	}
	resp := &openapi.ResponseObject{}
	if example != nil {
		schema, err := QuickSchema(reflect.TypeOf(example))
		if err != nil {
			return rb, err
		}
		resp.Content = map[string]*openapi.MediaType{
			common.MimeJSON: {Schema: schema, Example: example},
		}
	}
	rb.Options.Responses[fmt.Sprintf("%d", code)] = openapi.CloneResponseObject(resp)
	return rb, nil
}

func (rb *RouteBuilder) WithOKResponse(example any) *RouteBuilder {
	return rb.WithResponse(http.StatusOK, example)
}

func (rb *RouteBuilder) WithOKResponseErr(example any) (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusOK, example)
}

func (rb *RouteBuilder) WithCreatedResponse(example any) *RouteBuilder {
	return rb.WithResponse(http.StatusCreated, example)
}

func (rb *RouteBuilder) WithCreatedResponseErr(example any) (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusCreated, example)
}

func (rb *RouteBuilder) WithAcceptedResponse(example any) *RouteBuilder {
	return rb.WithResponse(http.StatusAccepted, example)
}

func (rb *RouteBuilder) WithAcceptedResponseErr(example any) (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusAccepted, example)
}

func (rb *RouteBuilder) WithNoContentResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusNoContent, nil)
}

func (rb *RouteBuilder) WithNoContentResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusNoContent, nil)
}

func (rb *RouteBuilder) WithNotFoundResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusNotFound, nil)
}

func (rb *RouteBuilder) WithNotFoundResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusNotFound, nil)
}

func (rb *RouteBuilder) WithConflictResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusConflict, common.DefaultProblem)
}

func (rb *RouteBuilder) WithConflictResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusConflict, common.DefaultProblem)
}

func (rb *RouteBuilder) WithBadRequestResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusBadRequest, common.DefaultProblem)
}

func (rb *RouteBuilder) WithBadRequestResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusBadRequest, common.DefaultProblem)
}

func (rb *RouteBuilder) WithStandardErrors() *RouteBuilder {
	return rb.WithBadRequestResponse().WithNotFoundResponse()
}

func (rb *RouteBuilder) WithStandardErrorsErr() (*RouteBuilder, error) {
	if _, err := rb.WithBadRequestResponseErr(); err != nil {
		return rb, err
	}
	return rb.WithNotFoundResponseErr()
}

// Redirect response methods
func (rb *RouteBuilder) With301Response() *RouteBuilder {
	return rb.WithResponse(http.StatusMovedPermanently, nil)
}

func (rb *RouteBuilder) With301ResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusMovedPermanently, nil)
}

func (rb *RouteBuilder) WithMovedPermanentlyResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusMovedPermanently, nil)
}

func (rb *RouteBuilder) WithMovedPermanentlyResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusMovedPermanently, nil)
}

func (rb *RouteBuilder) With302Response() *RouteBuilder {
	return rb.WithResponse(http.StatusFound, nil)
}

func (rb *RouteBuilder) With302ResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusFound, nil)
}

func (rb *RouteBuilder) WithFoundResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusFound, nil)
}

func (rb *RouteBuilder) WithFoundResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusFound, nil)
}

func (rb *RouteBuilder) With303Response() *RouteBuilder {
	return rb.WithResponse(http.StatusSeeOther, nil)
}

func (rb *RouteBuilder) With303ResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusSeeOther, nil)
}

func (rb *RouteBuilder) WithSeeOtherResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusSeeOther, nil)
}

func (rb *RouteBuilder) WithSeeOtherResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusSeeOther, nil)
}

func (rb *RouteBuilder) With307Response() *RouteBuilder {
	return rb.WithResponse(http.StatusTemporaryRedirect, nil)
}

func (rb *RouteBuilder) With307ResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusTemporaryRedirect, nil)
}

func (rb *RouteBuilder) WithTemporaryRedirectResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusTemporaryRedirect, nil)
}

func (rb *RouteBuilder) WithTemporaryRedirectResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusTemporaryRedirect, nil)
}

func (rb *RouteBuilder) With308Response() *RouteBuilder {
	return rb.WithResponse(http.StatusPermanentRedirect, nil)
}

func (rb *RouteBuilder) With308ResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusPermanentRedirect, nil)
}

func (rb *RouteBuilder) WithPermanentRedirectResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusPermanentRedirect, nil)
}

func (rb *RouteBuilder) WithPermanentRedirectResponseErr() (*RouteBuilder, error) {
	return rb.WithResponseErr(http.StatusPermanentRedirect, nil)
}

// WithJsonBody describes a JSON request body (required=true).
func (rb *RouteBuilder) WithJsonBody(example any) *RouteBuilder {
	if _, err := rb.WithJsonBodyErr(example); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithJsonBodyErr describes a JSON request body without panicking.
func (rb *RouteBuilder) WithJsonBodyErr(example any) (*RouteBuilder, error) {
	return rb.withBodyErr(example, common.MimeJSON)
}

// WithOneOfJsonBody describes a JSON request body using oneOf for polymorphic types.
// Pass example instances of each possible type as separate arguments.
func (rb *RouteBuilder) WithOneOfJsonBody(examples ...any) *RouteBuilder {
	if _, err := rb.WithOneOfJsonBodyErr(examples...); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithOneOfJsonBodyErr describes a JSON request body using oneOf without panicking.
func (rb *RouteBuilder) WithOneOfJsonBodyErr(examples ...any) (*RouteBuilder, error) {
	return rb.withCompositeBodyErr(examples, common.MimeJSON, "oneOf")
}

// WithAnyOfJsonBody describes a JSON request body using anyOf for polymorphic types.
// Pass example instances of each possible type as separate arguments.
func (rb *RouteBuilder) WithAnyOfJsonBody(examples ...any) *RouteBuilder {
	if _, err := rb.WithAnyOfJsonBodyErr(examples...); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithAnyOfJsonBodyErr describes a JSON request body using anyOf without panicking.
func (rb *RouteBuilder) WithAnyOfJsonBodyErr(examples ...any) (*RouteBuilder, error) {
	return rb.withCompositeBodyErr(examples, common.MimeJSON, "anyOf")
}

// WithAllOfJsonBody describes a JSON request body using allOf for composition.
// Pass example instances of each schema to compose as separate arguments.
func (rb *RouteBuilder) WithAllOfJsonBody(examples ...any) *RouteBuilder {
	if _, err := rb.WithAllOfJsonBodyErr(examples...); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithAllOfJsonBodyErr describes a JSON request body using allOf without panicking.
func (rb *RouteBuilder) WithAllOfJsonBodyErr(examples ...any) (*RouteBuilder, error) {
	return rb.withCompositeBodyErr(examples, common.MimeJSON, "allOf")
}

// WithFormBody describes an urlencoded form body.
func (rb *RouteBuilder) WithFormBody(example any) *RouteBuilder {
	if _, err := rb.WithFormBodyErr(example); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithFormBodyErr describes an urlencoded form body without panicking.
func (rb *RouteBuilder) WithFormBodyErr(example any) (*RouteBuilder, error) {
	return rb.withBodyErr(example, common.MimeFormURLEncoded)
}

// WithMultipartBody describes a multipart form body.
func (rb *RouteBuilder) WithMultipartBody(example any) *RouteBuilder {
	if _, err := rb.WithMultipartBodyErr(example); err != nil {
		panic(err.Error())
	}
	return rb
}

// WithMultipartBodyErr describes a multipart form body without panicking.
func (rb *RouteBuilder) WithMultipartBodyErr(example any) (*RouteBuilder, error) {
	return rb.withBodyErr(example, common.MimeMultipartFormData)
}

func (rb *RouteBuilder) withBody(example any, ctype string) *RouteBuilder {
	if _, err := rb.withBodyErr(example, ctype); err != nil {
		panic(err.Error())
	}
	return rb
}

func (rb *RouteBuilder) withBodyErr(example any, ctype string) (*RouteBuilder, error) {

	method := rb.Options.Method
	if method == http.MethodHead || method == http.MethodGet || method == http.MethodDelete {
		return rb, fmt.Errorf("HTTP method %s does not support a request body", method)
	}

	if example == nil {
		return rb, nil
	}
	schema, err := QuickSchema(reflect.TypeOf(example))
	if err != nil {
		return rb, err
	}
	rb.Options.RequestBody = openapi.CloneRequestBodyObject(&openapi.RequestBodyObject{
		Content:  map[string]*openapi.MediaType{ctype: {Schema: schema, Example: example}},
		Required: true,
	})
	return rb, nil
}

func (rb *RouteBuilder) withCompositeBody(examples []any, ctype string, compositionType string) *RouteBuilder {
	if _, err := rb.withCompositeBodyErr(examples, ctype, compositionType); err != nil {
		panic(err.Error())
	}
	return rb
}

func (rb *RouteBuilder) withCompositeBodyErr(examples []any, ctype string, compositionType string) (*RouteBuilder, error) {
	method := rb.Options.Method
	if method == http.MethodHead || method == http.MethodGet || method == http.MethodDelete {
		return rb, fmt.Errorf("HTTP method %s does not support a request body", method)
	}

	if len(examples) == 0 {
		return rb, nil
	}

	schemas := make([]*openapi.Schema, 0, len(examples))

	for _, example := range examples {
		if example == nil {
			continue
		}
		schema, err := QuickSchema(reflect.TypeOf(example))
		if err != nil {
			return rb, err
		}
		// Attach the example directly to the schema so it's available during generation
		schema.Example = example
		schemas = append(schemas, schema)
	}

	if len(schemas) == 0 {
		return rb, nil
	}

	compositeSchema := &openapi.Schema{}
	switch compositionType {
	case "oneOf":
		compositeSchema.OneOf = schemas
	case "anyOf":
		compositeSchema.AnyOf = schemas
	case "allOf":
		compositeSchema.AllOf = schemas
	default:
		return rb, fmt.Errorf("unsupported composition type: %s", compositionType)
	}

	// Use the first non-nil example as the overall example value for the media type
	var exampleValue any
	for _, ex := range examples {
		if ex != nil {
			exampleValue = ex
			break
		}
	}

	rb.Options.RequestBody = openapi.CloneRequestBodyObject(&openapi.RequestBodyObject{
		Content:  map[string]*openapi.MediaType{ctype: {Schema: compositeSchema, Example: exampleValue}},
		Required: true,
	})
	return rb, nil
}

// WithSummary sets the summary.
func (rb *RouteBuilder) WithSummary(s string) *RouteBuilder {
	rb.Options.Summary = s
	return rb
}

// WithDescription sets the description.
func (rb *RouteBuilder) WithDescription(d string) *RouteBuilder {
	rb.Options.Description = d
	return rb
}

// WithTags appends tags.
func (rb *RouteBuilder) WithTags(tags ...string) *RouteBuilder {
	if rb.Options.Tags == nil {
		rb.Options.Tags = []string{}
	}
	rb.Options.Tags = append(rb.Options.Tags, tags...)
	return rb
}

// WithExternalDocs adds external docs.
func (rb *RouteBuilder) WithExternalDocs(url, desc string) *RouteBuilder {
	rb.Options.ExternalDocs = &openapi.ExternalDocumentation{URL: url, Description: desc}
	return rb
}

// WithSecurity appends a security requirement.
func (rb *RouteBuilder) WithSecurity(sec *openapi.SecurityRequirement) *RouteBuilder {
	if rb.Options.Security == nil {
		rb.Options.Security = []*openapi.SecurityRequirement{}
	}
	rb.Options.Security = append(rb.Options.Security, openapi.CloneSecurityRequirement(sec))
	return rb
}

// WithDeprecated marks the route deprecated.
func (rb *RouteBuilder) WithDeprecated() *RouteBuilder {
	rb.Options.Deprecated = true
	return rb
}

// ---------- helpers ----------

func isValidParameterIn(in string) bool {
	switch in {
	case "query", "header", "path", "cookie":
		return true
	default:
		return false
	}
}

// QuickSchema is a lightweight type‑to‑schema helper used at build‑time.
// It intentionally avoids complex recursion — the full generator will later
// replace $refs with proper component registrations.
func QuickSchema(t reflect.Type) (*openapi.Schema, error) {
	if t == nil {
		return nil, fmt.Errorf("nil type")
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// handle known types
	if schema, ok := lookupKnownSchema(t); ok {
		return schema, nil
	}

	if name := t.Name(); name != "" && t.Kind() == reflect.Struct {
		return &openapi.Schema{Ref: "#/components/schemas/" + name}, nil
	}

	// handle anonymous structs by generating inline schema
	if t.Kind() == reflect.Struct && t.Name() == "" {
		return generateInlineStructSchema(t)
	}

	switch t.Kind() {
	case reflect.String:
		return &openapi.Schema{Type: "string"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &openapi.Schema{Type: "integer"}, nil
	case reflect.Bool:
		return &openapi.Schema{Type: "boolean"}, nil
	case reflect.Float32, reflect.Float64:
		return &openapi.Schema{Type: "number"}, nil
	case reflect.Slice, reflect.Array:
		items, err := QuickSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		return &openapi.Schema{Type: "array", Items: items}, nil
	case reflect.Map:
		// Handle map types
		keyType := t.Key()
		valueType := t.Elem()

		// Support string keys and numeric keys (which JSON converts to strings)
		switch keyType.Kind() {
		case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// These key types are valid - JSON converts numeric keys to strings
		default:
			return nil, fmt.Errorf("unsupported map key type %s, only string and numeric keys are supported", keyType.Kind())
		}

		// Generate schema for the value type
		valueSchema, err := QuickSchema(valueType)
		if err != nil {
			return nil, fmt.Errorf("generating schema for map value type: %w", err)
		}

		return &openapi.Schema{
			Type:                 "object",
			AdditionalProperties: valueSchema,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported param kind %s", t.Kind())
	}
}

var knownSchemas = map[reflect.Type]*openapi.Schema{
	reflect.TypeOf([]byte{}):          {Type: "string", Format: "byte"},
	reflect.TypeOf(json.RawMessage{}): {},
	reflect.TypeOf(net.IP{}):          {Type: "string", Format: "ipv4"},
	reflect.TypeOf(time.Time{}):       {Type: "string", Format: "date-time"},
	reflect.TypeOf(url.URL{}):         {Type: "string", Format: "uri"},
	reflect.TypeOf(uuid.UUID{}):       {Type: "string", Format: "uuid"},
}

func RegisterSchema(t reflect.Type, schema *openapi.Schema) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	knownSchemas[t] = openapi.CloneSchema(schema)
}

// RemoveSchema removes a schema from the knownSchemas registry (for test cleanup).
func RemoveSchema(t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	delete(knownSchemas, t)
}

func lookupKnownSchema(t reflect.Type) (*openapi.Schema, bool) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	schema, ok := knownSchemas[t]
	if !ok {
		return nil, false
	}
	return openapi.CloneSchema(schema), true
}

// generateInlineStructSchema creates an inline schema for anonymous structs
func generateInlineStructSchema(t reflect.Type) (*openapi.Schema, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct type, got %s", t.Kind())
	}

	schema := &openapi.Schema{
		Type:       "object",
		Properties: make(map[string]*openapi.Schema),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get the JSON tag name, or use the field name if no tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue // Skip fields marked with json:"-"
		}

		fieldName := field.Name
		if jsonTag != "" {
			// Handle json tags like "name,omitempty" - take only the name part
			if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
				fieldName = jsonTag[:commaIndex]
			} else {
				fieldName = jsonTag
			}
		}

		// Generate schema for the field type
		fieldSchema, err := QuickSchema(field.Type)
		if err != nil {
			return nil, fmt.Errorf("generating schema for field %s: %w", field.Name, err)
		}

		schema.Properties[fieldName] = fieldSchema
	}

	return schema, nil
}
