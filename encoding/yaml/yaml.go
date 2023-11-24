package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.Codec{
	NewEncoder: encoding.WrapEncoder(yaml.NewEncoder),
	NewDecoder: encoding.WrapDecoder(yaml.NewDecoder),
}

var MimeTypes = []string{
	"text/vnd.yaml",
	"text/yaml",
	"text/-xyaml",
	"application/x-yaml",
}
