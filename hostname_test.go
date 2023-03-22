package otf

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		listen   *net.TCPAddr
		want     string
	}{
		{
			name:     "hardcoded hostname",
			hostname: "hardcoded.com",
			want:     "hardcoded.com",
		},
		{
			name:   "hardcoded listening address",
			listen: &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080},
			want:   "127.0.0.1:8080",
		},
		{
			name:   "ipv6 unspecified",
			listen: &net.TCPAddr{IP: net.IPv6unspecified, Port: 8888},
			want:   "127.0.0.1:8888",
		},
		{
			name:   "ipv4 unspecified",
			listen: &net.TCPAddr{IP: net.IPv4zero, Port: 8888},
			want:   "127.0.0.1:8888",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setHostname(tt.hostname, tt.listen)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnspecifiedIP(t *testing.T) {
	t.Log(net.ParseIP("[::]").IsUnspecified())
}
