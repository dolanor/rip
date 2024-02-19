package yaml

import (
	"github.com/dolanor/rip/encoding/codecwrap"
	"gopkg.in/yaml.v3"
)

var Codec = codecwrap.Wrap(yaml.NewEncoder, yaml.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"text/vnd.yaml",
	"text/yaml",
	"text/x-yaml",
	"application/x-yaml",
}
