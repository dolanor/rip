package rip

type RouterOption func(cfg *RouterConfig)

func WithRouterAPITitle(title string) RouterOption {
	return func(cfg *RouterConfig) {
		cfg.APITitle = title
	}
}

func WithRouterAPIDescription(desc string) RouterOption {
	return func(cfg *RouterConfig) {
		cfg.APIDescription = desc
	}
}

func WithRouterAPIVersion(version string) RouterOption {
	return func(cfg *RouterConfig) {
		cfg.APIVersion = version
	}
}
