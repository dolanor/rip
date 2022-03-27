package rip

import (
	"testing"
)

func TestChooseHeaderValue(t *testing.T) {
	cases := map[string]struct {
		in  []string
		exp string
	}{
		"1 value":                     {[]string{"text/json"}, "text/json"},
		"2 values":                    {[]string{"text/xml", "text/json"}, "text/xml"},
		"2x2 values":                  {[]string{"text/xml, text/json", "text/json, text/plaintext"}, "text/xml"},
		"2x2 values with q":           {[]string{"text/xml; q=0.7, text/json; q=0.1", "text/json; q=0.3, text/plaintext;q=0.71"}, "text/xml"},
		"2x2 values with q and other": {[]string{"text/xml; nope; q=0.7, text/json; q=0.1", "text/json; q=0.3, text/plaintext; other;q=0.71"}, "text/xml"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := bestHeaderValue(c.in, AvailableEncodings)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.exp {
				t.Fatalf("result not equal: got %v, expected %v", got, c.exp)
			}
		})
	}
}
