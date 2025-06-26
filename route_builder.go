package mux

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var defaultProblem = &ProblemDetails{}

type RouteBuilder struct {
	Options *RouteOptions
}

// RouteOptions holds both runtime routing data (handler, auth etc.) and the
// OpenAPI Operation object that will be rendered into the spec.
// Only a subset of the Operation fields are exposed via the builder; callers
// may still tweak the embedded Operation directly if needed.

type RouteOptions struct {
	// ---- runtime routing metadata ----
	Method         string
	Pattern        string
	Handler        HandlerFunc
	AllowAnonymous bool
	Roles          []string
	Scopes         []string
	Permissions    []string
	RateLimit      int
	RateInterval   time.Duration
	AuthProvider   AuthProvider

	// ---- OpenAPI documentation ----
	Operation
}

var opIDValidator = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Route bootstraps a new builder for the given HTTP method and path pattern.
func Route(method, pattern string) *RouteBuilder {
	return &RouteBuilder{
		Options: &RouteOptions{
			Method:    strings.ToUpper(method),
			Pattern:   pattern,
			Operation: Operation{Responses: map[string]*ResponseObject{}},
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
	schema, err := quickSchema(reflect.TypeOf(example))
	if err != nil {
		panic(err)
	}
	rb.Options.Parameters = append(rb.Options.Parameters, &ParameterObject{
		Name:     name,
		In:       in,
		Required: required || in == "path", // paths are always required
		Schema:   schema,
	})
	return rb
}

// WithResponse registers a response example and schema for the given HTTP code.
func (rb *RouteBuilder) WithResponse(code int, example any) *RouteBuilder {
	if rb.Options.Responses == nil {
		rb.Options.Responses = map[string]*ResponseObject{}
	}
	resp := &ResponseObject{}
	if example != nil {
		schema, err := quickSchema(reflect.TypeOf(example))
		if err != nil {
			panic(err)
		}
		resp.Content = map[string]*MediaType{
			"application/json": {Schema: schema, Example: example},
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
	return rb.WithResponse(http.StatusConflict, defaultProblem)
}

func (rb *RouteBuilder) WithBadRequestResponse() *RouteBuilder {
	return rb.WithResponse(http.StatusBadRequest, defaultProblem)
}

func (rb *RouteBuilder) WithStandardErrors() *RouteBuilder {
	return rb.WithBadRequestResponse().WithNotFoundResponse()
}

// WithJsonBody describes a JSON request body (required=true).
func (rb *RouteBuilder) WithJsonBody(example any) *RouteBuilder {
	return rb.withBody(example, "application/json")
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
	if example == nil {
		return rb
	}
	schema, err := quickSchema(reflect.TypeOf(example))
	if err != nil {
		panic(err)
	}
	rb.Options.RequestBody = &RequestBodyObject{
		Content:  map[string]*MediaType{ctype: {Schema: schema, Example: example}},
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
	rb.Options.ExternalDocs = &ExternalDocumentation{URL: url, Description: desc}
	return rb
}

// WithSecurity appends a security requirement.
func (rb *RouteBuilder) WithSecurity(sec *SecurityRequirement) *RouteBuilder {
	if rb.Options.Security == nil {
		rb.Options.Security = []*SecurityRequirement{}
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

// quickSchema is a lightweight type‑to‑schema helper used at build‑time.
// It intentionally avoids complex recursion — the full generator will later
// replace $refs with proper component registrations.
func quickSchema(t reflect.Type) (*Schema, error) {
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
		return &Schema{Ref: "#/components/schemas/" + name}, nil
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}, nil
	case reflect.Bool:
		return &Schema{Type: "boolean"}, nil
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}, nil
	case reflect.Slice, reflect.Array:
		items, err := quickSchema(t.Elem())
		if err != nil {
			return nil, err
		}
		return &Schema{Type: "array", Items: items}, nil
	default:
		return nil, fmt.Errorf("unsupported param kind %s", t.Kind())
	}
}

var knownSchemas = map[reflect.Type]*Schema{
	reflect.TypeOf(uuid.UUID{}): {Type: "string", Format: "uuid"},
	reflect.TypeOf(time.Time{}): {Type: "string", Format: "date-time"},
	reflect.TypeOf([]byte{}):    {Type: "string", Format: "byte"},
	reflect.TypeOf(net.IP{}):    {Type: "string", Format: "ipv4"},
	reflect.TypeOf(url.URL{}):   {Type: "string", Format: "uri"},
}

func RegisterSchema(t reflect.Type, schema *Schema) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	knownSchemas[t] = schema
}

func lookupKnownSchema(t reflect.Type) (*Schema, bool) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	schema, ok := knownSchemas[t]
	return schema, ok
}
