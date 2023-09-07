package conn

import (
	"bufio"
	"sync"
	"time"
)

type corkedConn struct {
	StreamConn
	writeBuf *bufio.Writer
	enabled  bool
	delay    time.Duration
	err      error
	lock     sync.Mutex
	once     sync.Once
}

func NewTimedCorkConn(conn StreamConn, d time.Duration, bufSize int) StreamConn {
	return &corkedConn{
		StreamConn: conn,
		writeBuf:   bufio.NewWriterSize(conn, bufSize),
		enabled:    true,
		delay:      d,
	}
}

func (w *corkedConn) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if w.err != nil {
		return 0, w.err
	}
	if w.enabled {
		w.once.Do(func() {
			time.AfterFunc(w.delay, func() {
				w.lock.Lock()
				defer w.lock.Unlock()
				w.enabled = false
				w.err = w.writeBuf.Flush()
			})
		})
		return w.writeBuf.Write(p)
	}
	return w.StreamConn.Write(p)
}
