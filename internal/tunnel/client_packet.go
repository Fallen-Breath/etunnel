package tunnel

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/Fallen-Breath/etunnel/internal/proto/header"
	log "github.com/sirupsen/logrus"
	"math"
	"net"
	"sync"
	"time"
)

func (t *tunnelHandler) runPacketTunnel() {
	if true {
		t.logger.Errorf("Packet tunnel has not implemented yet")
		return
	}

	listener, err := net.ListenPacket(t.tunnel.Protocol, t.tunnel.Listen)
	if err != nil {
		t.logger.Errorf("Failed to listen on %s: %v", t.tunnel.Listen, err)
		return
	}
	defer func() { _ = listener.Close() }()
	t.logger.Infof("Packet tunnel (%s) start: -> [ %s -> %s ] ->", t.tunnel.Protocol, t.tunnel.Listen, t.serverAddr)
	go func() {
		<-t.stopCh
		_ = listener.Close()
	}()

	buf := make([]byte, math.MaxUint16)
	connPool := &connectionPool{list: list.New()}
	cid := 0
	for {
		n, addr, err := listener.ReadFrom(buf)
		if err != nil {
			t.logger.Errorf("Failed to accept: %v", err)
			continue
		}

		t.logger.Debugf("Accepted packet with size %d from %s", n, addr)

		go t.handlePacketConnection(buf[:n], addr, connPool, log.WithField("cid", cid))
		cid++
	}
}

func (t *tunnelHandler) handlePacketConnection(packet []byte, addr net.Addr, pool *connectionPool, logger *log.Entry) {
	length := len(packet)
	if len(packet) > math.MaxUint16 {
		logger.Errorf("UDP packet too large, %d > %d", length, math.MaxUint16)
	}

	holder, ok := pool.Pop()
	if !ok {
		svrConn, err := t.connectToServer(&header.ReqHead{
			Kind: proto.KindPacket,
			Id:   t.tunnel.Id,
		})
		if err != nil {
			logger.Errorf("Failed to connect to the server: %v", err)
			return
		}

		holder = &connectionHolder{conn: svrConn}
	}

	logger.Infof("Forward start: %s --(tunnel)-> dest", addr)

	buf := binary.BigEndian.AppendUint16(nil, uint16(length))
	if _, err := holder.conn.Write(buf); err != nil {
		_ = holder.conn.Close()
		logger.Errorf("Failed to write packet length to the server: %v", err)
		return
	}
	if _, err := holder.conn.Write(packet); err != nil {
		_ = holder.conn.Close()
		logger.Errorf("Failed to write packet to the server: %v", err)
		return
	}

	flow := ""
	if log.GetLevel() >= log.DebugLevel {
		flow = fmt.Sprintf(" (send %d)", length)
	}
	logger.Infof("Forward end: %s --(tunnel)-> dest%s", addr, flow)

	pool.Push(holder)
}

type connectionHolder struct {
	conn  conn.StreamConn
	timer *time.Timer
	el    *list.Element
}

type connectionPool struct {
	list  *list.List
	mutex sync.Mutex
}

func (p *connectionPool) Push(h *connectionHolder) {
	p.Lock()
	defer p.Unlock()

	h.timer = time.AfterFunc(500*time.Millisecond, func() {
		p.Lock()
		defer p.Unlock()
		if h.el != nil {
			p.list.Remove(h.el)
			_ = h.conn.Close()
			h.el = nil
		}
	})
	h.el = p.list.PushBack(h)
}

func (p *connectionPool) Pop() (*connectionHolder, bool) {
	p.Lock()
	defer p.Unlock()

	el := p.list.Back()
	if el == nil {
		return nil, false
	}
	p.list.Remove(el)
	holder := el.Value.(*connectionHolder)
	holder.timer.Stop()
	holder.el = nil
	return holder, true
}

func (p *connectionPool) Lock() {
	p.mutex.Lock()
}

func (p *connectionPool) Unlock() {
	p.mutex.Unlock()
}
