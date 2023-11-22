package protobuf

import (
	"github.com/dolanor/rip/encoding"
)

func init() {
	codec := encoding.Codec{
		NewEncoder: encoding.WrapEncoder(newEncoder),
		NewDecoder: encoding.WrapDecoder(newDecoder),
	}

	encoding.RegisterCodec(codec, "application/json")
}
