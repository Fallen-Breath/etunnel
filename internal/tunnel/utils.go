package tunnel

import (
	"github.com/Fallen-Breath/etunnel/internal/conn"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

func relayConnection(left, right net.Conn) (l2r int64, r2l int64) {
	var wg sync.WaitGroup

	forward := func(source net.Conn, target net.Conn, name string, count *int64) {
		defer wg.Done()

		log.Debugf("Forward start %s", name)
		n, err := io.Copy(target, source)
		log.Debugf("Forward end %s %v", name, err)
		*count = n

		if err == nil { // source EOF, no more incoming data from source, so stop sending more data to target
			if sc, ok := target.(conn.StreamConn); ok {
				_ = sc.CloseWrite()
			}
		} else if err == io.EOF { // target EOF, cannot send any more data to target, so stop receiving more data from source
			if sc, ok := source.(conn.StreamConn); ok {
				_ = sc.CloseRead()
			}
		} else {
			_ = source.Close()
			_ = target.Close()
		}
	}

	wg.Add(2)
	go forward(left, right, "L->R", &l2r)
	go forward(right, left, "L<-R", &r2l)
	wg.Wait()

	return
}
