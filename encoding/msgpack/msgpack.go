package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.WrapCodec(msgpack.NewEncoder, msgpack.NewDecoder)

var MimeTypes = []string{
	"application/msgpack",
}
