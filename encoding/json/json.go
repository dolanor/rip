package json

import (
	"encoding/json"

	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.Codec{
	NewEncoder: encoding.WrapEncoder(json.NewEncoder),
	NewDecoder: encoding.WrapDecoder(json.NewDecoder),
}

var MimeTypes = []string{
	"application/json",
}
