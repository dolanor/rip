package main

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/providers/mapprovider"
)

func main() {
	type Album struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Artist string `json:"artist"`
	}

	ap := mapprovider.New[Album](slog.Default())
	r := rip.NewRouter(http.NewServeMux())
	ro := rip.NewRouteOptions().
		WithCodecs(
			json.Codec,
			html.NewEntityCodec("/albums/", html.WithServeMux(r)),
			html.NewEntityFormCodec("/albums/", html.WithServeMux(r)),
		)

	r.HandleFunc(rip.HandleEntities("/albums/", ap, ro))

	// Generate OpenAPI documentation
	config := rip.OpenAPIConfig{
		Info: rip.OpenAPIInfo{
			Title:        "Album API",
			Description:  "REST API for managing music albums",
			Version:      "1.0.0",
			ContactName:  "API Support",
			ContactEmail: "support@example.com",
		},
		ServerURLs: []string{"http://localhost:9999"},
		OutputPath: "openapi.json",
	}

	err := r.GenerateOpenAPI(config)
	if err != nil {
		log.Fatalf("Failed to generate OpenAPI documentation: %v", err)
	}
	log.Printf("OpenAPI documentation generated at %s", config.OutputPath)

	// Start the HTTP server
	log.Printf("Starting HTTP server on :9999")
	log.Fatal(http.ListenAndServe(":9999", r))
}
