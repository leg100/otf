package main

import (
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/stretchr/testify/assert"
)

type FakeClientConfig struct{}

func (f FakeClientConfig) NewClient() (Client, error) { return &FakeClient{}, nil }

type FakeClient struct {
	Client
}

func (f FakeClient) Organizations() tfe.Organizations { return &FakeOrganizationsClient{} }

func (f FakeClient) Workspaces() tfe.Workspaces { return &FakeWorkspacesClient{} }

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
			name:    "already has scheme",
			address: "https://localhost:8080",
			want:    "https://localhost:8080",
		},
		{
			name:    "has wrong scheme",
			address: "http://localhost:8080",
			want:    "https://localhost:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := clientConfig{
				Config: tfe.Config{
					Address: tt.address,
				},
			}
			if assert.NoError(t, cfg.sanitizeAddress()) {
				assert.Equal(t, tt.want, cfg.Address)
			}
		})
	}
}
