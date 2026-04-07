package mux

import internalrouter "github.com/fgrzl/mux/internal/router"

// RouterOption configures router behavior or top-level OpenAPI info metadata.
type RouterOption struct {
	apply internalrouter.RouterOption
}

// WithTitle sets the API-level OpenAPI title. This is not the same as a
// route-level summary.
func WithTitle(title string) RouterOption {
	return RouterOption{apply: internalrouter.WithTitle(title)}
}

// WithSummary sets the API-level OpenAPI summary shown on the Info object.
// Use RouteBuilder.WithSummary for per-route summaries.
func WithSummary(summary string) RouterOption {
	return RouterOption{apply: internalrouter.WithSummary(summary)}
}

// WithDescription sets the API-level OpenAPI description shown on the Info
// object. Use RouteBuilder.WithDescription for per-route descriptions.
func WithDescription(description string) RouterOption {
	return RouterOption{apply: internalrouter.WithDescription(description)}
}

// WithTermsOfService sets the API-level OpenAPI termsOfService URL.
func WithTermsOfService(url string) RouterOption {
	return RouterOption{apply: internalrouter.WithTermsOfService(url)}
}

// WithVersion sets the API-level OpenAPI version string.
func WithVersion(version string) RouterOption {
	return RouterOption{apply: internalrouter.WithVersion(version)}
}

// WithContact sets the API-level OpenAPI contact object.
func WithContact(name, url, email string) RouterOption {
	return RouterOption{apply: internalrouter.WithContact(name, url, email)}
}

// WithLicense sets the API-level OpenAPI license object.
func WithLicense(name, url string) RouterOption {
	return RouterOption{apply: internalrouter.WithLicense(name, url)}
}

// WithClientURL sets the client-facing base URL used by router integrations
// that need to emit server metadata.
func WithClientURL(url string) RouterOption {
	return RouterOption{apply: internalrouter.WithClientURL(url)}
}

// WithContextPooling enables RouteContext pooling to reduce allocations on hot
// paths.
func WithContextPooling() RouterOption {
	return RouterOption{apply: internalrouter.WithContextPooling()}
}

// WithHeadFallbackToGet serves HEAD requests through the matching GET handler
// when no HEAD route is registered.
func WithHeadFallbackToGet() RouterOption {
	return RouterOption{apply: internalrouter.WithHeadFallbackToGet()}
}

// WithMaxBodyBytes limits request body size for Bind-based JSON and form
// parsing. Use it to protect generated CRUD endpoints from unbounded payloads.
func WithMaxBodyBytes(n int64) RouterOption {
	return RouterOption{apply: internalrouter.WithMaxBodyBytes(n)}
}

func toInternalRouterOptions(opts []RouterOption) []internalrouter.RouterOption {
	if len(opts) == 0 {
		return nil
	}
	internal := make([]internalrouter.RouterOption, 0, len(opts))
	for _, opt := range opts {
		if opt.apply != nil {
			internal = append(internal, opt.apply)
		}
	}
	return internal
}
