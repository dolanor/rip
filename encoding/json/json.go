package yaml

import (
	"encoding/json"

	"github.com/dolanor/rip/encoding"
)

func init() {

	encoding.RegisterCodec("application/json", encoding.Codec{NewEncoder: encoding.WrapEncoder(json.NewEncoder), NewDecoder: encoding.WrapDecoder(json.NewDecoder)})
}
