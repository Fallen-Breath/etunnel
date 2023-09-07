package conn

import "net"

type ReadWriteClosable interface {
	CloseRead() error
	CloseWrite() error
}

type StreamConn interface {
	net.Conn
	ReadWriteClosable
}

type PacketConn interface {
	net.PacketConn
}
