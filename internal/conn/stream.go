package conn

import (
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	"net"
	"reflect"
)

func NewEncryptedStreamConn(conn StreamConn, cipher sscore.Cipher) StreamConn {
	sc := cipher.StreamConn(conn)
	return getStreamConn(sc)
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
