package openapi

import "reflect"

func CloneInfoObject(info *InfoObject) *InfoObject {
	if info == nil {
		return nil
	}

	clone := *info
	clone.Contact = cloneContactObject(info.Contact)
	clone.License = cloneLicenseObject(info.License)
	clone.Extensions = cloneStringAnyMap(info.Extensions)

	return &clone
}

func CloneOperation(op *Operation) *Operation {
	if op == nil {
		return nil
	}

	state := cloneOpenAPIState{
		operations: make(map[*Operation]*Operation),
		pathItems:  make(map[*PathItem]*PathItem),
	}

	return state.cloneOperation(op)
}

type cloneOpenAPIState struct {
	operations map[*Operation]*Operation
	pathItems  map[*PathItem]*PathItem
}

func (state *cloneOpenAPIState) cloneOperation(op *Operation) *Operation {
	if op == nil {
		return nil
	}
	if cloned, ok := state.operations[op]; ok {
		return cloned
	}

	clone := *op
	state.operations[op] = &clone
	clone.Tags = cloneStringSlice(op.Tags)
	clone.ExternalDocs = cloneExternalDocumentation(op.ExternalDocs)
	if op.Parameters != nil {
		clone.Parameters = make([]*ParameterObject, len(op.Parameters))
		for index, param := range op.Parameters {
			clone.Parameters[index] = CloneParameterObject(param)
		}
	}
	clone.RequestBody = CloneRequestBodyObject(op.RequestBody)
	if op.Responses != nil {
		clone.Responses = make(map[string]*ResponseObject, len(op.Responses))
		for key, resp := range op.Responses {
			clone.Responses[key] = CloneResponseObject(resp)
		}
	}
	if op.Callbacks != nil {
		clone.Callbacks = make(map[string]*PathItem, len(op.Callbacks))
		for key, item := range op.Callbacks {
			clone.Callbacks[key] = state.clonePathItem(item)
		}
	}
	if op.Security != nil {
		clone.Security = make([]*SecurityRequirement, len(op.Security))
		for index, sec := range op.Security {
			clone.Security[index] = CloneSecurityRequirement(sec)
		}
	}
	if op.Servers != nil {
		clone.Servers = make([]*ServerObject, len(op.Servers))
		for index, server := range op.Servers {
			clone.Servers[index] = cloneServerObject(server)
		}
	}
	clone.Extensions = cloneStringAnyMap(op.Extensions)

	return &clone
}

func (state *cloneOpenAPIState) clonePathItem(item *PathItem) *PathItem {
	if item == nil {
		return nil
	}
	if cloned, ok := state.pathItems[item]; ok {
		return cloned
	}

	clone := *item
	state.pathItems[item] = &clone
	clone.Get = state.cloneOperation(item.Get)
	clone.Put = state.cloneOperation(item.Put)
	clone.Post = state.cloneOperation(item.Post)
	clone.Delete = state.cloneOperation(item.Delete)
	clone.Options = state.cloneOperation(item.Options)
	clone.Head = state.cloneOperation(item.Head)
	clone.Patch = state.cloneOperation(item.Patch)
	clone.Trace = state.cloneOperation(item.Trace)
	if item.Parameters != nil {
		clone.Parameters = make([]*ParameterObject, len(item.Parameters))
		for index, param := range item.Parameters {
			clone.Parameters[index] = CloneParameterObject(param)
		}
	}
	if item.Servers != nil {
		clone.Servers = make([]*ServerObject, len(item.Servers))
		for index, server := range item.Servers {
			clone.Servers[index] = cloneServerObject(server)
		}
	}
	clone.Extensions = cloneStringAnyMap(item.Extensions)

	return &clone
}

func cloneExternalDocumentation(doc *ExternalDocumentation) *ExternalDocumentation {
	if doc == nil {
		return nil
	}

	clone := *doc
	clone.Extensions = cloneStringAnyMap(doc.Extensions)

	return &clone
}

