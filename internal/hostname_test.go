package internal

import (
	"net"
	"testing"

	"github.com/leg100/otf/internal/logr"
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
			"no explicit hostnames set, use unspecified ipv4 listening address and port",
			NewHostnameService(
				logr.Discard(),
				"",
				"",
				&net.TCPAddr{IP: net.IPv4zero, Port: 1234},
			),
			"127.0.0.1:1234",
			"127.0.0.1:1234",
		},
		{
			"no explicit hostnames set, use unspecified ipv6 listening address and port",
			NewHostnameService(
				logr.Discard(),
				"",
				"",
				&net.TCPAddr{IP: net.IPv6unspecified, Port: 5678},
			),
			"127.0.0.1:5678",
			"127.0.0.1:5678",
		},
		{
			"no explicit hostnames set, use localhost listening address and port",
			NewHostnameService(
				logr.Discard(),
				"",
				"",
				&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9012},
			),
			"127.0.0.1:9012",
			"127.0.0.1:9012",
		},
		{
			"no explicit hostnames set, use non-localhost listening address and port",
			NewHostnameService(
				logr.Discard(),
				"",
				"",
				&net.TCPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 3456},
			),
			"192.168.0.1:3456",
			"192.168.0.1:3456",
		},
		{
			"use explicit hostname and port",
			NewHostnameService(
				logr.Discard(),
				"enterprise.otf.com:7890",
				"",
				&net.TCPAddr{},
			),
			"enterprise.otf.com:7890",
			"enterprise.otf.com:7890",
		},
		{
			"use explicit webhook hostname",
			NewHostnameService(
				logr.Discard(),
				"otf.local",
				"webhooks.otf.local",
				&net.TCPAddr{},
			),
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

func TestHostnameService_LocalURL(t *testing.T) {
	svc := NewHostnameService(
		logr.Discard(),
		"",
		"",
		&net.TCPAddr{Port: 3456},
	)
	got := svc.LocalURL("/foo")
	assert.Equal(t, "https://localhost:3456/foo", got)
}
