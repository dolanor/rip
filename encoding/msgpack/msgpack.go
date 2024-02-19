package msgpack

import (
	"github.com/dolanor/rip/encoding/codecwrap"
	"github.com/vmihailenco/msgpack/v5"
)

var Codec = codecwrap.Wrap(msgpack.NewEncoder, msgpack.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/msgpack",
}
