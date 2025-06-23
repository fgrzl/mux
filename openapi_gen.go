package mux

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Generator generates an OpenAPI 3.1 specification from a router.
type Generator struct {
	spec            *OpenAPISpec
	schemaGenerator func(reflect.Type, map[reflect.Type]bool) (*Schema, error)
	operationIDs    map[string]bool
}

// NewGenerator creates a new Generator instance.
func NewGenerator() *Generator {
	return &Generator{
		spec:            NewOpenAPISpec(),
		schemaGenerator: defaultSchemaGenerator,
		operationIDs:    make(map[string]bool),
	}
}

// WithSchemaGenerator sets a custom schema generator.
func (g *Generator) WithSchemaGenerator(fn func(reflect.Type, map[reflect.Type]bool) (*Schema, error)) *Generator {
	g.schemaGenerator = fn
	return g
}

// GenerateSpec generates an OpenAPI 3.1 specification.
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
		if err := appendRouteToSpec(g.spec, rd); err != nil {
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

// GenerateSwaggerUI generates an HTML file with Swagger UI.
func (g *Generator) GenerateSwaggerUI(outputPath, specPath string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4/swagger-ui.css">
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4/swagger-ui-bundle.js"></script>
    <script>
        SwaggerUIBundle({ url: %q, dom_id: '#swagger-ui' })
    </script>
</body>
</html>
`, specPath)
	return os.WriteFile(outputPath, []byte(html), 0644)
}

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

// appendRouteToSpec appends a route to the OpenAPI spec.
func appendRouteToSpec(spec *OpenAPISpec, rd routeData) error {
	if spec.Paths == nil {
		spec.Paths = make(map[string]PathItem)
	}
	path, method := rd.Path, strings.ToLower(rd.Method)
	opt := rd.Options

	// Validate inputs
	if !isValidHTTPMethod(method) {
		return fmt.Errorf("invalid HTTP method %q", method)
	}
	if opt.OperationID == "" {
		return fmt.Errorf("operationID is required for %s %s", method, path)
	}

	// Validate path parameters
	if err := validatePathParameters(path, opt.Parameters); err != nil {
		return fmt.Errorf("validating path parameters for %s %s: %w", method, path, err)
	}

	// Initialize PathItem
	item, ok := spec.Paths[path]
	if !ok {
		item = PathItem{}
	}

	// Create Operation (direct copy from RouteOptions)
	op := opt.Operation
	if len(op.Responses) == 0 {
		defaultCode := getDefaultResponseCode(method)
		op.Responses[fmt.Sprintf("%d", defaultCode)] = ResponseObject{
			Description: getDefaultResponseDescription(defaultCode),
		}
	}

	// Assign Operation to PathItem
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
		return fmt.Errorf("unsupported HTTP method %q", method)
	}

	spec.Paths[path] = item
	return nil
}

// defaultSchemaGenerator generates a JSON Schema from a Go type.
func defaultSchemaGenerator(t reflect.Type, visited map[reflect.Type]bool) (*Schema, error) {
	if t == nil {
		return nil, fmt.Errorf("type is nil")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeName := t.Name()
	if typeName != "" && visited[t] {
		return &Schema{Ref: "#/components/schemas/" + typeName}, nil
	}
	visited[t] = true
	defer delete(visited, t)

	schema := &Schema{}
	switch t.Kind() {
	case reflect.Struct:
		schema.Type = "object"
		schema.Properties = make(map[string]*Schema)
		var required []string
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name, requiredFlag := parseJSONTag(f)
			if name == "-" {
				continue
			}
			fieldSchema, err := defaultSchemaGenerator(f.Type, visited)
			if err != nil {
				return nil, fmt.Errorf("generating schema for field %s: %w", f.Name, err)
			}
			applyOpenAPITags(f, fieldSchema)
			schema.Properties[name] = fieldSchema
			if requiredFlag {
				required = append(required, name)
			}
		}
		if len(required) > 0 {
			schema.Required = required
		}
	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		itemSchema, err := defaultSchemaGenerator(t.Elem(), visited)
		if err != nil {
			return nil, fmt.Errorf("generating schema for array items: %w", err)
		}
		schema.Items = itemSchema
	case reflect.Map:
		schema.Type = "object"
		additionalProps, err := defaultSchemaGenerator(t.Elem(), visited)
		if err != nil {
			return nil, fmt.Errorf("generating schema for map values: %w", err)
		}
		schema.AdditionalProperties = additionalProps
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = "integer"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Interface:
		schema.Type = "object" // Fallback for interfaces
	default:
		return nil, fmt.Errorf("unsupported type %s", t.Kind())
	}

	if typeName != "" && schema.Type == "object" {
		schema.Ref = "#/components/schemas/" + typeName
	}
	return schema, nil
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
		if p.In == "path" {
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

// isValidParameterIn checks if a parameter 'in' value is valid.
func isValidParameterIn(in string) bool {
	return in == "query" || in == "path" || in == "header" || in == "cookie"
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
