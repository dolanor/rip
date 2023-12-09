package json

import (
	"encoding/json"

	"github.com/dolanor/rip/encoding"
)

var Codec = encoding.WrapCodec(json.NewEncoder, json.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/json",
}
