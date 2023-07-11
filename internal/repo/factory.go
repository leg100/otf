package repo

import (
	"fmt"
	"net/url"
	"path"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
)

type (
	factory struct {
		cloud.Service
		internal.HostnameService
	}

	newHookOptions struct {
		id            *uuid.UUID
		vcsProviderID string
		secret        *string
		identifier    string
		cloud         string  // cloud name
		cloudID       *string // cloud's webhook id
	}
)

// fromRow creates a hook from a database row
func (f factory) fromRow(row hookRow) (*hook, error) {
	opts := newHookOptions{
		id:            internal.UUID(row.WebhookID.Bytes),
		vcsProviderID: row.VCSProviderID.String,
		secret:        internal.String(row.Secret.String),
		identifier:    row.Identifier.String,
		cloud:         row.Cloud.String,
	}
	if row.VCSID.Status == pgtype.Present {
		opts.cloudID = internal.String(row.VCSID.String)
	}
	return f.newHook(opts)
}

func (f factory) newHook(opts newHookOptions) (*hook, error) {
	cloudConfig, err := f.GetCloudConfig(opts.cloud)
	if err != nil {
		return nil, fmt.Errorf("unknown cloud: %s", opts.cloud)
	}

	hook := hook{
		identifier:    opts.identifier,
		cloud:         opts.cloud,
		EventHandler:  cloudConfig.Cloud,
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

	hook.endpoint = (&url.URL{
		Scheme: "https",
		Host:   f.Hostname(),
		Path:   path.Join(handlerPrefix, hook.id.String()),
	}).String()

	return &hook, nil
}
