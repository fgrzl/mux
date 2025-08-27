package mux

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// OpenAPISpec represents the root of an OpenAPI 3.1 document.
// See: https://spec.openapis.org/oas/v3.1.0#openapi-object
type OpenAPISpec struct {
	OpenAPI           string                 `json:"openapi" yaml:"openapi"`
	Info              *InfoObject            `json:"info" yaml:"info"`
	JsonSchemaDialect string                 `json:"jsonSchemaDialect,omitempty" yaml:"jsonSchemaDialect,omitempty"`
	Servers           []*ServerObject        `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths             map[string]*PathItem   `json:"paths" yaml:"paths"`
	Webhooks          map[string]*PathItem   `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
	Components        *ComponentsObject      `json:"components,omitempty" yaml:"components,omitempty"`
	Security          []*SecurityRequirement `json:"security,omitempty" yaml:"security,omitempty"`
	Tags              []*TagObject           `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs      *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Extensions        map[string]any         `json:"-,inline" yaml:"-,inline"`
}

// NewOpenAPISpec creates a new OpenAPISpec with default values.
func NewOpenAPISpec() *OpenAPISpec {
	return &OpenAPISpec{
		OpenAPI:           "3.1.0",
		JsonSchemaDialect: "https://json-schema.org/draft/2020-12/schema",
		Info:              &InfoObject{Version: "1.0.0"},
		Paths:             make(map[string]*PathItem),
		Servers:           []*ServerObject{{URL: "/"}},
	}
}

func (spec *OpenAPISpec) Normalize() *OpenAPISpec {
	if spec.Components != nil {
		if len(spec.Components.Responses) == 0 {
			spec.Components.Responses = nil
		}
		if len(spec.Components.Parameters) == 0 {
			spec.Components.Parameters = nil
		}
		if len(spec.Components.Examples) == 0 {
			spec.Components.Examples = nil
		}
		if len(spec.Components.RequestBodies) == 0 {
			spec.Components.RequestBodies = nil
		}
		if len(spec.Components.Headers) == 0 {
			spec.Components.Headers = nil
		}
		if len(spec.Components.SecuritySchemes) == 0 {
			spec.Components.SecuritySchemes = nil
		}
		if len(spec.Components.Links) == 0 {
			spec.Components.Links = nil
		}
	}
	return spec
}

// Validate performs basic validation on the OpenAPISpec.
func (spec *OpenAPISpec) Validate() error {
	if spec.OpenAPI != "3.1.0" {
		return fmt.Errorf("openapi must be '3.1.0', got %q", spec.OpenAPI)
	}
	if spec.Info.Title == "" {
		return fmt.Errorf("info.title is required")
	}
	if spec.Info.Version == "" {
		return fmt.Errorf("info.version is required")
	}
	if len(spec.Paths) == 0 {
		return fmt.Errorf("at least one path is required")
	}
	return nil
}

func (spec *OpenAPISpec) MarshalToFile(path string) error {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".json":
		data, err := json.MarshalIndent(spec, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling to JSON: %w", err)
		}
		return os.WriteFile(path, data, 0644)
	case ".yml", ".yaml":
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("creating file %q: %w", path, err)
		}
		defer file.Close()

		enc := yaml.NewEncoder(file)
		enc.SetIndent(2)
		defer enc.Close()
		return enc.Encode(spec)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func (spec *OpenAPISpec) UnmarshalFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %q: %w", path, err)
	}

	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".json":
		return json.Unmarshal(data, spec)
	case ".yaml", ".yaml":
		return yaml.Unmarshal(data, spec)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// InfoObject contains metadata about the API.
// See: https://spec.openapis.org/oas/v3.1.0#info-object
type InfoObject struct {
	Title          string         `json:"title" yaml:"title"`
	Summary        string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description    string         `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string         `json:"version" yaml:"version"`
	Extensions     map[string]any `json:"-,inline" yaml:"-,inline"`
}

// PathItem describes operations available on a single path.
// See: https://spec.openapis.org/oas/v3.1.0#path-item-object
type PathItem struct {
	Ref         string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string             `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string             `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation         `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation         `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation         `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation         `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation         `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation         `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation         `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation         `json:"trace,omitempty" yaml:"trace,omitempty"`
	Parameters  []*ParameterObject `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Servers     []*ServerObject    `json:"servers,omitempty" yaml:"servers,omitempty"`
	Extensions  map[string]any     `json:"-,inline" yaml:"-,inline"`
}

