package proto

import "fmt"

type PayloadKind interface {
	Flag() uint8
	valid() bool
}

type payloadKind uint8

var _ PayloadKind = kindMin

const (
	KindStream payloadKind = 1
	KindPacket payloadKind = 2

	kindMin = KindStream
	kindMax = KindPacket
)

func NewPayloadKind(kind uint8) (PayloadKind, error) {
	p := payloadKind(kind)
	if !p.valid() {
		return nil, fmt.Errorf("invalid payload kind %d", kind)
	}
	return p, nil
}

func (p payloadKind) Flag() uint8 {
	return uint8(p)
}

func (p payloadKind) valid() bool {
	return kindMin <= p && p <= kindMax
}

func (p payloadKind) String() string {
	switch p {
	case KindStream:
		return "stream"
	case KindPacket:
		return "packet"
	default:
		return fmt.Sprintf("kind_%d", p)
	}
}
