package config

import (
	"errors"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"net"
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

func ValidateCrypt(crypt string) error {
	for _, c := range Crypts {
		if c == crypt {
			return nil
		}
	}
	return fmt.Errorf("unknown crypt %s", crypt)
}

func ValidateAddress(netAddr string) error {
	_, err := net.ResolveTCPAddr("tcp", netAddr)
	return err
}

func ParseCommunicateAddress(pa string) (protocol, address string, err error) {
	t := strings.Split(pa, "://")
	if len(t) == 1 {
		protocol = proto.Tcp // use tcp by default
	} else if len(t) == 2 {
		protocol = t[0]
		if ok := proto.CheckProtocol(protocol); !ok {
			return "", "", fmt.Errorf("invalid protocol %s", protocol)
		}
	} else {
		return "", "", errors.New("invalid protocol separation, check your '://' divider")
	}

	address = t[len(t)-1]
	if err := ValidateAddress(address); err != nil {
		return "", "", err
	}
	return protocol, address, nil
}

// ParseTunnel split an tunnel definition into listen + target
// Format: [protocol://]listen/target
// if protocol is missing, use tcp
func ParseTunnel(tun string) (protocol *proto.ProtocolMeta, listen, target string, err error) {
	t := strings.Split(tun, "://")
	if len(t) != 2 {
		return nil, "", "", errors.New("invalid protocol separation, check your '://' divider")
	}

	protocol, ok := proto.GetProtocolMeta(t[0])
	if !ok || !protocol.Tunnel {
		return nil, "", "", fmt.Errorf("invalid tunnel protocol %s", t[0])
	}

	t = strings.Split(t[1], "/")
	if len(t) != 2 {
		return nil, "", "", errors.New("invalid listen - target separation, check your '/' divider")
	}

	listen, target = t[0], t[1]
	if err = ValidateAddress(listen); err != nil {
		return nil, "", "", err
	}
	if err = ValidateAddress(target); err != nil {
		return nil, "", "", err
	}

	err = nil
	return
}

func ValidateTunnel(tun string) error {
	_, _, _, err := ParseTunnel(tun)
	return err
}
