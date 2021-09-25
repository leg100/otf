package http

import (
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
			address, err := sanitizeHostname(tt.address)
			require.NoError(t, err)
			assert.Equal(t, tt.want, address)
		})
	}
}

func TestClientSanitizeAddress(t *testing.T) {
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
			got, err := sanitizeAddress(tt.address)
			if assert.NoError(t, err) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
