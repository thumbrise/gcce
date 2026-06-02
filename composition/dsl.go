package composition

// --- DSL types ---

type Option interface {
	apply(cfg *config)
}

type ProvideOption interface {
	applyProvide(params *ProvideParams)
}

type (
	Constructor interface{}
	Interface   interface{}
)

// --- DSL functions ---

func Provide(ctor Constructor, options ...ProvideOption) Option {
	frame := stacktrace(0)

	return option(func(cfg *config) {
		cfg.provides = append(cfg.provides, provideOpt{
			frame:   frame,
			ctor:    ctor,
			options: options,
		})
	})
}

func As(iface Interface) ProvideOption {
	return provideOption(func(params *ProvideParams) {
		params.Interfaces = append(params.Interfaces, iface)
	})
}

func Meta(key, value string) ProvideOption {
	return provideOption(func(params *ProvideParams) {
		if params.Metadata == nil {
			params.Metadata = Metadata{}
		}

		params.Metadata[key] = value
	})
}
