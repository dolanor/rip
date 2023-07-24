package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/dolanor/rip/encoding"
)

func init() {

	encoding.RegisterCodec("application/msgpack", encoding.Codec{NewEncoder: encoding.WrapEncoder(msgpack.NewEncoder), NewDecoder: encoding.WrapDecoder(msgpack.NewDecoder)})
}
