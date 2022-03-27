package rip

import (
	"encoding/json"
	"encoding/xml"
	"io"
)

var AvailableEncodings = []string{
	"text/json",
	"text/xml",
}

type Decoder interface {
	Decode(v interface{}) error
}

func contentTypeDecoder(r io.Reader, contentTypeHeader string) Decoder {
	// TODO use a map[string]Decoder to be able to extend it
	switch contentTypeHeader {
	case "text/xml":
		return xml.NewDecoder(r)
	case "text/json":
		fallthrough
	default:
		return json.NewDecoder(r)
	}
}

type Encoder interface {
	Encode(v interface{}) error
}

func acceptEncoder(w io.Writer, acceptHeader string) Encoder {
	switch acceptHeader {
	case "text/xml":
		return xml.NewEncoder(w)
	case "text/json":
		fallthrough
	default:
		return json.NewEncoder(w)
	}
}
