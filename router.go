package rip

import (
	"net/http"
	"runtime/debug"

	"github.com/dolanor/rip/dist/css"
	"github.com/dolanor/rip/dist/js"
	"github.com/getkin/kin-openapi/openapi3"
)

var DefaultRouter = &Router{}

// Router allows to regroup all entity routes together and hold a description of the API in
// the OpenAPI format.
type Router struct {
	handler     HTTPServeMux
	openapiSpec openapi3.T
}

// HTTPServeMux is an interface for HTTP multiplexers
type HTTPServeMux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Handle(pattern string, handler http.Handler)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type RouterConfig struct {
	APITitle       string
	APIDescription string
	APIVersion     string
}

// NewRouter creates a new [Router], using mux for the http routing and
// takes options for configuring the documentation.
func NewRouter(mux HTTPServeMux, options ...RouterOption) *Router {
	var version string
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		version = buildInfo.Main.Version
	}

	cfg := RouterConfig{
		APITitle:   "API",
		APIVersion: version,
	}

	for _, o := range options {
		o(&cfg)
	}

	openAPISpec := newOpenApiSpec(cfg.APITitle, cfg.APIDescription, cfg.APIVersion)

	mux.Handle("/api-docs/js/", http.StripPrefix("/api-docs/js/", http.FileServerFS(js.FS)))
	mux.Handle("/api-docs/css/", http.StripPrefix("/api-docs/css/", http.FileServerFS(css.FS)))
	mux.HandleFunc("/api-docs/", handleSwaggerUI(cfg.APITitle))
	mux.HandleFunc("/api-docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		b, err := openAPISpec.MarshalJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(b)
	})

	return &Router{
		handler:     mux,
		openapiSpec: openAPISpec,
	}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// we just forward to the already registered routes from [Router.Handle] or
	// [Router.HandleFunc]
	rt.handler.ServeHTTP(w, r)
}

func (rt *Router) HandleFunc(path string, handler func(w http.ResponseWriter, r *http.Request)) {
	rt.handler.HandleFunc(path, handler)
}

func (rt *Router) HandleRoute(route Route) { //path string, handler http.HandlerFunc, openAPISchema *openapi3.T) {
	for k, v := range route.OpenAPISchema().Components.Schemas {
		rt.openapiSpec.Components.Schemas[k] = v
	}

	for k, v := range route.OpenAPISchema().Paths.Map() {
		rt.openapiSpec.Paths.Set(k, v)
	}

	rt.handler.HandleFunc(route.Path(), route.Handler())
}

func (rt *Router) PrintInfo() {
	// fmt.Printf("%+v\n", rt.openapiSpec.Info)
}
