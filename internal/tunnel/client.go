package tunnel

import (
	"fmt"
	"github.com/Fallen-Breath/etunnel/internal/config"
	"github.com/Fallen-Breath/etunnel/internal/conn"
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/Fallen-Breath/etunnel/internal/proto/header"
	sscore "github.com/shadowsocks/go-shadowsocks2/core"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type Client struct {
	*base

	wg       sync.WaitGroup
	handlers map[string]*tunnelHandler // tunnel definition -> handler
}

func newClient(base *base) (ITunnel, error) {
	return &Client{
		base:     base,
		handlers: make(map[string]*tunnelHandler),
	}, nil
}

func (t *Client) start() {
	t.reloadClient()
	t.wg.Wait()
}

func (t *Client) reloadClient() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	expectedTunnels := make(map[string]bool)
	for _, tun := range t.conf.Tunnels {
		expectedTunnels[tun.GetDefinition()] = true
		if _, ok := t.handlers[tun.GetDefinition()]; ok {
			// already exists, skipped
			continue
		}

		handler, err := newTunnelHandler(t.conf, *tun, t.cipher)
		if err != nil {
			log.Errorf("invalid tunnel %s", tun)
			continue
		}

		log.Infof("Creating tunnel %s on %s://%s", tun.Id, tun.Protocol, tun.Listen)
		t.wg.Add(1)
		t.handlers[tun.GetDefinition()] = handler
		go func() {
			defer t.wg.Done()
			handler.Start()
		}()
	}

	var tunnelsToRemove []string
	for tun := range t.handlers {
		if _, ok := expectedTunnels[tun]; !ok {
			tunnelsToRemove = append(tunnelsToRemove, tun)
		}
	}

	for _, tun := range tunnelsToRemove {
		if handler, ok := t.handlers[tun]; ok {
			log.Infof("Removing tunnel %s", tun)
			handler.Stop()
			delete(t.handlers, tun)
		}
	}
}

type tunnelHandler struct {
	cipher sscore.Cipher
	stopCh chan int
	tunnel config.Tunnel
	logger *log.Entry

	// configs
	serverAddr string
	corking    bool
}

func (t *tunnelHandler) GetDefinition() string {
	return t.tunnel.GetDefinition()
}

func newTunnelHandler(conf *config.Config, tun config.Tunnel, cipher sscore.Cipher) (*tunnelHandler, error) {
	return &tunnelHandler{
		cipher: cipher,
		stopCh: make(chan int, 1),
		tunnel: tun,
		logger: log.WithField("tid", tun.Id),

		serverAddr: conf.Server,
		corking:    conf.Cork,
	}, nil
}

func (t *tunnelHandler) Start() {
	switch t.tunnel.Protocol {
	case proto.Tcp:
		t.runStreamTunnel()
	case proto.Udp:
		t.runPacketTunnel()
	default:
		t.logger.Errorf("Unsupported protocol %s", t.tunnel.Protocol)
	}
}

func (t *tunnelHandler) Stop() {
	t.stopCh <- 0
}

func (t *tunnelHandler) connectToServer(head *header.ReqHead) (conn.StreamConn, error) {
	sc, err := net.Dial("tcp", t.serverAddr)
	if err != nil {
		return nil, fmt.Errorf("dial %s failed: %v", t.serverAddr, err)
	}
	svrConn := sc.(conn.StreamConn)
	defer func() {
		if err != nil {
			_ = svrConn.Close()
		}
	}()

	if t.corking {
		svrConn = conn.NewTimedCorkConn(svrConn, 10*time.Millisecond, 1280)
	}
	svrConn = conn.NewEncryptedStreamConn(svrConn, t.cipher)

	if err = head.MarshalTo(svrConn); err != nil {
		return nil, fmt.Errorf("send request header failed: %v", err)
	}

	var headRsp header.RspHead
	if err := headRsp.UnmarshalFrom(svrConn); err != nil {
		return nil, fmt.Errorf("receive response header failed: %v", err)
	}
	switch headRsp.Code {
	case header.CodeOk:
		// ok
	case header.CodeBadId:
		return nil, fmt.Errorf("server rejects the tunnel, bad id %s", t.tunnel.Id)
	case header.CodeBadKind:
		return nil, fmt.Errorf("server rejects the tunnel, bad protocol %s", t.tunnel.Protocol)
	default:
		return nil, fmt.Errorf("unknown response code %d", headRsp.Code)
	}

	return svrConn, nil
}
