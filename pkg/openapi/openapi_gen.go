// Updated mux generator to ensure nested components are registered without global state
package openapi

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"reflect"
	"regexp"
	"strings"

	"github.com/fgrzl/json/jsonschema"
	"github.com/google/uuid"
)

// RouteData is a small value type used by the generator. Route collection
// is performed by the mux adapter so the openapi package does not depend on
// router internals.
type RouteData struct {
	Path   string
	Method string
	// Options should carry the OpenAPI Operation information; we store a pointer
	// to Operation so the generator can use it without depending on the router's
	// RouteOptions type.
	Options *Operation
}

// GeneratorOption is a configuration option for the OpenAPI Generator.
// Options are functional options that mutate the Generator at creation time.
type GeneratorOption func(*Generator)

// WithExamples configures the Generator to include example values in
// generated component schemas and request/response bodies.
// By default examples are omitted.
func WithExamples() GeneratorOption {
	return func(g *Generator) {
		g.withExamples = true
	}
}

// WithPathPrefix restricts generated paths to those that start with the
// provided prefix. The prefix may be provided with or without a leading
// slash (e.g. "api/v1" or "/api/v1"). If empty, no filtering is applied.
// WithPathPrefix adds a single path prefix to the Generator filter.
// Multiple calls to WithPathPrefix will accumulate prefixes. The prefix
// may be provided with or without a leading slash ("api/v1" or "/api/v1").
// When set, only routes whose paths start with any configured prefix will
// be included in the generated spec.
func WithPathPrefix(prefix string) GeneratorOption {
	return func(g *Generator) {
		if prefix == "" {
			return
		}
		p := prefix
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		if g.includePrefixes == nil {
			g.includePrefixes = []string{}
		}
		g.includePrefixes = append(g.includePrefixes, p)
	}
}

// Generator generates an OpenAPI 3.1 specification and holds internal state.
type Generator struct {
	spec            *OpenAPISpec
	builder         *jsonschema.Builder
	visited         map[reflect.Type]bool
	componentNames  map[string]string
	componentKeys   map[string]string
	componentHashes map[string]string
	withExamples    bool // default is false
	includePrefixes []string
}

// NewGenerator creates a Generator configured with the provided options.
//
// The returned Generator can be reused to generate multiple specs with
// different routers or prefix filters.
//
// Example:
//
//	gen := NewGenerator(WithExamples(), WithPathPrefix("api/v1"))
//	spec, err := mux.GenerateSpecWithGenerator(gen, router)
func NewGenerator(opts ...GeneratorOption) *Generator {
	gen := &Generator{
		spec:            NewOpenAPISpec(),
		builder:         jsonschema.NewBuilder(),
		visited:         make(map[reflect.Type]bool),
		componentNames:  make(map[string]string),
		componentKeys:   make(map[string]string),
		componentHashes: make(map[string]string),
	}

	for _, opt := range opts {
		opt(gen)
	}

	return gen
}

// GenerateSpec builds an OpenAPI specification for the provided Router.
//
// The Generator's options control behavior: examples inclusion and path
// prefix filtering. Routes that carry OpenAPI metadata must define a non-empty
// OperationID or generation returns an error.
//
// The returned *OpenAPISpec is validated before being returned; callers
// may marshal it to disk with MarshalToFile or inspect it programmatically.
// GenerateSpecFromRoutes builds an OpenAPI spec from a pre-collected list of
// routes and the provided InfoObject. This keeps the generator independent
// from router internals and avoids package import cycles.
func (g *Generator) GenerateSpecFromRoutes(info *InfoObject, routes []RouteData) (*OpenAPISpec, error) {
	if info == nil {
		return nil, fmt.Errorf("info is nil")
	}

	g.spec = NewOpenAPISpec()
	g.visited = make(map[reflect.Type]bool)
	g.componentNames = make(map[string]string)
	g.componentKeys = make(map[string]string)
	g.componentHashes = make(map[string]string)
	g.spec.Info = info
	g.ensureComponentInit()
	// If includePrefixes is set, filter routes to those starting with any prefix.
	if len(g.includePrefixes) > 0 {
		prefixes := make([]string, 0, len(g.includePrefixes))
		for _, p := range g.includePrefixes {
			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			prefixes = append(prefixes, p)
		}
		filtered := make([]RouteData, 0, len(routes))
		for _, rd := range routes {
			for _, pref := range prefixes {
				if strings.HasPrefix(rd.Path, pref) {
					filtered = append(filtered, rd)
					break
				}
			}
		}
		routes = filtered
	}

	for _, rd := range routes {
		if rd.Options == nil {
			continue
		}
		if rd.Options.OperationID == "" {
			if operationRequiresOperationID(rd.Options) {
				return nil, fmt.Errorf("route %s %s missing OperationID", rd.Method, rd.Path)
			}
			continue
		}
		if err := g.appendRoute(rd); err != nil {
			return nil, err
		}
	}
	return g.spec, g.spec.Validate()
}

