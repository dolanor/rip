package rip

import (
	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/json"
)

func newDefaultOptions() *RouteOptions {
	ro := NewRouteOptions()

	ro.codecs.Register(json.Codec)

	return ro
}

var defaultOptions = newDefaultOptions()

type RouteOptions struct {
	middlewares []Middleware
	codecs      encoding.Codecs
}

func NewRouteOptions() *RouteOptions {
	return &RouteOptions{
		codecs: encoding.Codecs{
			Codecs: map[string]encoding.Codec{},
		},
	}
}

func (ro *RouteOptions) WithCodecs(codecs ...encoding.Codec) *RouteOptions {
	for _, c := range codecs {
		ro.codecs.Register(c)
	}
	return ro
}

func (ro *RouteOptions) WithMiddlewares(middlewares ...Middleware) *RouteOptions {
	ro.middlewares = append(ro.middlewares, middlewares...)
	return ro
}
