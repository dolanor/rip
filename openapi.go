package rip

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed swagger_ui.gohtml
var swaggerUITemplate string

func newOpenApiSpec(title, description, version string) openapi3.T {
	info := &openapi3.Info{
		Title:       title,
		Description: description,
		Version:     version,
	}
	spec := openapi3.T{
		OpenAPI: "3.0.3",
		Info:    info,
		Paths:   &openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas:       make(map[string]*openapi3.SchemaRef),
			RequestBodies: make(map[string]*openapi3.RequestBodyRef),
			Responses:     make(map[string]*openapi3.ResponseRef),
		},
	}
	return spec
}

func handleSwaggerUI(title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		tmpl, err := template.New("swagger-ui").Parse(swaggerUITemplate)
		if err != nil {
			http.Error(w, "Error rendering UI template", http.StatusInternalServerError)
			return
		}

		data := struct {
			Title string
			Path  string
		}{
			Title: title,
			Path:  "/api-docs/swagger.json",
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error rendering UI template", http.StatusInternalServerError)
		}
	}
}
