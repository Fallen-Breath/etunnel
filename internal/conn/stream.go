package conn

import (
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	"net"
)

type encryptedStreamConn struct {
	net.Conn
	innerConn StreamConn
}

var _ StreamConn = &encryptedStreamConn{}

func (e *encryptedStreamConn) CloseRead() error {
	return e.innerConn.CloseRead()
}

func (e *encryptedStreamConn) CloseWrite() error {
	return e.innerConn.CloseWrite()
}

func NewEncryptedStreamConn(conn StreamConn, cipher sscore.Cipher) StreamConn {
	sc := cipher.StreamConn(conn)
	return &encryptedStreamConn{
		Conn:      sc,
		innerConn: conn,
	}
}
