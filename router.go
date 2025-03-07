package rip

import (
	"fmt"
	"net/http"
)

var DefaultRouter = &Router{}

type Router struct {
	handler HTTPServeMux
}

// HTTPServeMux is an interface for HTTP multiplexers
type HTTPServeMux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Handle(pattern string, handler http.Handler)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

func NewRouter(handler HTTPServeMux) *Router {
	return &Router{
		handler: handler,
	}
}

type EntityRoute[
	Ent any,
	EP EntityProvider[Ent],
] struct {
	provider EP
}

func (rt *EntityRoute[Ent, EP]) HandleEntities(
	urlPath string,
	ep EP,
	options *RouteOptions,
) (path string, handler http.HandlerFunc) {
	// end HandleEntities OMIT
	if options == nil {
		options = defaultOptions
	}

	if len(options.codecs.Codecs) == 0 {
		err := fmt.Sprintf("no codecs defined on route: %s", urlPath)
		panic(err)
	}

	return handleEntityWithPath(urlPath, ep.Create, ep.Get, ep.Update, ep.Delete, ep.List, options)
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// we just forward to the already registered routes from [Router.Handle] or
	// [Router.HandleFunc]
	rt.handler.ServeHTTP(w, r)
}

func (rt *Router) HandleFunc(pattern string, handler func(w http.ResponseWriter, r *http.Request)) {
	rt.handler.HandleFunc(pattern, handler)
}

func (rt *Router) Handle(pattern string, handler http.Handler) {
	rt.handler.Handle(pattern, handler)
}