func cloneContactObject(contact *ContactObject) *ContactObject {
	if contact == nil {
		return nil
	}

	clone := *contact
	clone.Extensions = cloneStringAnyMap(contact.Extensions)

	return &clone
}

func cloneLicenseObject(license *LicenseObject) *LicenseObject {
	if license == nil {
		return nil
	}

	clone := *license
	clone.Extensions = cloneStringAnyMap(license.Extensions)

	return &clone
}

func cloneStringSlice(values []string) []string {
	if values == nil {
		return nil
	}

	clone := make([]string, len(values))
	copy(clone, values)
	return clone
}

func CloneParameterObject(param *ParameterObject) *ParameterObject {
	if param == nil {
		return nil
	}

	clone := *param
	if param.Explode != nil {
		explode := *param.Explode
		clone.Explode = &explode
	}
	clone.Schema = CloneSchema(param.Schema)
	clone.Example = cloneValue(param.Example)
	if param.Examples != nil {
		clone.Examples = make(map[string]*ExampleObject, len(param.Examples))
		for key, value := range param.Examples {
			clone.Examples[key] = CloneExampleObject(value)
		}
	}
	if param.Content != nil {
		clone.Content = make(map[string]*MediaType, len(param.Content))
		for key, value := range param.Content {
			clone.Content[key] = CloneMediaType(value)
		}
	}
	if param.Extensions != nil {
		clone.Extensions = make(map[string]*any, len(param.Extensions))
		for key, value := range param.Extensions {
			clone.Extensions[key] = cloneAnyPointer(value)
		}
	}

	return &clone
}

func CloneRequestBodyObject(body *RequestBodyObject) *RequestBodyObject {
	if body == nil {
		return nil
	}

	clone := *body
	if body.Content != nil {
		clone.Content = make(map[string]*MediaType, len(body.Content))
		for key, value := range body.Content {
			clone.Content[key] = CloneMediaType(value)
		}
	}
	clone.Extensions = cloneStringAnyMap(body.Extensions)

	return &clone
}

func CloneResponseObject(resp *ResponseObject) *ResponseObject {
	if resp == nil {
		return nil
	}

	clone := *resp
	if resp.Headers != nil {
		clone.Headers = make(map[string]*HeaderObject, len(resp.Headers))
		for key, value := range resp.Headers {
			clone.Headers[key] = cloneHeaderObject(value)
		}
	}
	if resp.Content != nil {
		clone.Content = make(map[string]*MediaType, len(resp.Content))
		for key, value := range resp.Content {
			clone.Content[key] = CloneMediaType(value)
		}
	}
	if resp.Links != nil {
		clone.Links = make(map[string]*LinkObject, len(resp.Links))
		for key, value := range resp.Links {
			clone.Links[key] = cloneLinkObject(value)
		}
	}
	clone.Extensions = cloneStringAnyMap(resp.Extensions)

	return &clone
}

func CloneMediaType(media *MediaType) *MediaType {
	if media == nil {
		return nil
	}

	clone := *media
	clone.Schema = CloneSchema(media.Schema)
	clone.Example = cloneValue(media.Example)
	if media.Examples != nil {
		clone.Examples = make(map[string]*ExampleObject, len(media.Examples))
		for key, value := range media.Examples {
			clone.Examples[key] = CloneExampleObject(value)
		}
	}
	if media.Encoding != nil {
		clone.Encoding = make(map[string]*EncodingObject, len(media.Encoding))
		for key, value := range media.Encoding {
			clone.Encoding[key] = cloneEncodingObject(value)
		}
	}
	clone.Extensions = cloneStringAnyMap(media.Extensions)

	return &clone
}

