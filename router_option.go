package mux

import internalrouter "github.com/fgrzl/mux/internal/router"

type RouterOption struct {
	apply internalrouter.RouterOption
}

func WithTitle(title string) RouterOption {
	return RouterOption{apply: internalrouter.WithTitle(title)}
}

func WithSummary(summary string) RouterOption {
	return RouterOption{apply: internalrouter.WithSummary(summary)}
}

func WithDescription(description string) RouterOption {
	return RouterOption{apply: internalrouter.WithDescription(description)}
}

func WithTermsOfService(url string) RouterOption {
	return RouterOption{apply: internalrouter.WithTermsOfService(url)}
}

func WithVersion(version string) RouterOption {
	return RouterOption{apply: internalrouter.WithVersion(version)}
}

func WithContact(name, url, email string) RouterOption {
	return RouterOption{apply: internalrouter.WithContact(name, url, email)}
}

func WithLicense(name, url string) RouterOption {
	return RouterOption{apply: internalrouter.WithLicense(name, url)}
}

func WithClientURL(url string) RouterOption {
	return RouterOption{apply: internalrouter.WithClientURL(url)}
}

func WithContextPooling() RouterOption {
	return RouterOption{apply: internalrouter.WithContextPooling()}
}

func WithHeadFallbackToGet() RouterOption {
	return RouterOption{apply: internalrouter.WithHeadFallbackToGet()}
}

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
