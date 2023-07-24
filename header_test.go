package rip

import (
	"testing"

	"github.com/dolanor/rip/encoding"
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
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := bestHeaderValue(c.in, "a", encoding.AvailableEncodings)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.exp {
				t.Fatalf("result not equal: got %v, expected %v", got, c.exp)
			}
		})
	}
}