func CloneSchema(schema *Schema) *Schema {
	if schema == nil {
		return nil
	}

	clone := *schema
	if schema.Properties != nil {
		clone.Properties = make(map[string]*Schema, len(schema.Properties))
		for key, value := range schema.Properties {
			clone.Properties[key] = CloneSchema(value)
		}
	}
	clone.Items = CloneSchema(schema.Items)
	if schema.Required != nil {
		clone.Required = make([]string, len(schema.Required))
		copy(clone.Required, schema.Required)
	}
	if schema.Enum != nil {
		clone.Enum = make([]any, len(schema.Enum))
		for index, value := range schema.Enum {
			clone.Enum[index] = cloneValue(value)
		}
	}
	clone.Default = cloneValue(schema.Default)
	clone.Example = cloneValue(schema.Example)
	if schema.Minimum != nil {
		minimum := *schema.Minimum
		clone.Minimum = &minimum
	}
	if schema.Maximum != nil {
		maximum := *schema.Maximum
		clone.Maximum = &maximum
	}
	clone.AdditionalProperties = CloneSchema(schema.AdditionalProperties)
	if schema.OneOf != nil {
		clone.OneOf = make([]*Schema, len(schema.OneOf))
		for index, value := range schema.OneOf {
			clone.OneOf[index] = CloneSchema(value)
		}
	}
	if schema.AnyOf != nil {
		clone.AnyOf = make([]*Schema, len(schema.AnyOf))
		for index, value := range schema.AnyOf {
			clone.AnyOf[index] = CloneSchema(value)
		}
	}
	if schema.AllOf != nil {
		clone.AllOf = make([]*Schema, len(schema.AllOf))
		for index, value := range schema.AllOf {
			clone.AllOf[index] = CloneSchema(value)
		}
	}
	if schema.Discriminator != nil {
		discriminator := *schema.Discriminator
		if schema.Discriminator.Mapping != nil {
			discriminator.Mapping = make(map[string]string, len(schema.Discriminator.Mapping))
			for key, value := range schema.Discriminator.Mapping {
				discriminator.Mapping[key] = value
			}
		}
		clone.Discriminator = &discriminator
	}
	clone.Extensions = cloneStringAnyMap(schema.Extensions)

	return &clone
}

func CloneExampleObject(example *ExampleObject) *ExampleObject {
	if example == nil {
		return nil
	}

	clone := *example
	clone.Value = cloneValue(example.Value)
	clone.Extensions = cloneStringAnyMap(example.Extensions)

	return &clone
}

func CloneSecurityRequirement(sec *SecurityRequirement) *SecurityRequirement {
	if sec == nil {
		return nil
	}
	if *sec == nil {
		var clone SecurityRequirement
		return &clone
	}

	clone := make(SecurityRequirement, len(*sec))
	for key, value := range *sec {
		clone[key] = cloneValue(value)
	}

	return &clone
}

func cloneHeaderObject(header *HeaderObject) *HeaderObject {
	if header == nil {
		return nil
	}

	clone := *header
	clone.Schema = CloneSchema(header.Schema)
	clone.Example = cloneValue(header.Example)
	clone.Examples = cloneStringAnyMap(header.Examples)
	if header.Content != nil {
		clone.Content = make(map[string]*MediaType, len(header.Content))
		for key, value := range header.Content {
			clone.Content[key] = CloneMediaType(value)
		}
	}
	clone.Extensions = cloneStringAnyMap(header.Extensions)

	return &clone
}

func cloneEncodingObject(encoding *EncodingObject) *EncodingObject {
	if encoding == nil {
		return nil
	}

	clone := *encoding
	if encoding.Headers != nil {
		clone.Headers = make(map[string]*HeaderObject, len(encoding.Headers))
		for key, value := range encoding.Headers {
			clone.Headers[key] = cloneHeaderObject(value)
		}
	}
	clone.Extensions = cloneStringAnyMap(encoding.Extensions)

	return &clone
}

func cloneLinkObject(link *LinkObject) *LinkObject {
	if link == nil {
		return nil
	}

	clone := *link
	clone.Parameters = cloneStringAnyMap(link.Parameters)
	clone.RequestBody = cloneValue(link.RequestBody)
	clone.Server = cloneServerObject(link.Server)
	clone.Extensions = cloneStringAnyMap(link.Extensions)

	return &clone
}

