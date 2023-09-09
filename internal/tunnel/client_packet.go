package tunnel

import (
	"container/list"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/Fallen-Breath/etunnel/internal/proto/header"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"net"
	"sync"
	"time"
)

type udpClient struct {
	l    net.PacketConn
	addr net.Addr
}

func (c *udpClient) Send(packet []byte) error {
	_, err := c.l.WriteTo(packet, c.addr)
	return err
}

func (t *tunnelHandler) runPacketTunnel() {
	listener, err := net.ListenPacket(t.tunnel.Protocol, t.tunnel.Listen)
	if err != nil {
		t.logger.Errorf("Failed to listen on %s: %v", t.tunnel.Listen, err)
		return
	}
	defer func() { _ = listener.Close() }()
	t.logger.Infof("Packet tunnel (%s) start: ->[ %s -> %s ]->", t.tunnel.Protocol, t.tunnel.Listen, t.serverAddr)
	go func() {
		<-t.stopCh
		_ = listener.Close()
	}()

	buf := make([]byte, proto.MaxUdpPacketSize)
	connPool := &connectionPool{list: list.New()}
	cid := 0
	for {
		n, cliAddr, err := listener.ReadFrom(buf)
		if err != nil {
			t.logger.Errorf("Failed to accept: %v", err)
			continue
		}

		t.logger.Debugf("Accepted packet with size %d from %s", n, cliAddr)
		client := &udpClient{
			l:    listener,
			addr: cliAddr,
		}

		cid_ := cid
		cid++
		go t.handlePacketConnection(buf[:n], client, connPool, log.WithField("cid", cid_))
	}
}

func (t *tunnelHandler) handlePacketConnection(packet []byte, client *udpClient, pool *connectionPool, logger *log.Entry) {
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

	logger.Infof("Forward start: %s --(tunnel)-> dest", client.addr)
	send, recv, err := t.relayPacketConnection(packet, client, holder.conn, logger)
	if err != nil {
		logger.Errorf("Relay packet connection failed: %v", err)
	}
	logger.Infof("Forward end: %s --(tunnel)-> dest%s", client.addr, makeFlowTail(send, recv))

	pool.Push(holder)
}

func (t *tunnelHandler) relayPacketConnection(packet []byte, client *udpClient, svrConn conn.StreamConn, logger *log.Entry) (send, recv int64, _ error) {
	if err := sendPacket(svrConn, packet); err != nil {
		return 0, 0, fmt.Errorf("send packet to server failed: %v", err)
	}
	send += int64(len(packet))

	for {
		buf, err := receivePacket(svrConn)
		if err == io.EOF {
			logger.Debugf("Connection closed by the server")
			return send, recv, nil
		}
		if err != nil {
			return send, recv, fmt.Errorf("receive packet from server failed: %v", err)
		}
		recv += int64(len(buf))
		logger.Debugf("Received packet with size %d, forwarding", len(buf))

		if err := client.Send(buf); err != nil {
			return send, recv, fmt.Errorf("send packet to client failed: %v", err)
		}
	}
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

	h.timer = time.AfterFunc(5*time.Millisecond, func() {
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
