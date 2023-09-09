package config

import (
	"errors"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"strings"
)

const (
	ModeServer = "server"
	ModeClient = "client"
	ModeTool   = "tool"

	modeRoot = "root" // for root command
)

// Crypts values see github.com/shadowsocks/go-shadowsocks2/core/cipher.go PickCipher
var Crypts = []string{
	"AES-128-GCM",
	"AES-256-GCM",
	"CHACHA20-IETF-POLY1305",
}

// ParseTunnel split an tunnel definition into id + protocol + address
// Format: [id:]protocol://address
func ParseTunnel(tun string) (id string, protocol *proto.ProtocolMeta, address string, err error) {
	t := strings.Split(tun, "://")
	if len(t) != 2 {
		return "", nil, "", errors.New("invalid protocol separation, check your '://' divider")
	}

	idProtocol, address := t[0], t[1]

	var protoStr string
	t = strings.Split(idProtocol, ":")
	if len(t) == 1 {
		id, protoStr = "", t[0]
	} else if len(t) == 2 {
		id, protoStr = t[0], t[1]
	} else {
		return "", nil, "", errors.New("invalid protocol separation, check your heading 'id:'")
	}

	protocol, ok := proto.GetProtocolMeta(protoStr)
	if !ok || !protocol.Tunnel {
		return "", nil, "", fmt.Errorf("invalid tunnel protocol %s", t[0])
	}

	if err = ValidateAddress(address); err != nil {
		return "", nil, "", err
	}

	err = nil
	return
}

func ValidateTunnel(tun string) error {
	_, _, _, err := ParseTunnel(tun)
	return err
}
