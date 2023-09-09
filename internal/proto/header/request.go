package header

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"io"
	"math"
)

const (
	marker     uint8 = 0b10110000
	markerMask uint8 = 0b11110000
)

type ReqHead struct {
	Kind proto.PayloadKind
	Id   string
}

func (h *ReqHead) MarshalTo(writer io.Writer) error {
	if len(h.Id) > math.MaxUint8 {
		return fmt.Errorf("id too large")
	}

	buf := make([]byte, 1+1+len(h.Id))
	buf[0] = marker | h.Kind.Flag()
	buf[1] = byte(len(h.Id))
	copy(buf[2:], h.Id)

	if _, err := writer.Write(buf); err != nil {
		return err
	}
	return nil
}

func (h *ReqHead) UnmarshalFrom(reader io.Reader) error {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return err
	}

	k, err := proto.NewPayloadKind(buf[0] & ^markerMask)
	if err != nil {
		return err
	}

	buf = make([]byte, math.MaxUint8)
	if _, err := io.ReadFull(reader, buf[:1]); err != nil {
		return err
	}
	length := int(buf[0])
	if _, err := io.ReadFull(reader, buf[:length]); err != nil {
		return err
	}

	h.Kind = k
	h.Id = string(buf[:length])

	return nil
}
