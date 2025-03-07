package rip

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
)

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

func defaultOpenAPIHandler(schemaURL string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="referrer" content="same-origin" />
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<title>OpenAPI specification</title>
	<script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
	<link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css" />
</head>
<body style="height: 100vh;">
	<elements-api
		apiDescriptionUrl="` + schemaURL + `"
		layout="responsive"
		router="hash"
		tryItCredentialsPolicy="same-origin"
	/>
</body>
</html>`))
	})
}