// Operation describes a single API operation.
// See: https://spec.openapis.org/oas/v3.1.0#operation-object
type Operation struct {
	Tags         []string                   `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string                     `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                     `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation     `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	OperationID  string                     `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []*ParameterObject         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *RequestBodyObject         `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    map[string]*ResponseObject `json:"responses" yaml:"responses"`
	Callbacks    map[string]*PathItem       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Deprecated   bool                       `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security     []*SecurityRequirement     `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      []*ServerObject            `json:"servers,omitempty" yaml:"servers,omitempty"`
	Extensions   map[string]any             `json:"-,inline" yaml:"-,inline"`
}

// ParameterObject defines a parameter for an operation or path.
// See: https://spec.openapis.org/oas/v3.1.0#parameter-object
type ParameterObject struct {
	Name            string                    `json:"name" yaml:"name"`
	In              string                    `json:"in" yaml:"in"`
	Description     string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                      `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                      `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                      `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string                    `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         *bool                     `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool                      `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema                   `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         any                       `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]*ExampleObject `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]*MediaType     `json:"content,omitempty" yaml:"content,omitempty"`
	Extensions      map[string]*any           `json:"-,inline" yaml:"-,inline"`
	// Converter is a runtime-only function that converts raw string values
	// (possibly multi-valued) into the typed value expected by the handler
	// and the binding logic. It is intentionally omitted from JSON/YAML
	// serialization because it's runtime-only.
	Converter func(values []string) (any, error) `json:"-" yaml:"-"`
}

// RequestBodyObject describes a request body.
// See: https://spec.openapis.org/oas/v3.1.0#request-body-object
type RequestBodyObject struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]*MediaType `json:"content" yaml:"content"`
	Required    bool                  `json:"required,omitempty" yaml:"required,omitempty"`
	Extensions  map[string]any        `json:"-,inline" yaml:"-,inline"`
}

// ResponseObject describes an API response.
// See: https://spec.openapis.org/oas/v3.1.0#response-object
type ResponseObject struct {
	Description string                   `json:"description" yaml:"description"`
	Headers     map[string]*HeaderObject `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]*MediaType    `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]*LinkObject   `json:"links,omitempty" yaml:"links,omitempty"`
	Extensions  map[string]any           `json:"-,inline" yaml:"-,inline"`
}

// MediaType defines a media type for request/response content.
// See: https://spec.openapis.org/oas/v3.1.0#media-type-object
type MediaType struct {
	Schema     *Schema                    `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example    any                        `json:"example,omitempty" yaml:"example,omitempty"`
	Examples   map[string]*ExampleObject  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding   map[string]*EncodingObject `json:"encoding,omitempty" yaml:"encoding,omitempty"`
	Extensions map[string]any             `json:"-,inline" yaml:"-,inline"`
}

// Schema represents a JSON Schema or a reference to one.
// See: https://spec.openapis.org/oas/v3.1.0#schema-object
type Schema struct {
	Ref                  string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type                 string             `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string             `json:"format,omitempty" yaml:"format,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty" yaml:"items,omitempty"`
	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []any              `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default              any                `json:"default,omitempty" yaml:"default,omitempty"`
	Example              any                `json:"example,omitempty" yaml:"example,omitempty"`
	Minimum              *float64           `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              *float64           `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Pattern              string             `json:"pattern,omitempty" yaml:"pattern,omitempty"` // Added for string regex constraints
	Extensions           map[string]any     `json:"-" yaml:"-,inline"`
}

// ComponentsObject holds reusable components.
// See: https://spec.openapis.org/oas/v3.1.0#components-object
type ComponentsObject struct {
	Schemas         map[string]*Schema            `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]*ResponseObject    `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]*ParameterObject   `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]*ExampleObject     `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]*RequestBodyObject `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]*HeaderObject      `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme    `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]*LinkObject        `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       map[string]any                `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	PathItems       map[string]any                `json:"pathItems,omitempty" yaml:"pathItems,omitempty"`
	Extensions      map[string]any                `json:"-,inline" yaml:"-,inline"`
}

// HeaderObject defines a response header.
// See: https://spec.openapis.org/oas/v3.1.0#header-object
type HeaderObject struct {
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                  `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated  bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Schema      *Schema               `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example     any                   `json:"example,omitempty" yaml:"example,omitempty"`
	Examples    map[string]any        `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content     map[string]*MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Extensions  map[string]any        `json:"-,inline" yaml:"-,inline"`
}

