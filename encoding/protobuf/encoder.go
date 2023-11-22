package protobuf

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
)

type encoder struct {
	w io.Writer
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{
		w: w,
	}
}

func (e *encoder) Encode(v any) error {
	switch v.(type) {
	case proto.Message:
	default:
		return fmt.Errorf("protobuf encode: bad message format: %T", v)
	}

	b, err := proto.Marshal(v.(proto.Message))
	if err != nil {
		return fmt.Errorf("protobuf encode: protobuf marshal: %w", err)
	}

	_, err = e.w.Write(b)
	if err != nil {
		return fmt.Errorf("protobuf encode: writer write: %w", err)
	}

	return nil
}