func operationRequiresOperationID(op *Operation) bool {
	if op == nil {
		return false
	}

	return len(op.Tags) > 0 ||
		op.Summary != "" ||
		op.Description != "" ||
		op.ExternalDocs != nil ||
		len(op.Parameters) > 0 ||
		op.RequestBody != nil ||
		len(op.Responses) > 0 ||
		len(op.Callbacks) > 0 ||
		op.Deprecated ||
		len(op.Security) > 0 ||
		len(op.Servers) > 0 ||
		len(op.Extensions) > 0
}

// GenerateAndSave generates an OpenAPI spec for the router and writes it
// to the given filesystem path. The spec is validated before being saved.
//
// Returns an error if generation or file writing fails.
// GenerateAndSave writes a spec to disk using the provided info and routes.
func (g *Generator) GenerateAndSave(info *InfoObject, routes []RouteData, path string) error {
	spec, err := g.GenerateSpecFromRoutes(info, routes)
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

func (g *Generator) appendRoute(rd RouteData) error {
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

	// Work on a detached copy so spec generation never mutates caller-owned
	// route metadata or the router's stored operation tree.
	newOp := cloneOperationForSpec(opt)
	if len(newOp.Responses) == 0 {
		code := getDefaultResponseCode(method)
		newOp.Responses = map[string]*ResponseObject{code: {Description: getDefaultResponseDescription(code)}}
	}

	if err := g.prepareOperationForSpec(newOp); err != nil {
		return err
	}

	switch method {
	case "get":
		item.Get = newOp
	case "post":
		item.Post = newOp
	case "put":
		item.Put = newOp
	case "delete":
		item.Delete = newOp
	case "patch":
		item.Patch = newOp
	}
	g.spec.Paths[path] = item
	return nil
}

func (g *Generator) prepareOperationForSpec(op *Operation) error {
	if op == nil {
		return nil
	}

	for _, param := range op.Parameters {
		if param == nil {
			continue
		}
		if err := g.ensureComponentSchema(param.Example, param.Schema); err != nil {
			return err
		}
		if !g.withExamples {
			param.Example = nil
			param.Examples = nil
		}
		param.Converter = nil
	}

	if op.RequestBody != nil {
		for _, media := range op.RequestBody.Content {
			if media == nil {
				continue
			}
			if err := g.ensureComponentSchema(media.Example, media.Schema); err != nil {
				return err
			}
			if !g.withExamples {
				media.Example = nil
				media.Examples = nil
			}
		}
	}

	for code, resp := range op.Responses {
		if resp == nil {
			continue
		}
		for _, media := range resp.Content {
			if media == nil {
				continue
			}
			if err := g.ensureComponentSchema(media.Example, media.Schema); err != nil {
				return err
			}
			if !g.withExamples {
				media.Example = nil
				media.Examples = nil
			}
		}
		if resp.Description == "" {
			resp.Description = getDefaultResponseDescription(code)
		}
	}

	return nil
}

func cloneOperationForSpec(op *Operation) *Operation {
	if op == nil {
		return nil
	}

	clone := *op
	if len(op.Parameters) > 0 {
		clone.Parameters = make([]*ParameterObject, 0, len(op.Parameters))
		for _, param := range op.Parameters {
			clone.Parameters = append(clone.Parameters, cloneParameterForSpec(param))
		}
	}
	if op.RequestBody != nil {
		clone.RequestBody = cloneRequestBodyForSpec(op.RequestBody)
	}
	if len(op.Responses) > 0 {
		clone.Responses = make(map[string]*ResponseObject, len(op.Responses))
		for code, resp := range op.Responses {
			clone.Responses[code] = cloneResponseForSpec(resp)
		}
	}

	return &clone
}

func cloneParameterForSpec(param *ParameterObject) *ParameterObject {
	return CloneParameterObject(param)
}

func cloneRequestBodyForSpec(body *RequestBodyObject) *RequestBodyObject {
	return CloneRequestBodyObject(body)
}

func cloneResponseForSpec(resp *ResponseObject) *ResponseObject {
	return CloneResponseObject(resp)
}

func cloneMediaTypeForSpec(media *MediaType) *MediaType {
	return CloneMediaType(media)
}

func cloneSchemaForSpec(schema *Schema) *Schema {
	return CloneSchema(schema)
}

func (g *Generator) ensureComponentSchema(example any, schema *Schema) error {
	if schema == nil {
		return nil
	}

	// Handle composite schemas (anyOf, oneOf, allOf) - process each sub-schema
	// For composite schemas, we need to recursively ensure their sub-schemas are registered
	// Note: Infinite recursion is prevented by checking if schemas are already registered
	// before generating them (see checks below and in GenerateSchemaForType's visited map).
	if len(schema.AnyOf) > 0 {
		for _, subSchema := range schema.AnyOf {
			if err := g.ensureSchemaRef(subSchema); err != nil {
				return err
			}
		}
	}
	if len(schema.OneOf) > 0 {
		for _, subSchema := range schema.OneOf {
			if err := g.ensureSchemaRef(subSchema); err != nil {
				return err
			}
		}
	}
	if len(schema.AllOf) > 0 {
		for _, subSchema := range schema.AllOf {
			if err := g.ensureSchemaRef(subSchema); err != nil {
				return err
			}
		}
	}

	// Handle array items - ensure item schemas (and their refs) are registered
	if schema.Type == "array" && schema.Items != nil {
		// Get the element type from the example
		var itemExample any
		if example != nil {
			t := reflect.TypeOf(example)
			if t != nil && (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) {
				// Create zero value of the element type
				elemType := t.Elem()
				if elemType != nil {
					itemExample = reflect.Zero(elemType).Interface()
				}
			}
		}
		if err := g.ensureComponentSchema(itemExample, schema.Items); err != nil {
			return err
		}
	}

	// Handle single schema ref
	if schema.Ref == "" {
		return nil
	}

	rawName, _ := strings.CutPrefix(schema.Ref, "#/components/schemas/")
	typeName := sanitizeComponentName(rawName)
	// Ensure the ref used in the operation points to the sanitized name
	schema.Ref = "#/components/schemas/" + typeName
	// Protection: Skip if already registered (prevents redundant work and circular issues)
	if _, exists := g.spec.Components.Schemas[typeName]; exists {
		return nil
	}
	t := reflect.TypeOf(example)
	if t == nil {
		return fmt.Errorf("missing type info for schema ref %q", schema.Ref)
	}
	// If the example is a pointer, slice/array, or map, unwrap to the
	// underlying element type so named element types (e.g. []MyModel) can be
	// registered as components. This ensures nested array/map element types
	// become top-level component schemas instead of causing "unnamed type"
	// errors.
	for {
		switch t.Kind() {
		case reflect.Ptr:
			t = t.Elem()
		case reflect.Slice, reflect.Array, reflect.Map:
			// For maps, register the value type
			t = t.Elem()
		default:
			goto doneUnwrap
		}
		if t == nil {
			break
		}
	}
doneUnwrap:
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

// ensureSchemaRef ensures that a schema ref is registered as a component.
// Unlike ensureComponentSchema, this works with schemas that may have examples attached.
func (g *Generator) ensureSchemaRef(schema *Schema) error {
	if schema == nil || schema.Ref == "" {
		return nil
	}

	rawName, _ := strings.CutPrefix(schema.Ref, "#/components/schemas/")
	typeName := sanitizeComponentName(rawName)
	schema.Ref = "#/components/schemas/" + typeName

	// Protection: Skip if already registered (prevents redundant work and circular issues)
	if _, exists := g.spec.Components.Schemas[typeName]; exists {
		// Clear the example from the ref schema since it will be in the component
		if !g.withExamples {
			schema.Example = nil
		}
		return nil
	}

	// Try to get type info from the attached example
	example := schema.Example
	if example == nil {
		// No example attached - we can't generate the schema without type information
		return fmt.Errorf("cannot resolve component schema %q without type information", typeName)
	}

	t := reflect.TypeOf(example)
	if t == nil {
		return fmt.Errorf("missing type info for schema ref %q", schema.Ref)
	}

	// Unwrap pointers
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		if t == nil {
			break
		}
	}

	// Generate the component schema
	s, err := g.GenerateSchemaForType(t)
	if err != nil {
		return err
	}

	g.spec.Components.Schemas[typeName] = s

	// Clear the example from the ref schema after processing (examples go on the component or mediatype, not the ref)
	if !g.withExamples {
		schema.Example = nil
	}

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
	scope := typeIdentity(t)
	typeName := g.resolveComponentName(t.Name(), scope, scope)
	// Protection: Check if we're already processing this type (prevents infinite recursion)
	// This is critical for types with circular references (e.g., Node.Parent -> *Node)
	if g.visited[t] {
		return &Schema{Ref: "#/components/schemas/" + typeName}, nil
	}
	g.visited[t] = true
	defer delete(g.visited, t)

	root, components := g.builder.SchemaWithComponents(t)
	componentJSON := make(map[string][]byte, len(components))
	componentHashes := make(map[string]string, len(components))
	for name, raw := range components {
		data, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("marshal component schema %q: %w", name, err)
		}
		componentJSON[name] = data
		componentHashes[name] = string(data)
	}

	// Build a name mapping for all components to sanitized names
	nameMap := make(map[string]string, len(components))
	for name := range components {
		nameMap[name] = g.resolveComponentName(name, scope, componentHashes[name])
	}

	// Insert/Update components with sanitized names and rewritten refs
	for oldName := range components {
		newName := nameMap[oldName]
		if _, exists := g.spec.Components.Schemas[newName]; exists {
			continue
		}
		s := &Schema{}
		if err := json.Unmarshal(componentJSON[oldName], &s); err != nil {
			return nil, fmt.Errorf("unmarshal component schema %q: %w", oldName, err)
		}
		rewriteSchemaRefs(s, nameMap)
		g.spec.Components.Schemas[newName] = s
	}

	data, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal root schema: %w", err)
	}
	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal root schema: %w", err)
	}

	// Rewrite refs in the root schema too
	rewriteSchemaRefs(&schema, nameMap)

	return &schema, nil
}

