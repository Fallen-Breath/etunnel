package tunnel

import (
	"io"
	"net"
	"sync"
	"time"
)

func closeWhatever(c io.Closer) {
	_ = c.Close()
}

func relayTcp(left, right net.Conn) {
	var wg sync.WaitGroup

	singleForward := func(source net.Conn, target net.Conn) {
		defer wg.Done()
		_, _ = io.Copy(target, source)
		_ = target.SetReadDeadline(time.Now().Add(5 * time.Second)) // TODO: figure out why
	}

	wg.Add(2)
	go singleForward(left, right)
	go singleForward(right, left)
	wg.Wait()
}
