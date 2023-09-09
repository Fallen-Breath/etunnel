package tunnel

import (
	"encoding/binary"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

func makeFlowTail(send, recv int64) string {
	if log.GetLevel() >= log.DebugLevel {
		return fmt.Sprintf(" (send %d, recv %d)", send, recv)
	}
	return ""
}

func relayStreamConnection(left, right net.Conn, logger *log.Entry) (l2r, r2l int64, err error) {
	var wg sync.WaitGroup

	forward := func(source net.Conn, target net.Conn, name string, count *int64) {
		defer wg.Done()

		logger.Debugf("Forward start %s", name)
		*count, err = io.Copy(target, source)
		logger.Debugf("Forward end %s %v", name, err)

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

func receivePacket(tunnelConn conn.StreamConn) ([]byte, error) {
	length, err := binary.ReadUvarint(newByteReader(tunnelConn))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(tunnelConn, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func sendPacket(tunnelConn conn.StreamConn, packet []byte) error {
	buf := binary.AppendUvarint(nil, uint64(len(packet)))
	if _, err := tunnelConn.Write(buf); err != nil {
		return err
	}

	if _, err := tunnelConn.Write(packet); err != nil {
		return err
	}

	return nil
}

type byteReaderFromReader struct {
	r io.Reader
}

func (br *byteReaderFromReader) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := br.r.Read(buf[:])
	return buf[0], err
}

func newByteReader(r io.Reader) io.ByteReader {
	return &byteReaderFromReader{r: r}
}
