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
	if !opIDValidator.MatchString(id) {
		panic(fmt.Sprintf("invalid OperationID %q (only alnum + _ allowed)", id))
	}
	rb.Options.OperationID = id
	return rb
}

// WithPathParam adds a required path parameter with a string example.
func (rb *RouteBuilder) WithPathParam(name string, example any) *RouteBuilder {
	return rb.WithParam(name, "path", example, true)
}

// WithQueryParam adds an optional query parameter with example value.
func (rb *RouteBuilder) WithQueryParam(name string, example any) *RouteBuilder {
	return rb.WithParam(name, "query", example, false)
}

// WithRequiredQueryParam adds a required query parameter with example value.
func (rb *RouteBuilder) WithRequiredQueryParam(name string, example any) *RouteBuilder {
	return rb.WithParam(name, "query", example, true)
}

// WithHeaderParam adds a header parameter with example value.
func (rb *RouteBuilder) WithHeaderParam(name string, example any, required bool) *RouteBuilder {
	return rb.WithParam(name, "header", example, required)
}

// WithCookieParam adds a cookie parameter with example value.
func (rb *RouteBuilder) WithCookieParam(name string, example any, required bool) *RouteBuilder {
	return rb.WithParam(name, "cookie", example, required)
}

// WithParam describes an input parameter and infers its schema from the example
// value.  "in" must be one of query|header|path|cookie.
func (rb *RouteBuilder) WithParam(name, in string, example any, required bool) *RouteBuilder {
	if name == "" || in == "" {
		panic("parameter name and 'in' cannot be empty")
	}
	if !isValidParameterIn(in) {
		panic(fmt.Sprintf("invalid parameter 'in': %q", in))
	}
	schema, err := QuickSchema(reflect.TypeOf(example))
	if err != nil {
		panic(err)
	}
	conv := binder.MakeConverter(reflect.TypeOf(example), schema)
	rb.Options.Parameters = append(rb.Options.Parameters, &openapi.ParameterObject{
		Name:      name,
		In:        in,
		Required:  required || in == "path", // paths are always required
		Schema:    schema,
		Example:   example,
		Converter: conv,
	})
	// Keep ParamIndex in sync for fast lookups at runtime
	rb.Options.ParamIndex = routing.BuildParamIndex(rb.Options.Parameters)
	return rb
}

// makeConverter is implemented in binding_convert.go

// WithResponse registers a response example and schema for the given HTTP code.
func (rb *RouteBuilder) WithResponse(code int, example any) *RouteBuilder {
	if rb.Options.Responses == nil {
		rb.Options.Responses = map[string]*openapi.ResponseObject{}
	}
	resp := &openapi.ResponseObject{}
	if example != nil {
		schema, err := QuickSchema(reflect.TypeOf(example))
		if err != nil {
			panic(err)
		}
		resp.Content = map[string]*openapi.MediaType{
			common.MimeJSON: {Schema: schema, Example: example},
		}
	}
	rb.Options.Responses[fmt.Sprintf("%d", code)] = resp
	return rb
}

func (rb *RouteBuilder) WithOKResponse(example any) *RouteBuilder {
	return rb.WithResponse(http.StatusOK, example)
}

func (rb *RouteBuilder) WithCreatedResponse(example any) *RouteBuilder {
	return rb.WithResponse(http.StatusCreated, example)
}

func (rb *RouteBuilder) WithAcceptedResponse(example any) *RouteBuilder {
	return rb.WithResponse(http.StatusAccepted, example)
}

func (rb *RouteBuilder) WithNoContentResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusNoContent, nil)
}

func (rb *RouteBuilder) WithNotFoundResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusNotFound, nil)
}

func (rb *RouteBuilder) WithConflictResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusConflict, common.DefaultProblem)
}

func (rb *RouteBuilder) WithBadRequestResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusBadRequest, common.DefaultProblem)
}

func (rb *RouteBuilder) WithStandardErrors() *RouteBuilder {
	return rb.WithBadRequestResponse().WithNotFoundResponse()
}

// WithJsonBody describes a JSON request body (required=true).
func (rb *RouteBuilder) WithJsonBody(example any) *RouteBuilder {
	return rb.withBody(example, common.MimeJSON)
}

// WithFormBody describes an urlencoded form body.
func (rb *RouteBuilder) WithFormBody(example any) *RouteBuilder {
	return rb.withBody(example, "application/x-www-form-urlencoded")
}

// WithMultipartBody describes a multipart form body.
func (rb *RouteBuilder) WithMultipartBody(example any) *RouteBuilder {
	return rb.withBody(example, "multipart/form-data")
}

func (rb *RouteBuilder) withBody(example any, ctype string) *RouteBuilder {

	method := rb.Options.Method
	if method == http.MethodHead || method == http.MethodGet || method == http.MethodDelete {
		panic(fmt.Sprintf("HTTP method %s does not support a request body", method))
	}

	if example == nil {
		return rb
	}
	schema, err := QuickSchema(reflect.TypeOf(example))
	if err != nil {
		panic(err)
	}
	rb.Options.RequestBody = &openapi.RequestBodyObject{
		Content:  map[string]*openapi.MediaType{ctype: {Schema: schema, Example: example}},
		Required: true,
	}
	return rb
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
	rb.Options.Security = append(rb.Options.Security, sec)
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
	knownSchemas[t] = schema
}

// removeSchema removes a schema from the knownSchemas registry (for test cleanup).
func removeSchema(t reflect.Type) {
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
	return schema, ok
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
