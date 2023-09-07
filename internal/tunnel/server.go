package tunnel

import (
	"github.com/shadowsocks/go-shadowsocks2/socks"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

// reference: github.com/shadowsocks/go-shadowsocks2/tcp.go tcpRemote
func (t *Tunnel) runServer() {
	listener, err := net.Listen("tcp", t.conf.Listen)
	if err != nil {
		log.Errorf("failed to listen on %s: %v", t.conf.Listen, err)
		return
	}

	var once sync.Once
	defer once.Do(func() { _ = listener.Close() })

	log.Infof("Listening on %s", t.conf.Listen)
	go func() {
		<-t.stopCh
		log.Infof("Closing connection listener")
		once.Do(func() { _ = listener.Close() })
	}()

	for {
		cliConn, err := listener.Accept()
		if t.stopped.Load() {
			_ = cliConn.Close()
			break
		}
		if err != nil {
			log.Errorf("Failed to accept: %v", err)
			continue
		}
		log.Infof("Accepted connection from %s", cliConn.RemoteAddr())

		go func() {
			defer doClose(cliConn)
			log.Infof("Read header")

			// TODO: TCP cork support
			originConn := cliConn
			cliConn := t.cipher.StreamConn(cliConn)

			if err = readMagic(cliConn); err != nil {
				log.Errorf("Failed to read magic: %v", err)
				return
			}
			target, err := socks.ReadAddr(cliConn)
			if err != nil {
				log.Errorf("failed to get target address from %v: %v", originConn.RemoteAddr(), err)
				// drain c to avoid leaking server behavioral features
				// see https://www.ndss-symposium.org/ndss-paper/detecting-probe-resistant-proxies/
				_, err = io.Copy(io.Discard, originConn)
				if err != nil {
					log.Errorf("discard error: %v", err)
				}
				return
			}

			log.Infof("Dial target %s start", target.String())
			targetConn, err := net.Dial("tcp", target.String())
			if err != nil {
				log.Errorf("Failed to connect to target %s: %v", target, err)
				return
			}
			defer doClose(targetConn)
			log.Infof("Dial target %s done", target.String())

			log.Infof("TCP forward start %s <-> %s", cliConn.RemoteAddr(), target)
			relayTcp(cliConn, targetConn)
			log.Infof("TCP forward end %s <-> %s", cliConn.RemoteAddr(), target)
		}()
	}
}
