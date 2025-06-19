package mux

import (
	"fmt"
	"reflect"
	"regexp"
	"time"
)

type RouteBuilder struct {
	Method  string
	Pattern string
	Options *RouteOptions
}

type RouteParam struct {
	Name     string
	In       string // "path", "query", etc.
	Type     string
	Required bool
}

type SchemaRef struct {
	Type        reflect.Type
	IsArray     bool
	IsMap       bool
	KeyType     reflect.Type // only for maps
	ElemType    reflect.Type // for maps and slices
	Format      string
	OneOf       []*SchemaRef
	AllOf       []*SchemaRef
	Description string
	ContentType string // used for request/response
}

type RouteOptions struct {
	Method         string
	Pattern        string
	Handler        HandlerFunc
	AllowAnonymous bool
	Roles          []string
	Scopes         []string
	Permissions    []string
	RateLimit      int
	RateInterval   time.Duration

	// Documentation fields
	OperationID string
	Description string
	Summary     string
	Parameters  []RouteParam
	RequestBody *SchemaRef
	Responses   map[int]*SchemaRef

	// Dependencies
	AuthProvider AuthProvider
}

var (
	routes        []*RouteBuilder
	opIDValidator = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	opIDRegistry  = map[string]bool{}
)

func Route(method, pattern string) *RouteBuilder {
	rb := &RouteBuilder{
		Method:  method,
		Pattern: pattern,
		Options: &RouteOptions{},
	}
	routes = append(routes, rb)
	return rb
}

func (rb *RouteBuilder) AllowAnonymous() *RouteBuilder {
	rb.Options.AllowAnonymous = true
	return rb
}

func (rb *RouteBuilder) RequirePermission(resource string, permissions ...string) *RouteBuilder {
	rb.Options.Permissions = append(rb.Options.Permissions, permissions...)
	return rb
}

func (rb *RouteBuilder) RequireRoles(roles ...string) *RouteBuilder {
	rb.Options.Roles = append(rb.Options.Roles, roles...)
	return rb
}

func (rb *RouteBuilder) RequireScopes(scopes ...string) *RouteBuilder {
	rb.Options.Scopes = append(rb.Options.Scopes, scopes...)
	return rb
}

func (rb *RouteBuilder) WithRateLimit(limit int, interval time.Duration) *RouteBuilder {
	rb.Options.RateLimit = limit
	rb.Options.RateInterval = interval
	return rb
}

func (rb *RouteBuilder) WithOperationID(id string) *RouteBuilder {
	if !opIDValidator.MatchString(id) {
		panic(fmt.Sprintf("invalid OperationID: %q — only alphanumeric and underscores allowed", id))
	}
	if opIDRegistry[id] {
		panic(fmt.Sprintf("duplicate OperationID: %q", id))
	}
	opIDRegistry[id] = true
	rb.Options.OperationID = id
	return rb
}

func (rb *RouteBuilder) WithParam(name, in, typ string, required bool) *RouteBuilder {
	rb.Options.Parameters = append(rb.Options.Parameters, RouteParam{
		Name: name, In: in, Type: typ, Required: required,
	})
	return rb
}

func (rb *RouteBuilder) WithResponse(code int, example any) *RouteBuilder {
	if rb.Options.Responses == nil {
		rb.Options.Responses = make(map[int]*SchemaRef)
	}
	if example == nil {
		rb.Options.Responses[code] = nil
		return rb
	}
	rb.Options.Responses[code] = buildSchemaRef(example, "application/json")
	return rb
}

func (rb *RouteBuilder) WithJsonBody(example any) *RouteBuilder {
	return rb.withBody(example, "application/json")
}

func (rb *RouteBuilder) WithFormBody(example any) *RouteBuilder {
	return rb.withBody(example, "application/x-www-form-urlencoded")
}

func (rb *RouteBuilder) WithMultipartBody(example any) *RouteBuilder {
	return rb.withBody(example, "multipart/form-data")
}

func (rb *RouteBuilder) withBody(example any, contentType string) *RouteBuilder {
	if example == nil {
		return rb
	}
	rb.Options.RequestBody = buildSchemaRef(example, contentType)
	return rb
}

func (rb *RouteBuilder) WithSummary(summary string) *RouteBuilder {
	rb.Options.Summary = summary
	return rb
}

func (rb *RouteBuilder) WithDescription(desc string) *RouteBuilder {
	rb.Options.Description = desc
	return rb
}

func (rb *RouteBuilder) WithTags(tags ...string) *RouteBuilder {
	// Tags field can be added to RouteOptions if needed
	return rb
}

func buildSchemaRef(example any, contentType string) *SchemaRef {
	t := reflect.TypeOf(example)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := &SchemaRef{
		Type:        t,
		ContentType: contentType,
	}

	switch t.Kind() {
	case reflect.Slice, reflect.Array:
		schema.IsArray = true
		schema.ElemType = t.Elem()
	case reflect.Map:
		schema.IsMap = true
		schema.KeyType = t.Key()
		schema.ElemType = t.Elem()
	}

	return schema
}
