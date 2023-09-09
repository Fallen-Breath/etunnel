package proto

type AddressType interface {
	Value() uint8
	valid() bool
}

type addressType uint8

var _ AddressType = addressType(0)

const (
	AddrRegular  addressType = 1
	AddrFilePath addressType = 2
)

func (a addressType) Value() uint8 {
	return uint8(a)
}

func (a addressType) valid() bool {
	return 1 <= a && a <= 2
}