// The function will register any nested component schemas into the
// Generator's components map and return a Schema that may be a $ref to a
// component or an inline schema. The provided type must be a named type
// (non-anonymous) or an error will be returned.
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
	for _, match := range pathParamRegex.FindAllStringSubmatch(path, -1) {
		pathParams[match[1]] = true
	}
	for _, p := range params {
		if p == nil {
			continue
		}
		if p.In == "path" {
			if !pathParams[p.Name] {
				return fmt.Errorf("path parameter %q not found in path %q", p.Name, path)
			}
			delete(pathParams, p.Name)
			continue
		}
		if p.In == "" && pathParams[p.Name] {
			return fmt.Errorf("path parameter %q in path %q must declare in=\"path\"", p.Name, path)
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
	method = strings.ToLower(method)
	switch method {
	case "post":
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

var pathParamRegex = regexp.MustCompile(`\{([^}]+)\}`)

// sanitizeComponentName converts potentially invalid component names (e.g., generics with package paths)
// into OpenAPI-safe identifiers. Examples:
//
//	"Page[*github.com/acme/project/pkg.Model]" => "PageModel"
//	"Result[github.com/x/y.User, *github.com/x/y.Error]" => "ResultUserError"
func sanitizeComponentName(name string) string {
	if name == "" {
		return name
	}
	base := name

	var args []string
	if i := strings.Index(name, "["); i >= 0 {
		base = name[:i]
		inner := name[i+1:]
		if j := strings.LastIndex(inner, "]"); j >= 0 {
			inner = inner[:j]
		}
		// Split by comma for multiple type args
		parts := strings.Split(inner, ",")
		if len(parts) > 0 {
			args = make([]string, 0, len(parts))
		}
		for _, p := range parts {
			p = strings.TrimSpace(p)
			// Remove pointer and slice/map tokens
			for strings.HasPrefix(p, "*") || strings.HasPrefix(p, "[]") || strings.HasPrefix(p, "map[") {
				if after, ok := strings.CutPrefix(p, "*"); ok {
					p = after
				} else if after, ok := strings.CutPrefix(p, "[]"); ok {
					p = after
				} else if after, ok := strings.CutPrefix(p, "map["); ok {
					// best-effort: drop map[...] prefix
					if k := strings.Index(after, "]"); k >= 0 {
						p = after[k+1:]
					} else {
						p = after
					}
				}
				p = strings.TrimSpace(p)
			}
			// Strip package path and qualifiers
			if idx := strings.LastIndex(p, "/"); idx >= 0 {
				p = p[idx+1:]
			}
			if idx := strings.LastIndex(p, "."); idx >= 0 {
				p = p[idx+1:]
			}
			// Remove any remaining non-alphanumeric characters
			cleaned := make([]rune, 0, len(p))
			for _, r := range p {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
					cleaned = append(cleaned, r)
				}
			}
			if len(cleaned) > 0 {
				args = append(args, string(cleaned))
			}
		}
	}
	// Clean base similarly (in case it contains package qualifiers)
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if idx := strings.LastIndex(base, "."); idx >= 0 {
		base = base[idx+1:]
	}
	baseClean := make([]rune, 0, len(base))
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			baseClean = append(baseClean, r)
		}
	}
	var b strings.Builder
	b.Grow(len(baseClean) + 16*len(args))
	b.WriteString(string(baseClean))
	for _, a := range args {
		b.WriteString(a)
	}
	result := b.String()
	if result == "" {
		return name // fallback
	}
	return result
}

