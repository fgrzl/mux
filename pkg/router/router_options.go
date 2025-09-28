package router

import (
	"log/slog"
	"net/url"

	openapi "github.com/fgrzl/mux/pkg/openapi"
)

type RouterOptions struct {
	openapi   *openapi.InfoObject
	clientURL *url.URL
	// HeadFallbackToGet, when true, serves HEAD requests using the GET handler
	// if a HEAD route is not explicitly registered for the matched path.
	HeadFallbackToGet bool
	// MaxBodyBytes sets the maximum size of request bodies for JSON/form binds.
	// If zero or negative, a default of 1MB is applied.
	MaxBodyBytes int64
	// ContextPooling enables sync.Pool reuse of RouteContext instances to
	// reduce allocations on the hot path.
	ContextPooling bool
}

func (o *RouterOptions) SetClientURL(clientURL *url.URL) {
	o.clientURL = clientURL
}

type RouterOption func(*RouterOptions)

// WithClientURL parses the provided clientURL string and sets it on the WebServer.
// If parsing fails, logs an error and does not set the clientURL.
func WithClientURL(clientURL string) RouterOption {
	return func(o *RouterOptions) {
		u, err := url.Parse(clientURL)
		if err != nil {
			slog.Error("Invalid clientURL", "clientURL", clientURL, "error", err)
			return
		}
		o.clientURL = u
	}
}

func WithTitle(title string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.Title = title
	}
}

func WithSummary(summary string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.Summary = summary
	}
}

func WithDescription(description string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.Description = description
	}
}

func WithTermsOfService(url string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.TermsOfService = url
	}
}

func WithVersion(version string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.Version = version
	}
}

func WithContact(name, url, email string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.Contact = &openapi.ContactObject{
			Name:  name,
			URL:   url,
			Email: email,
		}
	}
}

func WithLicense(name, url string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.License = &openapi.LicenseObject{
			Name: name,
			URL:  url,
		}
	}
}

// WithHeadFallbackToGet enables serving HEAD requests via the GET handler
// when a HEAD route isn't registered for a matched path.
func WithHeadFallbackToGet() RouterOption {
	return func(o *RouterOptions) {
		o.HeadFallbackToGet = true
	}
}

// WithMaxBodyBytes sets the maximum allowed size for request bodies used in binding.
// A value <= 0 will cause the router to use a 1MB default.
func WithMaxBodyBytes(n int64) RouterOption {
	return func(o *RouterOptions) {
		o.MaxBodyBytes = n
	}
}

// WithContextPooling enables pooling of RouteContext instances for reduced allocations.
func WithContextPooling() RouterOption {
	return func(o *RouterOptions) {
		o.ContextPooling = true
	}
}

// helper
func initInfo(o *RouterOptions) {
	if o.openapi == nil {
		o.openapi = &openapi.InfoObject{}
	}
}
