package repo

import (
	"path"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
)

type (
	factory struct {
		internal.HostnameService
	}

	newHookOptions struct {
		id            *uuid.UUID
		vcsProviderID string
		secret        *string
		identifier    string
		cloud         cloud.Kind
		cloudID       *string // cloud's webhook id
	}
)

// fromRow creates a hook from a database row
func (f factory) fromRow(row hookRow) (*Hook, error) {
	opts := newHookOptions{
		id:            internal.UUID(row.WebhookID.Bytes),
		vcsProviderID: row.VCSProviderID.String,
		secret:        internal.String(row.Secret.String),
		identifier:    row.Identifier.String,
		cloud:         cloud.Kind(row.Cloud.String),
	}
	if row.VCSID.Status == pgtype.Present {
		opts.cloudID = internal.String(row.VCSID.String)
	}
	return f.newHook(opts)
}

func (f factory) newHook(opts newHookOptions) (*Hook, error) {
	hook := Hook{
		identifier:    opts.identifier,
		cloud:         opts.cloud,
		cloudID:       opts.cloudID,
		vcsProviderID: opts.vcsProviderID,
	}

	if opts.id != nil {
		hook.id = *opts.id
	} else {
		hook.id = uuid.New()
	}

	if opts.secret != nil {
		hook.secret = *opts.secret
	} else {
		secret, err := internal.GenerateToken()
		if err != nil {
			return nil, err
		}
		hook.secret = secret
	}

	hook.endpoint = f.URL(path.Join(HandlerPrefix, hook.id.String()))

	return &hook, nil
}
