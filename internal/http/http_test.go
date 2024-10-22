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

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    error
	}{
		{
			name:    "valid http address",
			address: "http://localhost:8080",
		},
		{
			name:    "valid https address",
			address: "https://localhost:8080",
		},
		{
			name:    "valid https address with path",
			address: "https://localhost:8080/otf",
		},
		{
			name:    "invalid address missing scheme",
			address: "localhost:8080",
			want:    ErrParseURLMissingScheme,
		},
		{
			name:    "invalid address with non-http(s) scheme",
			address: "ftp://localhost:8080",
			want:    ErrParseURLMissingScheme,
		},
		{
			name:    "invalid address without host",
			address: "http:///otf",
			want:    ErrParseURLMissingHost,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseURL(tt.address)
			assert.ErrorIs(t, err, tt.want)
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
