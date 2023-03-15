package rip

import (
	"net/http"
	"testing"
)

func TestPreprocessRequest(t *testing.T) {
	rMethod := http.MethodPost
	hMethod := http.MethodPost
	h := http.Header{
		"Accept":       []string{"text/json; q=0.9", "text/xml"},
		"Content-Type": []string{"text/json"},
	}
	rURLPath := "/whatever/resource/id"

	// TODO: add some .xml/.json/.html
	_, accept, contentType, err := preprocessRequest(rMethod, hMethod, h, rURLPath)
	if err != nil {
		t.Fatal(err)
	}
	switch {
	case accept != "text/xml",
		contentType != "text/json":
		t.Fatal("could not find correct accept/content type")
	}
}
