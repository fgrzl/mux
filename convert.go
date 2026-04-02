package mux

import (
	"encoding/json"

	internalopenapi "github.com/fgrzl/mux/internal/openapi"
)

func cloneSlice[T any](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	out := make([]T, len(in))
	copy(out, in)
	return out
}

func convertViaJSON[T any](in any) (T, error) {
	var out T
	if in == nil {
		return out, nil
	}
	data, err := json.Marshal(in)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, err
	}
	return out, nil
}

func toInternalSecurityRequirement(sec SecurityRequirement) *internalopenapi.SecurityRequirement {
	if len(sec) == 0 {
		return nil
	}
	converted := make(internalopenapi.SecurityRequirement, len(sec))
	for scheme, scopes := range sec {
		converted[scheme] = cloneSlice(scopes)
	}
	return &converted
}
