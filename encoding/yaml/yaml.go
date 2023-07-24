package yaml

import (
	"gopkg.in/yaml.v3"

	"github.com/dolanor/rip/encoding"
)

func init() {

	encoding.RegisterCodec("application/yaml", encoding.Codec{NewEncoder: encoding.WrapEncoder(yaml.NewEncoder), NewDecoder: encoding.WrapDecoder(yaml.NewDecoder)})
}
