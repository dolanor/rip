package rip

import (
	"net/http"
	"testing"

	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/xml"
)

func TestPreprocessRequest(t *testing.T) {
	rMethod := http.MethodPost
	hMethod := http.MethodPost
	h := http.Header{
		"Accept":       []string{"application/json; q=0.9", "text/xml"},
		"Content-Type": []string{"application/json"},
	}
	rURLPath := "/whatever/entity/id"

	options := NewRouteOptions().WithCodecs(json.Codec, xml.Codec)
	_, accept, contentType, err := preprocessRequest(rMethod, hMethod, h, rURLPath, options)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case accept != "text/xml",
		contentType != "application/json":
		t.Fatal("could not find correct accept/content type")
	}
}
