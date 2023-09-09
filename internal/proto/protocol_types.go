package proto

type ProtocolMeta struct {
	Name        string
	Communicate bool        // valid communicate protocol?
	Tunnel      bool        // valid tunnel protocol?
	Kind        PayloadKind // payload kind it provides if it can be used in tunnel
}

const (
	Tcp      = "tcp"
	Udp      = "udp"
	Kcp      = "kcp"
	Quic     = "quic"
	Unix     = "unix"
	UnixGram = "unixgram"
)

var (
	TcpMeta      = ProtocolMeta{Tcp, true, true, KindStream}
	UdpMeta      = ProtocolMeta{Udp, false, true, KindPacket}
	KcpMeta      = ProtocolMeta{Kcp, true, false, nil}
	QuicMeta     = ProtocolMeta{Quic, true, false, nil}
	UnixMeta     = ProtocolMeta{Unix, false, true, KindStream}
	UnixGramMeta = ProtocolMeta{UnixGram, false, true, KindPacket}
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
