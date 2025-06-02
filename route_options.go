package rip

import (
	"log/slog"

	"github.com/dolanor/rip/encoding"
)

type StatusMap map[error]int

type entityRouteConfig struct {
	codecs          encoding.Codecs
	logger          *slog.Logger
	middlewares     []Middleware
	statusMap       StatusMap
	listPageSize    int
	listPageSizeMax int
}

// EntityRouteOption is the optional configuration for a [EntityRoute].
type EntityRouteOption func(cfg *entityRouteConfig)

// WithEntityRouteLogger configures the logger used for this route.
func WithEntityRouteLogger(logger *slog.Logger) EntityRouteOption {
	return func(cfg *entityRouteConfig) {
		cfg.logger = logger
	}
}

// WithCodecs configures the encoding available for this route.
func WithCodecs(codecs ...encoding.Codec) EntityRouteOption {
	return func(cfg *entityRouteConfig) {
		if cfg.codecs.Codecs == nil {
			cfg.codecs.Codecs = map[string]encoding.Codec{}
		}

		for _, c := range codecs {
			cfg.codecs.Register(c)
		}
	}
}

// WithErrors maps errors with an HTTP status code for this route.
func WithErrors(statusMap StatusMap) EntityRouteOption {
	return func(cfg *entityRouteConfig) {
		cfg.statusMap = statusMap
	}
}

// WithListPage configures the number of entities displayed in a list page for this route.
func WithListPage(size int, max int) EntityRouteOption {
	return func(cfg *entityRouteConfig) {
		cfg.listPageSize = size
		cfg.listPageSizeMax = max
	}
}

// WithMiddleware configures the middlewares for this route.
func WithMiddlewares(middlewares ...Middleware) EntityRouteOption {
	return func(cfg *entityRouteConfig) {
		for _, m := range middlewares {
			cfg.middlewares = append(cfg.middlewares, m)
		}
	}
}