func cloneServerObject(server *ServerObject) *ServerObject {
	if server == nil {
		return nil
	}

	clone := *server
	if server.Variables != nil {
		clone.Variables = make(map[string]*ServerVariable, len(server.Variables))
		for key, value := range server.Variables {
			if value == nil {
				continue
			}
			variable := *value
			if value.Enum != nil {
				variable.Enum = cloneStringSlice(value.Enum)
			}
			variable.Extensions = cloneStringAnyMap(value.Extensions)
			clone.Variables[key] = &variable
		}
	}
	clone.Extensions = cloneStringAnyMap(server.Extensions)

	return &clone
}

func cloneAnyPointer(value *any) *any {
	if value == nil {
		return nil
	}

	cloned := cloneValue(*value)
	return &cloned
}

func cloneStringAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}

	clone := make(map[string]any, len(input))
	for key, value := range input {
		clone[key] = cloneValue(value)
	}

	return clone
}

type cloneState struct {
	visited map[cloneVisit]reflect.Value
}

type cloneVisit struct {
	typ reflect.Type
	ptr uintptr
}

func cloneValue(value any) any {
	if value == nil {
		return nil
	}

	state := cloneState{visited: make(map[cloneVisit]reflect.Value)}
	cloned := state.cloneReflectValue(reflect.ValueOf(value))
	if !cloned.IsValid() {
		return nil
	}

	return cloned.Interface()
}

func (state *cloneState) cloneReflectValue(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}

	if cloned, ok := state.lookupVisited(value); ok {
		return cloned
	}

	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := state.cloneReflectValue(value.Elem())
		out := reflect.New(value.Type()).Elem()
		out.Set(cloned)
		return out
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := reflect.New(value.Type().Elem())
		state.rememberVisit(value, cloned)
		cloned.Elem().Set(state.cloneReflectValue(value.Elem()))
		return cloned
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := reflect.MakeMapWithSize(value.Type(), value.Len())
		state.rememberVisit(value, cloned)
		iter := value.MapRange()
		for iter.Next() {
			cloned.SetMapIndex(iter.Key(), state.cloneReflectValue(iter.Value()))
		}
		return cloned
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		cloned := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		state.rememberVisit(value, cloned)
		for index := 0; index < value.Len(); index++ {
			cloned.Index(index).Set(state.cloneReflectValue(value.Index(index)))
		}
		return cloned
	case reflect.Array:
		cloned := reflect.New(value.Type()).Elem()
		for index := 0; index < value.Len(); index++ {
			cloned.Index(index).Set(state.cloneReflectValue(value.Index(index)))
		}
		return cloned
	case reflect.Struct:
		cloned := reflect.New(value.Type()).Elem()
		cloned.Set(value)
		for index := 0; index < value.NumField(); index++ {
			field := cloned.Field(index)
			if !field.CanSet() {
				continue
			}
			field.Set(state.cloneReflectValue(value.Field(index)))
		}
		return cloned
	default:
		return value
	}
}

func (state *cloneState) lookupVisited(value reflect.Value) (reflect.Value, bool) {
	key, ok := cloneVisitKey(value)
	if !ok {
		return reflect.Value{}, false
	}

	cloned, found := state.visited[key]
	return cloned, found
}

func (state *cloneState) rememberVisit(value reflect.Value, clone reflect.Value) {
	key, ok := cloneVisitKey(value)
	if !ok {
		return
	}

	state.visited[key] = clone
}

func cloneVisitKey(value reflect.Value) (cloneVisit, bool) {
	switch value.Kind() {
	case reflect.Pointer, reflect.Map:
		if value.IsNil() {
			return cloneVisit{}, false
		}
		return cloneVisit{typ: value.Type(), ptr: value.Pointer()}, true
	case reflect.Slice:
		if value.IsNil() {
			return cloneVisit{}, false
		}
		ptr := value.Pointer()
		if ptr == 0 {
			return cloneVisit{}, false
		}
		return cloneVisit{typ: value.Type(), ptr: ptr}, true
	default:
		return cloneVisit{}, false
	}
}
