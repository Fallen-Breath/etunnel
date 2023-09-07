package config

import (
	"errors"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/Fallen-Breath/etunnel/internal/utils"
	"net"
	"strings"
)

const (
	ModeClient = "client"
	ModeServer = "server"
)

// Crypts values see github.com/shadowsocks/go-shadowsocks2/core/cipher.go PickCipher
var Crypts = []string{
	"AES-128-GCM",
	"AES-256-GCM",
	"CHACHA20-IETF-POLY1305",
}

func ValidateCrypt(crypt string) error {
	for _, c := range Crypts {
		if c == crypt {
			return nil
		}
	}
	return fmt.Errorf("unknown crypt %s", crypt)
}

func ValidateAddress(listen string) error {
	_, err := net.ResolveTCPAddr("tcp", listen)
	return err
}

func ParseTunnel(tun string) (protocol, listen, target string, err error) {
	t := strings.Split(tun, "://")
	if len(t) != 2 {
		return "", "", "", errors.New("invalid protocol separation, check your '://' divider")
	}

	protocol = t[0]
	if !utils.Contains(proto.Protocols, protocol) {
		return "", "", "", fmt.Errorf("invalid protocol %s", protocol)
	}

	t = strings.Split(t[1], "/")
	if len(t) != 2 {
		return "", "", "", errors.New("invalid listen - target separation, check your '/' divider")
	}

	listen, target = t[0], t[1]
	if err = ValidateAddress(listen); err != nil {
		return "", "", "", err
	}
	if err = ValidateAddress(target); err != nil {
		return "", "", "", err
	}

	err = nil
	return
}

func ValidateTunnel(tun string) error {
	_, _, _, err := ParseTunnel(tun)
	return err
}
