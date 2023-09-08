package conn

import (
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	"net"
	"reflect"
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
		innerConn: getStreamConn(sc),
	}
}

func getStreamConn(conn net.Conn) StreamConn {
	val := reflect.ValueOf(conn)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	field := val.FieldByName("Conn")
	if !field.IsValid() {
		panic("field Conn is not valid")
	}

	return field.Interface().(StreamConn)
}
