package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/json"
)

func toUpper(ctx context.Context, input string) (output string, err error) {
	output = strings.ToUpper(input)
	return output, nil
}

func main() {
	handler := rip.Handle(http.MethodPost, toUpper, rip.WithCodecs(json.Codec))
	http.HandleFunc("/uppercase/", handler)

	slog.Info("listening on :8888")
	http.ListenAndServe(":8888", nil)
}
