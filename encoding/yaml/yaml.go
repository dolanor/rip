package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/dolanor/rip/encoding"
)

func init() {
	codec := encoding.Codec{
		NewEncoder: encoding.WrapEncoder(yaml.NewEncoder),
		NewDecoder: encoding.WrapDecoder(yaml.NewDecoder),
	}

	encoding.RegisterCodec(codec, "application/yaml")
}
