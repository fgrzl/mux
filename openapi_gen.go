// Updated mux generator to ensure nested components are registered without global state
package mux

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/fgrzl/json/jsonschema"
	"github.com/google/uuid"
)

type GeneratorOption func(*Generator)

func WithExamples() GeneratorOption {
	return func(g *Generator) {
		g.withExamples = true
	}
}

// Generator generates an OpenAPI 3.1 specification and holds internal state.
type Generator struct {
	spec         *OpenAPISpec
	builder      *jsonschema.Builder
	visited      map[reflect.Type]bool
	withExamples bool // default is false
}

func NewGenerator(opts ...GeneratorOption) *Generator {
	gen := &Generator{
		spec:    NewOpenAPISpec(),
		builder: jsonschema.NewBuilder(),
		visited: make(map[reflect.Type]bool),
	}

	for _, opt := range opts {
		opt(gen)
	}

	return gen
}

func (g *Generator) GenerateSpec(router *Router) (*OpenAPISpec, error) {
	if router == nil || router.options.openapi == nil {
		return nil, fmt.Errorf("router or router info is nil")
	}

	g.spec.Info = router.options.openapi
	g.ensureComponentInit()

	routes, err := collectRoutes(router.registry.root)
	if err != nil {
		return nil, err
	}
	for _, rd := range routes {
		if rd.Options == nil || rd.Options.OperationID == "" {
			continue
		}
		if err := g.appendRoute(rd); err != nil {
			return nil, err
		}
	}
	return g.spec, g.spec.Validate()
}

func (g *Generator) GenerateAndSave(router *Router, path string) error {
	spec, err := g.GenerateSpec(router)
	if err != nil {
		return err
	}
	return spec.MarshalToFile(path)
}

func (g *Generator) ensureComponentInit() {
	if g.spec.Components == nil {
		g.spec.Components = &ComponentsObject{
			Schemas: make(map[string]*Schema),
		}
	}
}

func (g *Generator) appendRoute(rd routeData) error {
	if g.spec.Paths == nil {
		g.spec.Paths = map[string]*PathItem{}
	}

	path, method := rd.Path, strings.ToLower(rd.Method)
	opt := rd.Options
	if err := validatePathParameters(path, opt.Parameters); err != nil {
		return err
	}

	item := g.spec.Paths[path]
	if item == nil {
		item = new(PathItem)
	}
	op := &opt.Operation
	if len(op.Responses) == 0 {
		code := getDefaultResponseCode(method)
		op.Responses = map[string]*ResponseObject{code: {Description: getDefaultResponseDescription(code)}}
	}

	for i, param := range op.Parameters {
		if err := g.ensureComponentSchema(param.Example, param.Schema); err != nil {
			return err
		}
		if !g.withExamples {
			param.Example = nil
			param.Examples = nil
		}
		op.Parameters[i] = param
	}

	if op.RequestBody != nil {
		newContent := map[string]*MediaType{}
		for ctype, media := range op.RequestBody.Content {
			if err := g.ensureComponentSchema(media.Example, media.Schema); err != nil {
				return err
			}
			if !g.withExamples {
				media.Example = nil
				media.Examples = nil
			}
			newContent[ctype] = media
		}
		op.RequestBody.Content = newContent
	}

	newResponses := map[string]*ResponseObject{}
	for code, resp := range op.Responses {
		if resp.Content != nil {
			newContent := map[string]*MediaType{}
			for ctype, media := range resp.Content {
				if err := g.ensureComponentSchema(media.Example, media.Schema); err != nil {
					return err
				}
				if !g.withExamples {
					media.Example = nil
					media.Examples = nil
				}
				newContent[ctype] = media
			}
			resp.Content = newContent
		}
		if resp.Description == "" {
			resp.Description = getDefaultResponseDescription(code)
		}
		newResponses[code] = resp
	}
	op.Responses = newResponses

	switch method {
	case "get":
		item.Get = op
	case "post":
		item.Post = op
	case "put":
		item.Put = op
	case "delete":
		item.Delete = op
	case "patch":
		item.Patch = op
	}
	g.spec.Paths[path] = item
	return nil
}

func (g *Generator) ensureComponentSchema(example any, schema *Schema) error {
	if schema == nil || schema.Ref == "" {
		return nil
	}
	typeName := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
	if _, exists := g.spec.Components.Schemas[typeName]; exists {
		return nil
	}
	t := reflect.TypeOf(example)
	if t == nil {
		return fmt.Errorf("missing type info for schema ref %q", schema.Ref)
	}
	s, err := g.GenerateSchemaForType(t)
	if err != nil {
		return err
	}

	if g.withExamples && !isZero(example) {
		schema.Example = example
	}

	g.spec.Components.Schemas[typeName] = s

	return nil
}

func (g *Generator) GenerateSchemaForType(t reflect.Type) (*Schema, error) {
	if t == nil {
		return nil, fmt.Errorf("type is nil")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Name() == "" {
		return nil, fmt.Errorf("unnamed types cannot be registered")
	}
	if g.visited[t] {
		return &Schema{Ref: "#/components/schemas/" + t.Name()}, nil
	}
	g.visited[t] = true
	defer delete(g.visited, t)

	root, components := g.builder.SchemaWithComponents(t)

	for name, raw := range components {
		if _, exists := g.spec.Components.Schemas[name]; !exists {
			data, err := json.Marshal(raw)
			if err != nil {
				return nil, fmt.Errorf("marshal component schema %q: %w", name, err)
			}
			s := &Schema{}
			if err := json.Unmarshal(data, &s); err != nil {
				return nil, fmt.Errorf("unmarshal component schema %q: %w", name, err)
			}
			g.spec.Components.Schemas[name] = s
		}
	}

	data, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal root schema: %w", err)
	}
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal root schema: %w", err)
	}

	return &schema, nil
}

func isZero(v any) bool {
	if v == nil {
		return true
	}
	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return true
	}
	switch val.Kind() {
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	case reflect.Slice, reflect.Map, reflect.Array, reflect.String:
		return val.Len() == 0
	}
	if u, ok := v.(uuid.UUID); ok {
		return u == uuid.Nil
	}
	zero := reflect.Zero(val.Type())
	return reflect.DeepEqual(val.Interface(), zero.Interface())
}

func validatePathParameters(path string, params []*ParameterObject) error {
	pathParams := map[string]bool{}
	for _, match := range regexp.MustCompile(`\{([^}]+)\}`).FindAllStringSubmatch(path, -1) {
		pathParams[match[1]] = true
	}
	for _, p := range params {
		if p.In == "" || p.In == "path" {
			if !pathParams[p.Name] {
				return fmt.Errorf("path parameter %q not found in path %q", p.Name, path)
			}
			delete(pathParams, p.Name)
		}
	}
	if len(pathParams) > 0 {
		keys := make([]string, 0, len(pathParams))
		for k := range pathParams {
			keys = append(keys, k)
		}
		return fmt.Errorf("path parameters %v not defined in parameters for path %q", keys, path)
	}
	return nil
}

func getDefaultResponseCode(method string) string {
	switch method {
	case http.MethodPost:
		return "201"
	case "delete":
		return "204"
	default:
		return "200"
	}
}

func getDefaultResponseDescription(code string) string {
	switch code {
	case "200":
		return "Success"
	case "201":
		return "Created"
	case "204":
		return "No Content"
	case "400":
		return "Bad Request"
	case "401":
		return "Unauthorized"
	case "403":
		return "Forbidden"
	case "404":
		return "Not Found"
	case "500":
		return "Internal Server Error"
	default:
		return fmt.Sprintf("Response %s", code)
	}
}
