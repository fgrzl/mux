package mux

import (
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
