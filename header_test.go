package rip

import (
	"errors"
	"testing"

	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/encoding/xml"
)

func TestChooseHeaderValue(t *testing.T) {
	cases := map[string]struct {
		in  map[string][]string
		exp string
	}{
		"1 value":                     {map[string][]string{"a": {"application/json"}}, "application/json"},
		"2 values":                    {map[string][]string{"a": {"text/xml", "application/json"}}, "text/xml"},
		"2x2 values":                  {map[string][]string{"a": {"text/xml, application/json", "application/json, text/plaintext"}}, "text/xml"},
		"2x2 values with q":           {map[string][]string{"a": {"text/xml; q=0.7, application/json; q=0.1", "application/json; q=0.3, text/plaintext;q=0.71"}}, "text/xml"},
		"2x2 values with q and other": {map[string][]string{"a": {"text/xml; nope; q=0.7, application/json; q=0.1", "application/json; q=0.3, text/plaintext; other;q=0.71"}}, "text/xml"},
		"nothing":                     {map[string][]string{"a": {""}}, ""},
	}

	options := NewRouteOptions().WithCodecs(json.Codec, xml.Codec)
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := contentNegociateBestHeaderValue(c.in, "a", options.codecs.OrderedMimeTypes)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.exp {
				t.Fatalf("result not equal: got %v, expected %v", got, c.exp)
			}
		})
	}
}

func TestBadHeaderQArgValue(t *testing.T) {
	in := map[string][]string{"a": {"application/json; q=HAHA"}}
	expErr := Error{Source: ErrorSource{Header: "a"}}

	options := NewRouteOptions().WithCodecs(json.Codec, xml.Codec)
	_, err := contentNegociateBestHeaderValue(in, "a", options.codecs.OrderedMimeTypes)
	if err == nil {
		t.Fatal(err)
	}

	var badQArgErr Error
	if errors.As(err, &badQArgErr) {
		if badQArgErr.Code != errorCodeBadQArg {
			t.Fatal(err)
		}
		if badQArgErr.Source.Header != expErr.Source.Header {
			t.Fatal(err)
		}
	}
}
