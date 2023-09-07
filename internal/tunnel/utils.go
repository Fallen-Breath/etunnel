package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/conn"
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

func relayConnection(left, right net.Conn) {
	var wg sync.WaitGroup

	singleForward := func(name string, source net.Conn, target net.Conn) {
		defer wg.Done()

		log.Infof("Forward start %s", name)
		_, err := io.Copy(target, source)
		log.Infof("Forward end %s %v", name, err)

		if err == nil { // source EOF, no more incoming data from source, so stop sending more data to target
			if sc, ok := target.(conn.StreamConn); ok {
				_ = sc.CloseWrite()
			}
		} else if err == io.EOF { // target EOF, cannot send any more data to target, so stop receiving more data from source
			if sc, ok := source.(conn.StreamConn); ok {
				_ = sc.CloseRead()
			}
		} else {

		}
	}

	wg.Add(2)
	go singleForward("L->R", left, right)
	go singleForward("L<-R", right, left)
	wg.Wait()
}
