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
	r.HandleRoute(rip.NewEntityRoute("/albums/", ap, rip.WithCodecs(
		json.Codec,
		html.NewEntityCodec("/albums/", html.WithServeMux(r)),
		html.NewEntityFormCodec("/albums/", html.WithServeMux(r)),
	),
	))

	slog.Info("server started listening", "port", "9999")
	err := http.ListenAndServe(":9999", r)
	if err != nil {
		slog.Error("listen and serve", "error", err)
	}
}
