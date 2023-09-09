package config

import (
	"github.com/Fallen-Breath/etunnel/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseTunnel(t *testing.T) {
	defaultId := "DEFAULT"
	cases := []struct {
		tun      string
		id       string
		protocol string
		address  string
		errKey   string
	}{
		// with id
		{"foo:tcp://127.0.0.1:8888", "foo", proto.Tcp, "127.0.0.1:8888", ""},
		{"foo:udp://127.0.0.1:8888", "foo", proto.Udp, "127.0.0.1:8888", ""},

		// test all tunnel protocols
		{"tcp://127.0.0.1:8888", defaultId, proto.Tcp, "127.0.0.1:8888", ""},
		{"udp://127.0.0.1:8888", defaultId, proto.Udp, "127.0.0.1:8888", ""},
		{"unix:///some/path.sock", defaultId, proto.Unix, "/some/path.sock", ""},
		{"unixgram:///some/path.sock", defaultId, proto.UnixGram, "/some/path.sock", ""},

		// addresses
		{"udp://:8888", defaultId, proto.Udp, ":8888", ""},
		{"udp://[2001:0000:130F:0000:0000:09C0:876A:130B]:8888", defaultId, proto.Udp, "[2001:0000:130F:0000:0000:09C0:876A:130B]:8888", ""},

		// test domain name (they're not validated yet
		{"tcp://whateverRandomString:80", defaultId, proto.Tcp, "whateverRandomString:80", ""},
		{"udp://random.domain.:443", defaultId, proto.Udp, "random.domain.:443", ""},

		// non tunnel protocols
		{"kcp://127.0.0.1:8888", "", "", "", "invalid tunnel protocol"},
		{"quic://127.0.0.1:8888", "", "", "", "invalid tunnel protocol"},
		{"ftp://127.0.0.1:8888", "", "", "", "invalid tunnel protocol"},

		// bad formats
		{"foo:bar:tcp://127.0.0.1:8888", "", "", "", "invalid id separation"},
		{"foo:bar://tcp://127.0.0.1:8888", "", "", "", "invalid protocol separation"},
		{"foo!!bar:tcp://127.0.0.1:8888", "", "", "", "invalid tunnel id"},
		{"tcp://127.0.0.1:abc", "", "", "", "invalid address"},
		{"tcp://127.0.0.1:65536", "", "", "", "invalid address"},
		{"tcp://127.0.0.1", "", "", "", "invalid address"},
		{"tcp://127.0.0.1::", "", "", "", "invalid address"},
	}
	for _, c := range cases {
		i, p, a, e := ParseTunnel(c.tun, defaultId)

		if e != nil {
			errStr := e.Error()
			require.NotEmptyf(t, c.errKey, "case %+v: unexpected error: %s", c, errStr)
			require.Containsf(t, errStr, c.errKey, "case %+v: error mismatch", c)
		} else {
			require.Emptyf(t, c.errKey, "case %+v: expected no error, but error found", c)

			protocol, ok := proto.GetProtocolMeta(c.protocol)
			require.True(t, ok, "case %+v", c)

			assert.Equalf(t, c.id, i, "case %+v: id mismatch", c)
			assert.Equalf(t, protocol, p, "case %+v: protocol mismatch", c)
			assert.Equalf(t, c.address, a, "case %+v: address mismatch", c)
		}
	}
}
