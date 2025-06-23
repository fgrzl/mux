package mux

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// RouteBuilder defines a route with OpenAPI metadata.
type RouteBuilder struct {
	Method  string
	Pattern string
	Options *RouteOptions
}

// RouteOptions contains routing metadata and OpenAPI documentation.
type RouteOptions struct {
	// Routing metadata
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

	// OpenAPI documentation
	Operation
}

var opIDValidator = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// Route creates a new RouteBuilder for the given method and pattern.
func Route(method, pattern string) *RouteBuilder {
	rb := &RouteBuilder{
		Method:  strings.ToUpper(method),
		Pattern: pattern,
		Options: &RouteOptions{
			Operation: Operation{
				Responses: make(map[string]ResponseObject),
			},
		},
	}
	return rb
}

// AllowAnonymous allows anonymous access to the route.
func (rb *RouteBuilder) AllowAnonymous() *RouteBuilder {
	rb.Options.AllowAnonymous = true
	return rb
}

// RequirePermission adds required permissions.
func (rb *RouteBuilder) RequirePermission(resource string, permissions ...string) *RouteBuilder {
	rb.Options.Permissions = append(rb.Options.Permissions, permissions...)
	return rb
}

// RequireRoles adds required roles.
func (rb *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	rb.Options.Roles = append(rb.Options.Roles, roles...)
	return rb
}

// RequireScopes adds required scopes.
func (rb *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	rb.Options.Scopes = append(rb.Options.Scopes, scopes...)
	return rb
}

// WithRateLimit sets rate limiting.
func (rb *RouteBuilder) WithRateLimit(limit int, interval time.Duration) *RouteBuilder {
	rb.Options.RateLimit = limit
	rb.Options.RateInterval = interval
	return rb
}

// WithOperationID sets the OpenAPI OperationID.
func (rb *RouteBuilder) WithOperationID(id string) *RouteBuilder {
	if !opIDValidator.MatchString(id) {
		panic(fmt.Sprintf("invalid OperationID: %q — only alphanumeric and underscores allowed", id))
	}
	rb.Options.OperationID = id
	return rb
}

// WithParam adds an OpenAPI parameter using an example value.
func (rb *RouteBuilder) WithParam(name, in string, example any, required bool) *RouteBuilder {
	if name == "" || in == "" {
		panic(fmt.Sprintf("parameter name or 'in' cannot be empty"))
	}
	if !isValidParameterIn(in) {
		panic(fmt.Sprintf("invalid parameter 'in' value: %q", in))
	}
	schema, err := defaultSchemaGenerator(reflect.TypeOf(example), make(map[reflect.Type]bool))
	if err != nil {
		panic(fmt.Sprintf("invalid parameter type for %s: %v", name, err))
	}
	rb.Options.Parameters = append(rb.Options.Parameters, ParameterObject{
		Name:     name,
		In:       in,
		Required: required || in == "path",
		Schema:   schema,
		Example:  example,
	})
	return rb
}

// WithResponse adds an OpenAPI response using an example value.
func (rb *RouteBuilder) WithResponse(code int, example any) *RouteBuilder {
	resp := ResponseObject{}
	if example != nil {
		schema, err := defaultSchemaGenerator(reflect.TypeOf(example), make(map[reflect.Type]bool))
		if err != nil {
			panic(fmt.Sprintf("invalid response type for code %d: %v", code, err))
		}
		resp.Content = map[string]MediaType{
			"application/json": {Schema: schema, Example: example},
		}
	}
	if len(rb.Options.Responses) == 0 {
		rb.Options.Responses = make(map[string]ResponseObject)
	}

	rb.Options.Responses[fmt.Sprintf("%d", code)] = resp
	return rb
}

// WithJsonBody adds an OpenAPI request body with JSON content.
func (rb *RouteBuilder) WithJsonBody(example any) *RouteBuilder {
	return rb.withBody(example, "application/json")
}

// WithFormBody adds an OpenAPI request body with form content.
func (rb *RouteBuilder) WithFormBody(example any) *RouteBuilder {
	return rb.withBody(example, "application/x-www-form-urlencoded")
}

// WithMultipartBody adds an OpenAPI request body with multipart content.
func (rb *RouteBuilder) WithMultipartBody(example any) *RouteBuilder {
	return rb.withBody(example, "multipart/form-data")
}

func (rb *RouteBuilder) withBody(example any, contentType string) *RouteBuilder {
	if example == nil {
		return rb
	}
	schema, err := defaultSchemaGenerator(reflect.TypeOf(example), make(map[reflect.Type]bool))
	if err != nil {
		panic(fmt.Sprintf("invalid request body type: %v", err))
	}
	rb.Options.RequestBody = &RequestBodyObject{
		Content: map[string]MediaType{
			contentType: {Schema: schema, Example: example},
		},
		Required: true,
	}
	return rb
}

// WithSummary sets the OpenAPI summary.
func (rb *RouteBuilder) WithSummary(summary string) *RouteBuilder {
	rb.Options.Summary = summary
	return rb
}

// WithDescription sets the OpenAPI description.
func (rb *RouteBuilder) WithDescription(desc string) *RouteBuilder {
	rb.Options.Description = desc
	return rb
}

// WithTags sets the OpenAPI tags.
func (rb *RouteBuilder) WithTags(tags ...string) *RouteBuilder {
	rb.Options.Tags = append(rb.Options.Tags, tags...)
	return rb
}

// WithExternalDocs sets the OpenAPI external documentation.
func (rb *RouteBuilder) WithExternalDocs(url, desc string) *RouteBuilder {
	rb.Options.ExternalDocs = &ExternalDocumentation{URL: url, Description: desc}
	return rb
}

// WithSecurity adds an OpenAPI security requirement.
func (rb *RouteBuilder) WithSecurity(security SecurityRequirement) *RouteBuilder {
	rb.Options.Security = append(rb.Options.Security, security)
	return rb
}

// WithDeprecated marks the route as deprecated.
func (rb *RouteBuilder) WithDeprecated() *RouteBuilder {
	rb.Options.Deprecated = true
	return rb
}
