package tunnel

import (
	"container/list"
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/Fallen-Breath/etunnel/internal/proto/header"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"sync/atomic"
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
	scPool := &serverConnectionPool{
		list:   list.New(),
		logger: t.logger,
	}
	var cid atomic.Uint64
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

		// TODO: use cliAddr as the key of the cached connection
		// TODO: check if goroutine is necessary
		go t.handlePacketConnection(buf[:n], client, scPool, t.logger.WithField("cid", cid.Add(1)))
	}
}

func (t *tunnelHandler) handlePacketConnection(packet []byte, client *udpClient, scPool *serverConnectionPool, logger *log.Entry) {
	holder, ok := scPool.Pop()
	if !ok {
		svrConn, err := t.connectToServer(&header.ReqHead{
			Kind: proto.KindPacket,
			Id:   t.tunnel.Id,
		})
		if err != nil {
			logger.Errorf("Failed to connect to the server: %v", err)
			return
		}

		holder = &connectionHolder{conn: svrConn, scId: scPool.scCount.Add(1)}
		logger.Debugf("Created server connection #%d", holder.scId)
	}

	logger.Infof("Forward start: %s --(tunnel)-> dest", client.addr)
	send, recv, err := t.relayPacketConnection(packet, client, holder.conn, logger)
	if err != nil {
		logger.Errorf("Relay packet connection failed: %v", err)
	}
	logger.Infof("Forward end: %s --(tunnel)-> dest%s", client.addr, makeFlowTail(send, recv))

	//scPool.Push(holder)
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
		//logger.Debugf("%s", buf)

		if err := client.Send(buf); err != nil {
			return send, recv, fmt.Errorf("send packet to client failed: %v", err)
		}
	}
}

const serverConnectionTTL = 1 * time.Second

type connectionHolder struct {
	conn  conn.StreamConn
	scId  uint64
	timer *time.Timer
	el    *list.Element
}

type serverConnectionPool struct {
	list    *list.List
	mutex   sync.Mutex
	logger  *log.Entry
	scCount atomic.Uint64
}

func (p *serverConnectionPool) Push(holder *connectionHolder) {
	p.Lock()
	defer p.Unlock()

	holder.timer = time.AfterFunc(serverConnectionTTL, func() {
		p.Lock()
		defer p.Unlock()
		if holder.el != nil {
			p.list.Remove(holder.el)
			_ = holder.conn.Close()
			p.logger.Debugf("Closed server connection #%d", holder.scId)
			holder.el = nil
		}
	})
	holder.el = p.list.PushBack(holder)
}

func (p *serverConnectionPool) Pop() (*connectionHolder, bool) {
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

func (p *serverConnectionPool) Lock() {
	p.mutex.Lock()
}

func (p *serverConnectionPool) Unlock() {
	p.mutex.Unlock()
}
