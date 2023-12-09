package msgpack

import (
	"github.com/vmihailenco/msgpack/v5"

	"github.com/dolanor/rip/encoding"
)

var Codec = encoding.WrapCodec(msgpack.NewEncoder, msgpack.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/msgpack",
}
