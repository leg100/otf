package repo

import (
	"encoding/json"
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
	newHookOpts struct {
		id            *uuid.UUID
		vcsProviderID string
		secret        *string
		identifier    string
		cloud         string // cloud name
		cloudID       *string
	}
)

func newFactory(hostnameService internal.HostnameService, cloudService cloud.Service) factory {
	return factory{cloudService, hostnameService}
}

// Unmarshal creates a hook from a json-encoded database row.
func (f factory) UnmarshalRow(data []byte) (any, error) {
	var row hookRow
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, err
	}
	return f.fromRow(row)
}

// fromDB creates a hook from a database row
func (f factory) fromRow(row hookRow) (*hook, error) {
	opts := newHookOpts{
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

func (f factory) newHook(opts newHookOpts) (*hook, error) {
	cloudConfig, err := f.GetCloudConfig(opts.cloud)
	if err != nil {
		return nil, fmt.Errorf("unknown cloud: %s", opts.cloud)
	}

	hook := hook{
		identifier:    opts.identifier,
		vcsProviderID: opts.vcsProviderID,
		cloud:         opts.cloud,
		EventHandler:  cloudConfig.Cloud,
		cloudID:       opts.cloudID,
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
