package mux

import "net/url"

type RouterOptions struct {
	openapi   *InfoObject
	clientURL *url.URL
}

func (o *RouterOptions) SetClientURL(clientURL *url.URL) {
	o.clientURL = clientURL
}

type RouterOption func(*RouterOptions)

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
		o.openapi.Contact = &ContactObject{
			Name:  name,
			URL:   url,
			Email: email,
		}
	}
}

func WithLicense(name, url string) RouterOption {
	return func(o *RouterOptions) {
		initInfo(o)
		o.openapi.License = &LicenseObject{
			Name: name,
			URL:  url,
		}
	}
}

// helper
func initInfo(o *RouterOptions) {
	if o.openapi == nil {
		o.openapi = &InfoObject{}
	}
}
