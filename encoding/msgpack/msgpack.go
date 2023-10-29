package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/dolanor/rip/encoding"
)

func init() {
	codec := encoding.Codec{
		NewEncoder: encoding.WrapEncoder(msgpack.NewEncoder),
		NewDecoder: encoding.WrapDecoder(msgpack.NewDecoder),
	}

	encoding.RegisterCodec(codec, "application/msgpack")
}
