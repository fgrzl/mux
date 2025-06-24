package mux

import (
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fgrzl/json/jsonschema"
)

// routeData represents a route for OpenAPI generation.
type routeData struct {
	Path    string
	Method  string
	Options *RouteOptions
}

// collectRoutes collects all routes from a routeNode tree.
func collectRoutes(node *routeNode) ([]routeData, error) {
	if node == nil {
		return nil, fmt.Errorf("route node is nil")
	}
	var routes []routeData
	var walk func(string, *routeNode) error
	walk = func(prefix string, n *routeNode) error {
		for method, opt := range n.RouteOptions {
			if method == "" {
				return fmt.Errorf("empty method in route options at path %q", prefix)
			}
			routes = append(routes, routeData{
				Path:    cleanPath(prefix),
				Method:  strings.ToUpper(method),
				Options: opt,
			})
		}
		for seg, child := range n.Children {
			if seg == "" {
				return fmt.Errorf("empty segment in children at path %q", prefix)
			}
			if err := walk(path.Join(prefix, seg), child); err != nil {
				return err
			}
		}
		if n.ParamChild != nil {
			if n.ParamChild.ParamName == "" {
				return fmt.Errorf("empty param name at path %q", prefix)
			}
			if err := walk(path.Join(prefix, "{"+n.ParamChild.ParamName+"}"), n.ParamChild); err != nil {
				return err
			}
		}
		if n.Wildcard != nil {
			if err := walk(path.Join(prefix, "*"), n.Wildcard); err != nil {
				return err
			}
		}
		if n.CatchAll != nil {
			if err := walk(path.Join(prefix, "**"), n.CatchAll); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk("", node); err != nil {
		return nil, err
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})

	return routes, nil
}

// cleanPath ensures consistent path formatting.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	p = path.Clean("/" + strings.Trim(p, "/"))
	if p == "." {
		return "/"
	}
	return p
}

// Generator generates an OpenAPI 3.1 specification from a router.
type Generator struct {
	spec            *OpenAPISpec
	schemaGenerator func(reflect.Type, map[reflect.Type]bool) (*Schema, error)
}

// NewGenerator creates a new Generator instance.
func NewGenerator() *Generator {
	return &Generator{
		spec:            NewOpenAPISpec(),
		schemaGenerator: jsonSchemaGenerator,
	}
}

// WithSchemaGenerator sets a custom schema generator.
func (g *Generator) WithSchemaGenerator(fn func(reflect.Type, map[reflect.Type]bool) (*Schema, error)) *Generator {
	g.schemaGenerator = fn
	return g
}

// GenerateSpec generates an OpenAPI 3.1 specification from a router.
func (g *Generator) GenerateSpec(router *Router) (*OpenAPISpec, error) {
	if router == nil {
		return nil, fmt.Errorf("router is nil")
	}
	if router.options.openapi == nil {
		return nil, fmt.Errorf("router info is nil")
	}

	// Initialize spec
	g.spec.Info = *router.options.openapi
	if g.spec.Components == nil {
		g.spec.Components = &ComponentsObject{
			Schemas:         make(map[string]Schema),
			Responses:       make(map[string]ResponseObject),
			Parameters:      make(map[string]ParameterObject),
			Examples:        make(map[string]ExampleObject),
			RequestBodies:   make(map[string]RequestBodyObject),
			Headers:         make(map[string]HeaderObject),
			SecuritySchemes: make(map[string]SecurityScheme),
			Links:           make(map[string]LinkObject),
		}
	}

	// Collect routes
	routes, err := collectRoutes(router.registry.root)
	if err != nil {
		return nil, fmt.Errorf("collecting routes: %w", err)
	}

	// Process routes
	for _, rd := range routes {
		if rd.Options == nil || rd.Options.OperationID == "" {
			continue
		}
		if err := appendRouteToSpec(g.spec, rd, g.schemaGenerator); err != nil {
			return nil, fmt.Errorf("appending route %s %s: %w", rd.Method, rd.Path, err)
		}
	}

	// Validate spec
	if err := g.spec.Validate(); err != nil {
		return nil, fmt.Errorf("validating spec: %w", err)
	}

	return g.spec, nil
}

// GenerateAndSave generates and saves the spec as YAML or JSON.
func (g *Generator) GenerateAndSave(router *Router, outputPath string) error {
	spec, err := g.GenerateSpec(router)
	if err != nil {
		return err
	}
	return spec.MarshalToFile(outputPath)
}

