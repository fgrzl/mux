package mux

import (
	"encoding/json"
	"fmt"

	internalopenapi "github.com/fgrzl/mux/internal/openapi"
	"gopkg.in/yaml.v3"
)

type OpenAPISpec struct {
	inner *internalopenapi.OpenAPISpec
}

func wrapOpenAPISpec(inner *internalopenapi.OpenAPISpec) *OpenAPISpec {
	if inner == nil {
		return nil
	}
	return &OpenAPISpec{inner: inner}
}

func (spec *OpenAPISpec) Normalize() *OpenAPISpec {
	if spec == nil || spec.inner == nil {
		return spec
	}
	spec.inner.Normalize()
	return spec
}

func (spec *OpenAPISpec) Validate() error {
	if spec == nil || spec.inner == nil {
		return fmt.Errorf("spec is nil")
	}
	return spec.inner.Validate()
}

func (spec *OpenAPISpec) MarshalJSON() ([]byte, error) {
	if spec == nil || spec.inner == nil {
		return json.Marshal((*internalopenapi.OpenAPISpec)(nil))
	}
	return json.Marshal(spec.inner)
}

func (spec *OpenAPISpec) UnmarshalJSON(data []byte) error {
	if spec == nil {
		return fmt.Errorf("spec is nil")
	}
	if string(data) == "null" {
		spec.inner = nil
		return nil
	}

	var inner internalopenapi.OpenAPISpec
	if err := json.Unmarshal(data, &inner); err != nil {
		return err
	}
	spec.inner = &inner
	return nil
}

func (spec *OpenAPISpec) MarshalYAML() (any, error) {
	if spec == nil || spec.inner == nil {
		return nil, nil
	}
	return spec.inner, nil
}

func (spec *OpenAPISpec) UnmarshalYAML(value *yaml.Node) error {
	if spec == nil {
		return fmt.Errorf("spec is nil")
	}
	if value == nil {
		spec.inner = nil
		return nil
	}

	var inner internalopenapi.OpenAPISpec
	if err := value.Decode(&inner); err != nil {
		return err
	}
	spec.inner = &inner
	return nil
}

func (spec *OpenAPISpec) MarshalToFile(path string) error {
	if spec == nil || spec.inner == nil {
		return fmt.Errorf("spec is nil")
	}
	return spec.inner.MarshalToFile(path)
}

func (spec *OpenAPISpec) UnmarshalFromFile(path string) error {
	if spec == nil {
		return fmt.Errorf("spec is nil")
	}

	var inner internalopenapi.OpenAPISpec
	if err := inner.UnmarshalFromFile(path); err != nil {
		return err
	}
	spec.inner = &inner
	return nil
}

type SecurityRequirement map[string][]string
