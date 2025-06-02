package rip_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/dolanor/rip"
	"github.com/dolanor/rip/encoding/json"
)

func ExampleHandle() {
	toUpper := func(ctx context.Context, input string) (output string, err error) {
		output = strings.ToUpper(input)
		return output, nil
	}

	handler := rip.Handle(http.MethodPost, toUpper, rip.WithCodecs(json.Codec))

	http.HandleFunc("/uppercase/", handler)

	slog.Info("listening on :8888")
	http.ListenAndServe(":8888", nil)
}

func toUpper(ctx context.Context, input string) (output string, err error) {
	output = strings.ToUpper(input)
	return output, nil
}

func ExampleHandle_withClient() {
	handler := rip.Handle(http.MethodPost, toUpper, rip.WithCodecs(json.Codec))
	http.HandleFunc("/uppercase/", handler)

	ts := httptest.NewServer(http.DefaultServeMux)
	defer ts.Close()

	buf := bytes.NewBufferString(`"hello world"`)

	res, err := http.Post(ts.URL+"/uppercase/", "application/json", buf)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	greeting, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", greeting)
	// Output: "HELLO WORLD"
}
