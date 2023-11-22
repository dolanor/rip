package protobuf

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

type decoder struct {
	reader io.Reader
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{
		reader: r,
	}
}

func (d *decoder) Decode(v any) error {
	m, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("protobuf decode: bad message format: %T", v)
	}

	b, err := io.ReadAll(d.reader)
	if err != nil {
		return fmt.Errorf("protobuf decode: %w", err)
	}

	err = proto.Unmarshal(b, m)
	if err != nil {
		return fmt.Errorf("protobuf decode: protobuf unmarshal: %w", err)
	}

	return nil
}
