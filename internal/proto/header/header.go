package header

import (
	"errors"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"io"
)

const (
	marker = 0b10000000

	// protocol
	pTcp  = 0
	pUdp  = 1
	pUnix = 2

	// address type
	aIpv4   = 0 << 4
	aIpv6   = 1 << 4
	aDomain = 2 << 4
	aString = 3 << 4
)

type Header struct {
	Protocol string
	Target   string
}

func (h *Header) MarshalTo(writer io.Writer) error {
	var magic byte = marker
	var targetPacker func(string) (byte, []byte, error)

	switch h.Protocol {
	case proto.Tcp:
		magic |= pTcp
		targetPacker = packAddress
	case proto.Udp:
		magic |= pUdp
		targetPacker = packAddress
	case proto.Unix:
		magic |= pUnix
		targetPacker = packString
	default:
		return fmt.Errorf("unsupported protocol %s", h.Protocol)
	}

	addrType, addrBuf, err := targetPacker(h.Target)
	if err != nil {
		return fmt.Errorf("failed to pack address %s: %v", h.Target, h.Protocol)
	}
	magic |= addrType

	if _, err := writer.Write([]byte{magic}); err != nil {
		return err
	}
	if _, err := writer.Write(addrBuf); err != nil {
		return err
	}

	return nil
}

func (h *Header) UnmarshalFrom(reader io.Reader) error {
	b := make([]byte, 1)
	if _, err := io.ReadFull(reader, b); err != nil {
		return err
	}
	magic := b[0]
	if magic&marker == 0 {
		return errors.New("invalid marker")
	}

	protocolType := magic & 0b00001111
	addrType := magic & 0b01110000
	var targetReader func(io.Reader, byte) (string, error)

	switch protocolType {
	case pTcp:
		h.Protocol = proto.Tcp
		targetReader = readAddress
	case pUdp:
		h.Protocol = proto.Udp
		targetReader = readAddress
	case pUnix:
		h.Protocol = proto.Unix
		targetReader = readString
	default:
		return fmt.Errorf("unsupported protocol type %d", protocolType)
	}

	target, err := targetReader(reader, addrType)
	if err != nil {
		return fmt.Errorf("failed to read address with type %d: %v", addrType, h.Protocol)
	}
	h.Target = target

	return nil
}
