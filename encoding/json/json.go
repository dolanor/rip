package json

import (
	"encoding/json"

	"github.com/dolanor/rip/encoding"
)

func init() {
	codec := encoding.Codec{
		NewEncoder: encoding.WrapEncoder(json.NewEncoder),
		NewDecoder: encoding.WrapDecoder(json.NewDecoder),
	}

	encoding.RegisterCodec(codec, "application/json")
}
