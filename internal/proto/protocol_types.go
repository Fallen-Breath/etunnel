package proto

type ProtocolMeta struct {
	Name        string
	Tunnel      bool // valid tunnel protocol?
	Communicate bool // valid communicate protocol?
}

const (
	Tcp  = "tcp"
	Udp  = "udp"
	Kcp  = "kcp"
	Quic = "quic"
	//Unix     = "unix"
	//UnixGram = "unixgram"
)

var (
	TcpMeta  = ProtocolMeta{Tcp, true, true}
	UdpMeta  = ProtocolMeta{Udp, true, false}
	KcpMeta  = ProtocolMeta{Kcp, false, true}
	QuicMeta = ProtocolMeta{Quic, false, true}
	//UnixMeta     = ProtocolMeta{"unix", true, false}
	//UnixGramMeta = ProtocolMeta{"unixgram", true, false}
)

var protocolMap = map[string]ProtocolMeta{}

func init() {
	for _, p := range []ProtocolMeta{TcpMeta, UdpMeta, KcpMeta, QuicMeta} {
		protocolMap[p.Name] = p
	}
}

func GetProtocolMeta(s string) (*ProtocolMeta, bool) {
	if v, ok := protocolMap[s]; ok {
		return &v, true
	} else {
		return nil, false
	}
}

func CheckProtocol(s string) bool {
	_, ok := GetProtocolMeta(s)
	return ok
}

func ListTunnelProtocols() []ProtocolMeta {
	var l []ProtocolMeta
	for _, p := range protocolMap {
		if p.Tunnel {
			l = append(l, p)
		}
	}
	return l
}

func ListCommunicateProtocols() []ProtocolMeta {
	var l []ProtocolMeta
	for _, p := range protocolMap {
		if p.Communicate {
			l = append(l, p)
		}
	}
	return l
}
