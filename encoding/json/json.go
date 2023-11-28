package json

import (
	"encoding/json"

	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.WrapCodec(json.NewEncoder, json.NewDecoder)

var MimeTypes = []string{
	"application/json",
}
