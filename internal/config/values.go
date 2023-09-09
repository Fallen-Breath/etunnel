package config

import (
	"errors"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/Fallen-Breath/etunnel/internal/utils"
	"net"
	"regexp"
	"strconv"
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

var regexId = regexp.MustCompile("^[a-zA-Z0-9_-]{1,255}$")

// ParseTunnel split an tunnel definition into id + protocol + address
// Format: [id:]protocol://address
func ParseTunnel(tun string, defaultId string) (id string, protocol *proto.ProtocolMeta, address string, err error) {
	var protoStr string
	t := strings.Split(tun, "://")

	if len(t) == 2 { // has protocol
		var idProtocol string
		idProtocol, address = t[0], t[1]

		t = strings.Split(idProtocol, ":")
		if len(t) == 1 {
			id, protoStr = defaultId, t[0]
		} else if len(t) == 2 {
			id, protoStr = t[0], t[1]
		} else {
			return "", nil, "", errors.New("invalid id separation, check your heading 'id:'")
		}

	} else {
		return "", nil, "", errors.New("invalid protocol separation, check your '://' divider")
	}

	if !regexId.MatchString(id) {
		return "", nil, "", fmt.Errorf("invalid tunnel id %s", id)
	}

	protocol, ok := proto.GetProtocolMeta(protoStr)
	if !ok || !protocol.Tunnel {
		return "", nil, "", fmt.Errorf("invalid tunnel protocol %s", t[0])
	}

	if protocol.Addr == proto.AddrRegular {
		if err = ValidateAddress(address); err != nil {
			return "", nil, "", fmt.Errorf("invalid address %s: %v", address, err)
		}
	}

	err = nil
	return
}

func ValidateCrypt(crypt string) error {
	if utils.Contains(Crypts, crypt) {
		return nil
	}
	return fmt.Errorf("unknown crypt %s", crypt)
}

// ValidateAddress
// Notes: The host part is not validated
func ValidateAddress(netAddr string) error {
	_, port, err := net.SplitHostPort(netAddr)
	if err != nil {
		return err
	}
	if portNum, err := strconv.Atoi(port); err != nil || portNum > 65535 || portNum < 1 {
		return errors.New("invalid port")
	}

	return nil
}