func typeIdentity(t reflect.Type) string {
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if pkg := t.PkgPath(); pkg != "" {
		return pkg + "." + t.String()
	}
	return t.String()
}

func componentNameSuffix(key string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return fmt.Sprintf("%08x", h.Sum32())[:8]
}

func (g *Generator) resolveComponentName(rawName, scope, hash string) string {
	if rawName == "" {
		return rawName
	}
	if g.spec == nil {
		g.spec = NewOpenAPISpec()
	}
	g.ensureComponentInit()
	if g.componentNames == nil {
		g.componentNames = make(map[string]string)
	}
	if g.componentKeys == nil {
		g.componentKeys = make(map[string]string)
	}
	if g.componentHashes == nil {
		g.componentHashes = make(map[string]string)
	}

	key := rawName
	if scope != "" {
		key = scope + "::" + rawName
	}
	if resolved, ok := g.componentNames[key]; ok {
		return resolved
	}

	base := sanitizeComponentName(rawName)
	suffix := componentNameSuffix(key)
	for attempt := 0; ; attempt++ {
		candidate := base
		switch attempt {
		case 0:
		case 1:
			candidate = base + "_" + suffix
		default:
			candidate = fmt.Sprintf("%s_%s_%d", base, suffix, attempt)
		}

		if existingHash, ok := g.componentHashes[candidate]; ok {
			if hash != "" && existingHash == hash {
				g.componentNames[key] = candidate
				return candidate
			}
			if existingKey, ok := g.componentKeys[candidate]; ok && existingKey != key {
				continue
			}
		}
		if _, exists := g.spec.Components.Schemas[candidate]; exists {
			if existingKey, ok := g.componentKeys[candidate]; !ok || existingKey != key {
				continue
			}
		}

		g.componentNames[key] = candidate
		g.componentKeys[candidate] = key
		if hash != "" {
			g.componentHashes[candidate] = hash
		}
		return candidate
	}
}

// rewriteSchemaRefs updates $ref values in a schema tree according to a name mapping.
func rewriteSchemaRefs(s *Schema, nameMap map[string]string) {
	if s == nil {
		return
	}
	if s.Ref != "" {
		if after, ok := strings.CutPrefix(s.Ref, "#/components/schemas/"); ok {
			old := after
			if newName, ok := nameMap[old]; ok && newName != old {
				s.Ref = "#/components/schemas/" + newName
			} else {
				// Also handle already-sanitized names for idempotency
				s.Ref = "#/components/schemas/" + sanitizeComponentName(old)
			}
		}
	}
	if s.Items != nil {
		rewriteSchemaRefs(s.Items, nameMap)
	}
	if s.AdditionalProperties != nil {
		rewriteSchemaRefs(s.AdditionalProperties, nameMap)
	}
	for _, prop := range s.Properties {
		rewriteSchemaRefs(prop, nameMap)
	}
}
