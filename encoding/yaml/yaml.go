package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/dolanor/rip/encoding"
)

var Codec = encoding.WrapCodec(yaml.NewEncoder, yaml.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"text/vnd.yaml",
	"text/yaml",
	"text/x-yaml",
	"application/x-yaml",
}
