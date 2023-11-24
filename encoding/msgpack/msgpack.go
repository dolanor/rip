package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.Codec{
	NewEncoder: encoding.WrapEncoder(msgpack.NewEncoder),
	NewDecoder: encoding.WrapDecoder(msgpack.NewDecoder),
}

var MimeTypes = []string{
	"application/msgpack",
}
