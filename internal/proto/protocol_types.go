package proto

const (
	Tcp      = "tcp"
	Udp      = "udp"
	Unix     = "unix"
	UnixGram = "unixgram"
)

var Protocols = []string{
	Tcp,
	Udp,
	Unix,
	UnixGram,
}