// ExampleObject defines example data.
// See: https://spec.openapis.org/oas/v3.1.0#example-object
type ExampleObject struct {
	Summary       string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string         `json:"description,omitempty" yaml:"description,omitempty"`
	Value         any            `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string         `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
	Extensions    map[string]any `json:"-,inline" yaml:"-,inline"`
}

// EncodingObject defines encoding for a media type.
// See: https://spec.openapis.org/oas/v3.1.0#encoding-object
type EncodingObject struct {
	ContentType   string                   `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]*HeaderObject `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string                   `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool                     `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool                     `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Extensions    map[string]any           `json:"-,inline" yaml:"-,inline"`
}

// LinkObject defines a link between operations.
// See: https://spec.openapis.org/oas/v3.1.0#link-object
type LinkObject struct {
	OperationRef string         `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationID  string         `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  any            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string         `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *ServerObject  `json:"server,omitempty" yaml:"server,omitempty"`
	Extensions   map[string]any `json:"-,inline" yaml:"-,inline"`
}

// SecurityScheme defines a security scheme.
// See: https://spec.openapis.org/oas/v3.1.0#security-scheme-object
type SecurityScheme struct {
	Type             string            `json:"type" yaml:"type"`
	Description      string            `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string            `json:"name,omitempty" yaml:"name,omitempty"`
	In               string            `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string            `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string            `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlowsObject `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIdConnectUrl string            `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
	Extensions       map[string]any    `json:"-,inline" yaml:"-,inline"`
}

// OAuthFlowsObject defines OAuth 2.0 flows.
// See: https://spec.openapis.org/oas/v3.1.0#oauth-flows-object
type OAuthFlowsObject struct {
	Implicit          *OAuthFlowObject `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlowObject `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlowObject `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlowObject `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
	Extensions        map[string]any   `json:"-,inline" yaml:"-,inline"`
}

// OAuthFlowObject defines an OAuth flow configuration.
// See: https://spec.openapis.org/oas/v3.1.0#oauth-flow-object
type OAuthFlowObject struct {
	AuthorizationUrl string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenUrl         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshUrl       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	Extensions       map[string]any    `json:"-,inline" yaml:"-,inline"`
}

// ServerObject defines an API server.
// See: https://spec.openapis.org/oas/v3.1.0#server-object
type ServerObject struct {
	URL         string                     `json:"url" yaml:"url"`
	Description string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]*ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
	Extensions  map[string]any             `json:"-,inline" yaml:"-,inline"`
}

// ServerVariable defines a server variable for URL substitution.
// See: https://spec.openapis.org/oas/v3.1.0#server-variable-object
type ServerVariable struct {
	Enum        []string       `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string         `json:"default" yaml:"default"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	Extensions  map[string]any `json:"-,inline" yaml:"-,inline"`
}

// TagObject defines a tag for grouping operations.
// See: https://spec.openapis.org/oas/v3.1.0#tag-object
type TagObject struct {
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Extensions   map[string]any         `json:"-,inline" yaml:"-,inline"`
}

// ExternalDocumentation provides external documentation links.
// See: https://spec.openapis.org/oas/v3.1.0#external-documentation-object
type ExternalDocumentation struct {
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string         `json:"url" yaml:"url"`
	Extensions  map[string]any `json:"-,inline" yaml:"-,inline"`
}

// ContactObject contains contact information.
// See: https://spec.openapis.org/oas/v3.1.0#contact-object
type ContactObject struct {
	Name       string         `json:"name,omitempty" yaml:"name,omitempty"`
	URL        string         `json:"url,omitempty" yaml:"url,omitempty"`
	Email      string         `json:"email,omitempty" yaml:"email"`
	Extensions map[string]any `json:"-,inline" yaml:"-,inline"`
}

// LicenseObject defines API license information.
// See: https://spec.openapis.org/oas/v3.1.0#license-object
type LicenseObject struct {
	Name       string         `json:"name" yaml:"name"`
	Identifier string         `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	URL        string         `json:"url,omitempty" yaml:"url,omitempty"`
	Extensions map[string]any `json:"-,inline" yaml:"-,inline"`
}

// SecurityRequirement specifies security requirements for an operation or API.
// See: https://spec.openapis.org/oas/v3.1.0#security-requirement-object
type SecurityRequirement map[string]any
