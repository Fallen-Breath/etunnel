package header

import (
	"io"
)

const (
	CodeOk      uint8 = 0
	CodeBadId   uint8 = 1
	CodeBadKind uint8 = 2
)

type RspHead struct {
	Code uint8
}

func (h *RspHead) MarshalTo(writer io.Writer) error {
	if _, err := writer.Write([]byte{h.Code}); err != nil {
		return err
	}
	return nil
}

func (h *RspHead) UnmarshalFrom(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return err
	}

	h.Code = buf[0]
	return nil
}
