package otf

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostnameService(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		listen   string
		want     string
	}{
		{
			name:     "hardcoded hostname",
			hostname: "hardcoded.com",
			want:     "hardcoded.com",
		},
		{
			name:   "hardcoded listening address",
			listen: "127.0.0.1:8080",
			want:   "127.0.0.1:8080",
		},
		{
			name:   "only port provided",
			listen: ":8888",
			want:   "127.0.0.1:8888",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &HostnameService{}
			err := svc.SetHostname(tt.hostname, tt.listen)
			require.NoError(t, err)
			assert.Equal(t, tt.want, svc.hostname)
		})
	}
}

func TestUnspecifiedIP(t *testing.T) {
	t.Log(net.ParseIP("[::]").IsUnspecified())
}
