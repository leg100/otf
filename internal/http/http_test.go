package http

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeHostname(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{
			name:    "no scheme",
			address: "localhost:8080",
			want:    "localhost:8080",
		},
		{
			name:    "has scheme",
			address: "https://localhost:8080",
			want:    "localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := SanitizeHostname(tt.address)
			require.NoError(t, err)
			assert.Equal(t, tt.want, address)
		})
	}
}

func TestSanitizeAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    string
	}{
		{
			name:    "add scheme",
			address: "localhost:8080",
			want:    "https://localhost:8080",
		},
		{
			name:    "no change",
			address: "https://localhost:8080",
			want:    "https://localhost:8080",
		},
		{
			name:    "fix scheme",
			address: "http://localhost:8080",
			want:    "https://localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeAddress(tt.address)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name      string
		forwarded string
		remote    string
		want      string
	}{
		{
			name:   "no forwarded header",
			remote: "1.2.3.4:1234",
			want:   "1.2.3.4",
		},
		{
			name:      "forwarded header",
			remote:    "1.2.3.4:1234",
			forwarded: "5.6.7.8",
			want:      "5.6.7.8",
		},
		{
			name:      "forwarded header with multiple ips",
			remote:    "1.2.3.4:1234",
			forwarded: "5.6.7.8, 9.0.1.2",
			want:      "5.6.7.8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("", "/", nil)
			r.Header.Add("X-Forwarded-For", tt.forwarded)
			r.RemoteAddr = tt.remote
			got, err := GetClientIP(r)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
