package proto

type ProtocolMeta struct {
	Name        string
	Communicate bool        // valid communicate protocol?
	Tunnel      bool        // valid tunnel protocol?
	Kind        PayloadKind // payload kind it provides, if it can be used in tunnel
	Addr        AddressType // address type it uses, if it can be used in tunnel
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
	TcpMeta      = ProtocolMeta{Tcp, true, true, KindStream, AddrRegular}
	UdpMeta      = ProtocolMeta{Udp, false, true, KindPacket, AddrRegular}
	UnixMeta     = ProtocolMeta{Unix, false, true, KindStream, AddrFilePath}
	UnixGramMeta = ProtocolMeta{UnixGram, false, true, KindPacket, AddrFilePath}
	KcpMeta      = ProtocolMeta{Kcp, true, false, nil, nil}
	QuicMeta     = ProtocolMeta{Quic, true, false, nil, nil}
)

var protocolMap = map[string]ProtocolMeta{}

func init() {
	for _, p := range []ProtocolMeta{TcpMeta, UdpMeta, KcpMeta, QuicMeta, UnixMeta, UnixGramMeta} {
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

func ListTunnelProtocols() []string {
	var l []string
	for _, p := range protocolMap {
		if p.Tunnel {
			l = append(l, p.Name)
		}
	}
	return l
}

func ListCommunicateProtocols() []string {
	var l []string
	for _, p := range protocolMap {
		if p.Communicate {
			l = append(l, p.Name)
		}
	}
	return l
}
