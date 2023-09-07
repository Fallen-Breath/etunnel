package tunnel

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

func doClose(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Debugf("Close error %v", err)
	}
}

func relayTcp(left, right net.Conn) {
	var wg sync.WaitGroup

	singleForward := func(name string, source net.Conn, target net.Conn) {
		defer wg.Done()

		log.Infof("Forward start %s", name)
		_, err := io.Copy(target, source)
		log.Infof("Forward end %s %v", name, err)

		if err == nil { // source EOF
			if tcpConn, ok := target.(*net.TCPConn); ok {
				log.Infof("Close %s target", name)
				_ = tcpConn.CloseRead()
				_ = tcpConn.CloseWrite()
			}
		}
		if err == io.EOF { // target EOF
			if tcpConn, ok := source.(*net.TCPConn); ok {
				log.Infof("Close %s source", name)
				_ = tcpConn.CloseRead()
				_ = tcpConn.CloseWrite()
			}
		}
		//_ = target.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	}

	wg.Add(2)
	go singleForward("L->R", left, right)
	go singleForward("L<-R", right, left)
	wg.Wait()
}
