package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/protocol/header"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"time"
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

			// TODO: TCP cork support

			originConn := cliConn.(*net.TCPConn)
			cliConn := conn.NewEncryptedStreamConn(originConn, t.cipher)

			log.Infof("Read header")
			var head header.Header
			if err := head.UnmarshalFrom(cliConn); err != nil {
				log.Errorf("Failed to read header: %v", err)
				// drain originConn to avoid leaking server behavioral features
				// see https://www.ndss-symposium.org/ndss-paper/detecting-probe-resistant-proxies/
				_ = originConn.SetDeadline(time.Now().Add(30 * time.Second)) // TODO: find a nice way to close connection
				_, err = io.Copy(io.Discard, originConn)
				if err != nil {
					log.Errorf("Discard error: %v", err)
				}
				return
			}

			target := head.Target
			log.Infof("Dial target %s://%s start", head.Protocol, head.Target)
			targetConn, err := net.Dial(head.Protocol, target)
			if err != nil {
				log.Errorf("Failed to connect to target %s://%s: %v", head.Protocol, head.Target, err)
				return
			}
			defer doClose(targetConn)

			log.Infof("Forward start %s <-> %s", cliConn.RemoteAddr(), target)
			relayConnection(cliConn, targetConn)
			log.Infof("Forward end %s <-> %s", cliConn.RemoteAddr(), target)
		}()
	}
}
