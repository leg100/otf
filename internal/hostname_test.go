package internal

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostnameService(t *testing.T) {
	tests := []struct {
		name                string
		svc                 *HostnameService
		wantHostname        string
		wantWebhookHostname string
	}{
		{
			"default",
			NewHostnameService("localhost:8080"),
			"localhost:8080",
			"localhost:8080",
		},
		{
			"set hostname",
			func() *HostnameService {
				svc := NewHostnameService("")
				svc.SetHostname("otf.local")
				return svc
			}(),
			"otf.local",
			"otf.local",
		},
		{
			"set webhook hostname",
			func() *HostnameService {
				svc := NewHostnameService("localhost:8080")
				svc.SetWebhookHostname("otf.local")
				return svc
			}(),
			"localhost:8080",
			"otf.local",
		},
		{
			"set both hostnames",
			func() *HostnameService {
				svc := NewHostnameService("localhost:8080")
				svc.SetHostname("otf.local")
				svc.SetWebhookHostname("webhooks.otf.local")
				return svc
			}(),
			"otf.local",
			"webhooks.otf.local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantHostname, tt.svc.Hostname())
			assert.Equal(t, tt.wantWebhookHostname, tt.svc.WebhookHostname())
		})
	}
}

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
