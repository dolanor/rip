package main

import (
	"log/slog"
	"net/http"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/html"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/providers/mapprovider"
)

func main() {
	type Album struct {
		ID     string
		Name   string
		Artist string
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

	http.ListenAndServe(":9999", r)
}
