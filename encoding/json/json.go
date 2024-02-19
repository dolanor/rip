package json

import (
	"encoding/json"

	"github.com/dolanor/rip/encoding/codecwrap"
)

var Codec = codecwrap.Wrap(json.NewEncoder, json.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/json",
}
