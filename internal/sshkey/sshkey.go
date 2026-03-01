// Package sshkey manages SSH keys for organizations.
package sshkey

import (
	"time"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

// SSHKey is an SSH key belonging to an organization.
type SSHKey struct {
	ID           resource.TfeID   `db:"ssh_key_id"`
	CreatedAt    time.Time        `db:"created_at"`
	UpdatedAt    time.Time        `db:"updated_at"`
	Name         string           `db:"name"`
	Organization organization.Name `db:"organization_name"`
	PrivateKey   string           `db:"private_key"`
}

// CreateOptions are the options for creating a new SSH key.
type CreateOptions struct {
	Organization *organization.Name
	Name         *string
	PrivateKey   *string
}

// UpdateOptions are the options for updating an SSH key.
type UpdateOptions struct {
	Name       *string
	PrivateKey *string
}
