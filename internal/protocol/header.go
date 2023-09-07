package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/shadowsocks/go-shadowsocks2/socks"
	"io"
)

const (
	marker = 0b10000000

	// protocol
	pTcp  = 0
	pUdp  = 1
	pUnix = 2

	// address type
	aIpv4     = 0 << 4
	aIpv6     = 1 << 4
	aHostname = 2 << 4
	aString   = 3 << 4
)

type Header struct {
	protocol string
	address  string
}

var _ io.WriterTo = &Header{}
var _ io.ReaderFrom = &Header{}

func (h *Header) WriteTo(writer io.Writer) (int64, error) {
	var magic byte = marker

	switch h.protocol {
	case "tcp":
		magic |= pTcp
	case "udp":
		magic |= pUdp
	case "unix":
		magic |= pUnix
	default:
		return 0, fmt.Errorf("unsupported protocol %s", h.protocol)
	}

	var addrBuf []byte
	switch h.address {
	case "tcp", "udp":
		addr := socks.ParseAddr(h.address)
		if addr == nil {
			return 0, errors.New("invalid address")
		}
		var flag byte
		switch addr[0] {
		case socks.AtypIPv4:
			flag = aIpv4
		case socks.AtypIPv6:
			flag = aIpv6
		case socks.AtypDomainName:
			flag = aHostname
		default:
			return 0, fmt.Errorf("invalid atype %d", addr[0])
		}
		magic |= flag << 4
		addrBuf = addr[1:]

	case "unix":
		magic |= aString
		addrBuf = append(binary.AppendVarint(nil, int64(len(h.address))), []byte(h.address)...)

	default:
		return 0, fmt.Errorf("unsupported protocol %s", h.protocol)
	}

	if _, err := writer.Write([]byte{magic}); err != nil {
		return 0, err
	}
	if _, err := writer.Write(addrBuf); err != nil {
		return 1, err
	}

	return int64(1 + len(addrBuf)), nil
}

func (h *Header) ReadFrom(reader io.Reader) (n int64, err error) {
	b := make([]byte, 1)
	var cnt int64

	if n, err := io.ReadFull(reader, b); err != nil {
		return int64(n), err
	}
	cnt += n

	magic := b[0]
	if magic&marker == 0 {
		return cnt, errors.New("invalid marker")
	}
	protocolType := magic & 0b00001111
	addressType := magic & 0b01110000
	switch protocolType {
	case pTcp, pUdp:
		if protocolType == pTcp {
			h.protocol = "tcp"
		} else {
			h.protocol = "udp"
		}
		switch addressType {
		case aIpv4:

		}
	}

	return cnt, nil
}
