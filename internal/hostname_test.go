package internal

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name string
		addr *net.TCPAddr
		want string
	}{
		{"hardcoded listening address", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}, "127.0.0.1:8080"},
		{"ipv6 unspecified", &net.TCPAddr{IP: net.IPv6unspecified, Port: 8888}, "127.0.0.1:8888"},
		{"ipv4 unspecified", &net.TCPAddr{IP: net.IPv4zero, Port: 8888}, "127.0.0.1:8888"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeAddress(tt.addr))
		})
	}
}

func TestUnspecifiedIP(t *testing.T) {
	t.Log(net.ParseIP("[::]").IsUnspecified())
}
