package sshkey

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// TFESSHKey represents an SSH key in the TFE API.
type TFESSHKey struct {
	ID        resource.TfeID `jsonapi:"primary,ssh-keys"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created-at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated-at"`
	Name      string         `jsonapi:"attribute" json:"name"`
	// Value contains the private key. Populated in GET responses so runners can fetch it.
	Value string `jsonapi:"attribute" json:"value,omitempty"`
}

// TFESSHKeyCreateOptions are the options for creating a new SSH key via the TFE API.
type TFESSHKeyCreateOptions struct {
	// Type is used by JSON:API to set the resource type.
	Type string `jsonapi:"primary,ssh-keys"`

	// The name of the SSH key.
	Name *string `jsonapi:"attribute" json:"name"`

	// The private key value (PEM-encoded). Write-only.
	Value *string `jsonapi:"attribute" json:"value"`
}

// TFESSHKeyUpdateOptions are the options for updating an SSH key via the TFE API.
type TFESSHKeyUpdateOptions struct {
	// Type is used by JSON:API to set the resource type.
	Type string `jsonapi:"primary,ssh-keys"`

	// The name of the SSH key.
	Name *string `jsonapi:"attribute" json:"name,omitempty"`

	// The private key value (PEM-encoded). Write-only.
	Value *string `jsonapi:"attribute" json:"value,omitempty"`
}
