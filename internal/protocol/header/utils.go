package header

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
)

func packAddress(address string) (byte, []byte, error) {
	var addrType byte
	var buf []byte
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return addrType, nil, err
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			addrType = aIpv4
			buf = make([]byte, net.IPv4len+2)
			copy(buf, ip4)
		} else {
			addrType = aIpv6
			buf = make([]byte, net.IPv6len+2)
			copy(buf, ip)
		}
	} else {
		if len(host) > math.MaxUint8 {
			return 0, nil, fmt.Errorf("hostname too large")
		}
		addrType = aDomain
		buf = make([]byte, 1+len(host)+2)
		buf[0] = byte(len(host))
		copy(buf[1:], host)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > math.MaxUint16 {
		return 0, nil, fmt.Errorf("invalid port %s", portStr)
	}

	buf[len(buf)-2], buf[len(buf)-1] = byte(port>>8), byte(port)
	return addrType, buf, nil
}

func readAddress(reader io.Reader, addrType byte) (string, error) {
	var buf []byte
	var host string
	switch addrType {
	case aIpv4:
		buf = make([]byte, net.IPv4len)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return "", err
		}
		host = net.IP(buf).String()
	case aIpv6:
		buf = make([]byte, net.IPv6len)
		if _, err := io.ReadFull(reader, buf); err != nil {
			return "", err
		}
		host = net.IP(buf).String()
	case aDomain:
		buf = make([]byte, math.MaxUint8)
		if _, err := io.ReadFull(reader, buf[:1]); err != nil {
			return "", err
		}
		length := int(buf[0])
		if _, err := io.ReadFull(reader, buf[:length]); err != nil {
			return "", err
		}
		host = string(buf[:length])
	default:
		return "", fmt.Errorf("unsupported address type %d", addrType)
	}

	if len(buf) < 2 {
		buf = make([]byte, 2)
	}

	if _, err := io.ReadFull(reader, buf[:2]); err != nil {
		return "", err
	}

	port := (uint16(buf[0]) << 8) | uint16(buf[1])
	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
}

func packString(target string) (byte, []byte, error) {
	return aString, append(binary.AppendVarint(nil, int64(len(target))), []byte(target)...), nil
}

func readString(reader io.Reader, addrType byte) (string, error) {
	if addrType != aString {
		return "", fmt.Errorf("addrType should be %d, but found %d", aString, addrType)
	}
	length, err := binary.ReadVarint(newByteReader(reader))
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return "", err
	}
	return string(buf), err
}

type byteReader struct {
	r io.Reader
}

func (br *byteReader) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := br.r.Read(buf[:])
	return buf[0], err
}

func newByteReader(r io.Reader) io.ByteReader {
	return &byteReader{r: r}
}
