// Package sshkey manages SSH keys for organizations.
package sshkey

import (
	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

// SSHKey is an SSH key belonging to an organization.
type SSHKey struct {
	ID           resource.TfeID    `db:"ssh_key_id"`
	Name         string            `db:"name"`
	Organization organization.Name `db:"organization_name"`
}

// CreateOptions are the options for creating a new SSH key.
type CreateOptions struct {
	Organization organization.Name `schema:"organization_name"`
	Name         string
	PrivateKey   string `schema:"private-key"`
}

// UpdateOptions are the options for updating an SSH key.
type UpdateOptions struct {
	Name *string
}

func New(opts CreateOptions) (*SSHKey, []byte, error) {
	if opts.Organization.String() == "" {
		return nil, nil, &internal.ErrMissingParameter{Parameter: "organization"}
	}
	if opts.Name == "" {
		return nil, nil, &internal.ErrMissingParameter{Parameter: "name"}
	}
	if len(opts.PrivateKey) == 0 {
		return nil, nil, &internal.ErrMissingParameter{Parameter: "private_key"}
	}
	key := &SSHKey{
		ID:           resource.NewTfeID(resource.SSHKeyKind),
		Name:         opts.Name,
		Organization: opts.Organization,
	}
	return key, []byte(opts.PrivateKey), nil
}

// LogValue implements slog.LogValuer.
func (key *SSHKey) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", key.ID.String()),
		slog.Any("organization", key.Organization),
		slog.String("name", key.Name),
	)
}