// appendRouteToSpec appends a route to the OpenAPI spec, resolving schemas.
func appendRouteToSpec(
	spec *OpenAPISpec,
	rd routeData,
	schemaGen func(reflect.Type, map[reflect.Type]bool) (*Schema, error),
) error {
	if spec.Paths == nil {
		spec.Paths = make(map[string]PathItem)
	}
	path, method := rd.Path, strings.ToLower(rd.Method)
	opt := rd.Options

	if !isValidHTTPMethod(method) {
		return fmt.Errorf("invalid HTTP method %q", method)
	}
	if opt.OperationID == "" {
		return fmt.Errorf("operationID is required for %s %s", method, path)
	}
	if err := validatePathParameters(path, opt.Parameters); err != nil {
		return fmt.Errorf("validating path parameters: %w", err)
	}

	item, ok := spec.Paths[path]
	if !ok {
		item = PathItem{}
	}

	op := opt.Operation
	if len(op.Responses) == 0 {
		code := getDefaultResponseCode(method)
		op.Responses = map[string]ResponseObject{
			code: {Description: getDefaultResponseDescription(code)},
		}
	}

	for i, param := range op.Parameters {
		if err := ensureComponentSchema(spec, param.Example, param.Schema, schemaGen); err != nil {
			return fmt.Errorf("param %s: %w", param.Name, err)
		}
		if isZero(param.Example) {
			param.Example = nil
		}
		op.Parameters[i] = param
	}

	if op.RequestBody != nil {
		newContent := make(map[string]MediaType)
		for ctype, media := range op.RequestBody.Content {
			if err := ensureComponentSchema(spec, media.Example, media.Schema, schemaGen); err != nil {
				return fmt.Errorf("request body (%s): %w", ctype, err)
			}
			if isZero(media.Example) {
				media.Example = nil
			}
			newContent[ctype] = media
		}
		op.RequestBody = &RequestBodyObject{
			Content:     newContent,
			Required:    op.RequestBody.Required,
			Description: op.RequestBody.Description,
			Extensions:  op.RequestBody.Extensions,
		}
	}

	newResponses := make(map[string]ResponseObject)
	for code, resp := range op.Responses {
		if resp.Content != nil {
			newContent := make(map[string]MediaType)
			for ctype, media := range resp.Content {
				if err := ensureComponentSchema(spec, media.Example, media.Schema, schemaGen); err != nil {
					return fmt.Errorf("response %s (%s): %w", code, ctype, err)
				}
				if isZero(media.Example) {
					media.Example = nil
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
		item.Get = &op
	case "post":
		item.Post = &op
	case "put":
		item.Put = &op
	case "delete":
		item.Delete = &op
	case "patch":
		item.Patch = &op
	case "options":
		item.Options = &op
	case "head":
		item.Head = &op
	case "trace":
		item.Trace = &op
	default:
		return fmt.Errorf("unsupported method %q", method)
	}

	spec.Paths[path] = item
	return nil
}

func ensureComponentSchema(
	spec *OpenAPISpec,
	example any,
	schema *Schema,
	gen func(reflect.Type, map[reflect.Type]bool) (*Schema, error),
) error {
	if schema == nil || schema.Ref == "" {
		return nil
	}
	typeName := strings.TrimPrefix(schema.Ref, "#/components/schemas/")
	if _, exists := spec.Components.Schemas[typeName]; exists {
		return nil
	}
	t := reflect.TypeOf(example)
	if t == nil {
		return fmt.Errorf("missing type info for schema ref %q", schema.Ref)
	}
	genSchema, err := gen(t, map[reflect.Type]bool{})
	if err != nil {
		return fmt.Errorf("generating schema for %q: %w", typeName, err)
	}
	if !isZero(example) {
		schema.Example = example
	}
	spec.Components.Schemas[typeName] = *genSchema
	return nil
}

// ... rest unchanged

// jsonSchemaGenerator adapts jsonschema.GenerateSchema to the Generator interface.
func jsonSchemaGenerator(t reflect.Type, visited map[reflect.Type]bool) (*Schema, error) {
	if t == nil {
		return nil, fmt.Errorf("type is nil")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Name() == "" {
		return nil, fmt.Errorf("unnamed types cannot be registered")
	}
	if visited[t] {
		return &Schema{Ref: "#/components/schemas/" + t.Name()}, nil
	}
	visited[t] = true
	defer delete(visited, t)

	raw := jsonschema.GenerateSchema(t)
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshaling generated schema: %w", err)
	}

	var s Schema
	if err := json.Unmarshal(jsonBytes, &s); err != nil {
		return nil, fmt.Errorf("unmarshaling into Schema: %w", err)
	}

	// Optionally apply openapi tags if struct
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name, _ := parseJSONTag(f)
			if name == "-" || s.Properties[name] == nil {
				continue
			}
			applyOpenAPITags(f, s.Properties[name])
		}
	}

	return &s, nil
}

// applyOpenAPITags applies OpenAPI-specific tags to a schema.
func applyOpenAPITags(f reflect.StructField, schema *Schema) {
	tag := f.Tag.Get("openapi")
	if tag == "" {
		return
	}
	for _, part := range strings.Split(tag, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key, value := kv[0], kv[1]
		switch key {
		case "format":
			schema.Format = value
		case "pattern":
			schema.Pattern = value
		case "minimum":
			if min, err := strconv.ParseFloat(value, 64); err == nil {
				schema.Minimum = &min
			}
		case "maximum":
			if max, err := strconv.ParseFloat(value, 64); err == nil {
				schema.Maximum = &max
			}
		case "enum":
			vals := strings.Split(value, "|")
			schema.Enum = make([]interface{}, len(vals))
			for i, v := range vals {
				schema.Enum[i] = v
			}
		}
	}
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

	zero := reflect.Zero(val.Type())
	return reflect.DeepEqual(val.Interface(), zero.Interface())
}

// parseJSONTag extracts the JSON field name and required flag.
func parseJSONTag(f reflect.StructField) (name string, required bool) {
	jsonTag := f.Tag.Get("json")
	if jsonTag == "" {
		return f.Name, true
	}
	parts := strings.Split(jsonTag, ",")
	name = parts[0]
	if name == "" {
		name = f.Name
	}
	return name, len(parts) == 1 || parts[1] != "omitempty"
}

// validatePathParameters ensures path parameters match Parameters.
func validatePathParameters(path string, params []ParameterObject) error {
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

// isValidHTTPMethod checks if a method is valid.
func isValidHTTPMethod(method string) bool {
	validMethods := map[string]struct{}{
		"get":     {},
		"post":    {},
		"put":     {},
		"delete":  {},
		"patch":   {},
		"options": {},
		"head":    {},
		"trace":   {},
	}
	_, ok := validMethods[method]
	return ok
}

// getDefaultResponseDescription returns a default description.
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

// getDefaultResponseCode returns a default response code.
func getDefaultResponseCode(method string) string {
	switch method {
	case "post":
		return "201"
	case "delete":
		return "204"
	default:
		return "200"
	}
}
