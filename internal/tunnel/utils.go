package tunnel

import (
	"encoding/binary"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

func relayConnection(left, right net.Conn, logger *log.Entry) (l2r, r2l int64) {
	var wg sync.WaitGroup

	forward := func(source net.Conn, target net.Conn, name string, count *int64) {
		defer wg.Done()

		logger.Debugf("Forward start %s", name)
		n, err := io.Copy(target, source)
		logger.Debugf("Forward end %s %v", name, err)
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

func relayUdpConnection(cliConn conn.StreamConn, targetConn net.Conn, logger *log.Entry) int64 {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(cliConn, buf); err != nil {
		logger.Errorf("Failed to receive udp packet length")
		return 0
	}

	length := binary.BigEndian.Uint16(buf)
	buf = make([]byte, length)
	if _, err := io.ReadFull(cliConn, buf); err != nil {
		logger.Errorf("Failed to receive udp packet body")
		return 0
	}

	if _, err := targetConn.Write(buf); err != nil {
		logger.Errorf("Failed to send udp packet")
		return 0
	}

	return int64(length)
}
