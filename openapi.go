package mux

import (
	"fmt"
	"reflect"
	"strings"
)

type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi" yaml:"openapi"`
	Info       InfoObject          `json:"info" yaml:"info"`
	Paths      map[string]PathItem `json:"paths" yaml:"paths"`
	Components ComponentsObject    `json:"components" yaml:"components"`
}

type InfoObject struct {
	Title   string `json:"title" yaml:"title"`
	Version string `json:"version" yaml:"version"`
}

type PathItem map[string]Operation // keyed by method: get, post, etc.

type Operation struct {
	OperationID string                    `json:"operationId" yaml:"operationId"`
	Summary     string                    `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []ParameterObject         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBodyObject        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]ResponseObject `json:"responses" yaml:"responses"`
}

type ParameterObject struct {
	Name     string       `json:"name" yaml:"name"`
	In       string       `json:"in" yaml:"in"`
	Required bool         `json:"required" yaml:"required"`
	Schema   ComponentRef `json:"schema" yaml:"schema"`
}

type RequestBodyObject struct {
	Content map[string]MediaType `json:"content" yaml:"content"`
}

type ResponseObject struct {
	Description string               `json:"description" yaml:"description"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

type MediaType struct {
	Schema ComponentRef `json:"schema" yaml:"schema"`
}

type ComponentsObject struct {
	Schemas map[string]ComponentRef `json:"schemas" yaml:"schemas"`
}

type ComponentRef struct {
	Ref  string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

func GenerateSpec(router *Router) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: InfoObject{
			Title:   "API",
			Version: "1.0.0",
		},
		Paths:      make(map[string]PathItem),
		Components: ComponentsObject{Schemas: map[string]ComponentRef{}},
	}
	type routeData struct {
		Path    string
		Method  string
		Options *RouteOptions
	}

	var collect func(path string, node *routeNode, acc *[]routeData)
	collect = func(path string, node *routeNode, acc *[]routeData) {
		for method, opt := range node.RouteOptions {
			*acc = append(*acc, routeData{Path: path, Method: method, Options: opt})
		}
		for seg, child := range node.Children {
			collect(path+"/"+seg, child, acc)
		}
		if node.ParamChild != nil {
			collect(path+"/{"+node.ParamChild.ParamName+"}", node.ParamChild, acc)
		}
		if node.Wildcard != nil {
			collect(path+"/*", node.Wildcard, acc)
		}
		if node.CatchAll != nil {
			collect(path+"/**", node.CatchAll, acc)
		}
	}

	var routes []routeData
	collect("", router.registry.root, &routes)

	for _, rd := range routes {
		path := rd.Path
		method := strings.ToLower(rd.Method)
		opt := rd.Options

		if _, ok := spec.Paths[path]; !ok {
			spec.Paths[path] = PathItem{}
		}

		item := spec.Paths[path]
		op := Operation{
			OperationID: opt.OperationID,
			Summary:     opt.Summary,
			Description: opt.Description,
			Responses:   map[string]ResponseObject{},
		}

		for _, p := range opt.Parameters {
			typeName := getOrRegisterSchema(spec, reflect.TypeOf(p.Model))
			op.Parameters = append(op.Parameters, ParameterObject{
				Name:     p.Name,
				In:       p.Source,
				Required: true,
				Schema:   ComponentRef{Ref: "#/components/schemas/" + typeName},
			})
		}

		if opt.RequestBody != nil {
			typeName := getOrRegisterSchema(spec, opt.RequestBody.Type)
			op.RequestBody = &RequestBodyObject{
				Content: map[string]MediaType{
					opt.RequestBody.ContentType: {
						Schema: ComponentRef{Ref: "#/components/schemas/" + typeName},
					},
				},
			}
		}

		for code, schema := range opt.Responses {
			resp := ResponseObject{
				Description: "Success",
			}
			if schema != nil {
				typeName := getOrRegisterSchema(spec, schema.Type)
				resp.Content = map[string]MediaType{
					schema.ContentType: {
						Schema: ComponentRef{Ref: "#/components/schemas/" + typeName},
					},
				}
			} else {
				resp.Description = "No content"
			}
			op.Responses[fmt.Sprintf("%d", code)] = resp
		}

		item[method] = op
		spec.Paths[path] = item
	}

	return spec
}

func getOrRegisterSchema(spec *OpenAPISpec, t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeName := t.Name()
	if _, ok := spec.Components.Schemas[typeName]; !ok {
		spec.Components.Schemas[typeName] = ComponentRef{Type: typeToOpenAPI(t)}
	}
	return typeName
}

func typeToOpenAPI(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map, reflect.Struct:
		return "object"
	default:
		return "string"
	}
}
