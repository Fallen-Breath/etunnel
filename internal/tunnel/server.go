package tunnel

import (
	"github.com/shadowsocks/go-shadowsocks2/socks"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

// reference: github.com/shadowsocks/go-shadowsocks2/tcp.go tcpRemote
func (t *Tunnel) runServer(listen string) {
	listener, err := net.Listen("tcp", t.conf.Listen)
	if err != nil {
		log.Errorf("failed to listen on %s: %v", t.conf.Listen, err)
		return
	}

	defer closeWhatever(listener)
	log.Infof("Listening on %s", t.conf.Listen)

	for {
		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("Failed to accept: %v", err)
			continue
		}

		go func() {
			defer closeWhatever(cliConn)

			// TODO: TCP cork support
			shadowedConn := t.cipher.StreamConn(cliConn)

			target, err := socks.ReadAddr(shadowedConn)
			if err != nil {
				log.Errorf("failed to get target address from %v: %v", cliConn.RemoteAddr(), err)
				// drain c to avoid leaking server behavioral features
				// see https://www.ndss-symposium.org/ndss-paper/detecting-probe-resistant-proxies/
				_, err = io.Copy(io.Discard, cliConn)
				if err != nil {
					log.Errorf("discard error: %v", err)
				}
				return
			}

			targetConn, err := net.Dial("tcp", target.String())
			if err != nil {
				log.Errorf("Failed to connect to target %s: %v", target, err)
				return
			}
			defer closeWhatever(targetConn)

			log.Errorf("proxy %s <-> %s", cliConn.RemoteAddr(), target)
			relayTcp(shadowedConn, targetConn)
		}()
	}
}
