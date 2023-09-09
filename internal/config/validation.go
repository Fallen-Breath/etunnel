package config

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/utils"
	"net"
)

func ValidateCrypt(crypt string) error {
	if utils.Contains(Crypts, crypt) {
		return nil
	}
	return fmt.Errorf("unknown crypt %s", crypt)
}

func ValidateAddress(netAddr string) error {
	_, err := net.ResolveTCPAddr("tcp", netAddr)
	return err
}
