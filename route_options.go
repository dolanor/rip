package rip

import (
	"maps"
	"slices"

	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/json"
)

func newDefaultOptions() *RouteOptions {
	ro := NewRouteOptions()

	ro.codecs.Register(json.Codec)

	return ro
}

var defaultOptions = newDefaultOptions()

// RouteOptions allows to pass options to the route handler.
// It make each route able to have its own set of middlewares
// or codecs.
// It also allows to be reused betwenn multiple routes.
type RouteOptions struct {
	middlewares     []Middleware
	codecs          encoding.Codecs
	statusMap       StatusMap
	listPageSize    int
	listPageSizeMax int
}

func NewRouteOptions() *RouteOptions {
	return &RouteOptions{
		codecs: encoding.Codecs{
			Codecs: map[string]encoding.Codec{},
		},
		listPageSize:    20,
		listPageSizeMax: 100,
	}
}

func (ro *RouteOptions) WithCodecs(codecs ...encoding.Codec) *RouteOptions {
	newRO := cloneRouteOptions(*ro)
	for _, c := range codecs {
		newRO.codecs.Register(c)
	}
	return &newRO
}

func (ro *RouteOptions) WithMiddlewares(middlewares ...Middleware) *RouteOptions {
	newRO := cloneRouteOptions(*ro)
	newRO.middlewares = append(newRO.middlewares, middlewares...)
	return &newRO
}

type StatusMap map[error]int

// WithErrors maps errors with an HTTP status code.
func (ro *RouteOptions) WithErrors(statusMap StatusMap) *RouteOptions {
	newRO := cloneRouteOptions(*ro)
	newRO.statusMap = statusMap
	return &newRO
}

// WithListPageSize configures the number of entities displayed in a list page.
func (ro *RouteOptions) WithListPageSize(pageSize int) *RouteOptions {
	newRO := cloneRouteOptions(*ro)
	newRO.listPageSize = pageSize
	return &newRO
}

// WithListPageSizeMax configures the maximum number of entities displayed in a list page.
func (ro *RouteOptions) WithListPageSizeMax(pageSizeMax int) *RouteOptions {
	newRO := cloneRouteOptions(*ro)
	newRO.listPageSizeMax = pageSizeMax
	return &newRO
}

func cloneRouteOptions(ro RouteOptions) RouteOptions {
	middlewares := slices.Clone(ro.middlewares)
	codecs := maps.Clone(ro.codecs.Codecs)

	orderedMimeTypes := slices.Clone(ro.codecs.OrderedMimeTypes)

	return RouteOptions{
		middlewares: middlewares,
		codecs: encoding.Codecs{
			Codecs:           codecs,
			OrderedMimeTypes: orderedMimeTypes,
		},
		listPageSize:    ro.listPageSize,
		listPageSizeMax: ro.listPageSizeMax,
	}
}
