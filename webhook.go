package otf

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

// WebhookEvents are those events webhooks should subscribe to.
var WebhookEvents = []VCSEventType{
	VCSPushEventType,
	VCSPullEventType,
}

type Webhook struct {
	WebhookID  uuid.UUID // otf's ID
	VCSID      string    // vcs provider's webhook ID
	Secret     string    // secret token
	Identifier string    // identifier is <repo_owner>/<repo_name>

	cloudConfig CloudConfig // provides handler for webhook's events
}

func (h *Webhook) ID() string        { return h.WebhookID.String() }
func (h *Webhook) Owner() string     { return strings.Split(h.Identifier, "/")[0] }
func (h *Webhook) Repo() string      { return strings.Split(h.Identifier, "/")[1] }
func (h *Webhook) CloudName() string { return h.cloudConfig.Name }

func (h *Webhook) HandleEvent(w http.ResponseWriter, r *http.Request) VCSEvent {
	return h.cloudConfig.HandleEvent(w, r, HandleEventOptions{
		WebhookID: h.WebhookID,
		Secret:    h.Secret,
	})
}

type WebhookStore interface {
	// SyncWebhook ensures webhook configuration is present and
	// equal in both the store and the VCS provider. The idUpdater is called
	// after the webhook is created in the store (or retrieved if it already
	// exists) and it should be used to ensure the webhook exists and is up to
	// date in the VCS provider before returning the ID the provider uses to
	// identify the webhook. SyncWebhook will then update the store with the ID if
	// it differs from its present value in the store.
	SyncWebhook(ctx context.Context, opts SyncWebhookOptions) (*Webhook, error)
	// GetWebhook retrieves a webhook by its ID
	GetWebhook(ctx context.Context, id uuid.UUID) (*Webhook, error)
	// DeleteWebhook deletes the webhook from the store.
	DeleteWebhook(ctx context.Context, id uuid.UUID) error
}

type SyncWebhookOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud that the webhook belongs to

	CreateWebhookFunc func(context.Context, WebhookCreatorOptions) (*Webhook, error)
	UpdateWebhookFunc func(context.Context, WebhookUpdaterOptions) (string, error)
}

type WebhookCreatorOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud providing webhook
}

type WebhookCreator struct {
	VCSProviderService // for creating webhook on vcs provider
	CloudService       // for retrieving event handler
	HostnameService    // for retrieving system hostname
}

func (wc *WebhookCreator) Create(ctx context.Context, opts WebhookCreatorOptions) (*Webhook, error) {
	webhookID := uuid.New()
	secret, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	// lookup event cloudConfig using cloud name
	cloudConfig, err := wc.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}

	// create webhook on vcs provider
	id, err := wc.CreateWebhook(ctx, opts.ProviderID, CreateWebhookOptions{
		Identifier: opts.Identifier,
		Secret:     secret,
		Events:     WebhookEvents,
		Endpoint:   webhookEndpoint(wc.Hostname(), webhookID.String()),
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating webhook")
	}
	// return webhook for persistence to db
	return &Webhook{
		WebhookID:   webhookID,
		VCSID:       id,
		Secret:      secret,
		Identifier:  opts.Identifier,
		cloudConfig: cloudConfig,
	}, nil
}

type WebhookUpdater struct {
	VCSProviderService // for creating webhook on vcs provider
	HostnameService    // for retrieving system hostname
}

type WebhookUpdaterOptions struct {
	ProviderID string `schema:"vcs_provider_id,required"`

	*Webhook
}

func (wc *WebhookUpdater) Update(ctx context.Context, opts WebhookUpdaterOptions) (string, error) {
	createOpts := CreateWebhookOptions{
		Identifier: opts.Identifier,
		Secret:     opts.Secret,
		Events:     WebhookEvents,
		Endpoint:   webhookEndpoint(wc.Hostname(), opts.WebhookID.String()),
	}

	// first retrieve webhook from vcs provider
	vcsHook, err := wc.GetWebhook(ctx, opts.ProviderID, GetWebhookOptions{
		Identifier: opts.Identifier,
		ID:         opts.VCSID,
	})
	if errors.Is(err, ErrResourceNotFound) {
		// webhook not found, need to create it
		return wc.CreateWebhook(ctx, opts.ProviderID, createOpts)
	} else if err != nil {
		return "", err
	}

	// webhook exists; check if it needs updating
	if webhookDiff(vcsHook, opts.Webhook, wc.Hostname()) {
		err := wc.UpdateWebhook(ctx, opts.ProviderID, UpdateWebhookOptions{
			ID:                   vcsHook.ID,
			CreateWebhookOptions: createOpts,
		})
		if err != nil {
			return "", err
		}
	}
	return vcsHook.ID, nil
}

type WebhookRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
}

func (u *Unmarshaler) UnmarshalWebhookRow(row WebhookRow) (*Webhook, error) {
	cloudConfig, err := u.GetCloudConfig(row.Cloud.String)
	if err != nil {
		return nil, fmt.Errorf("unknown cloud: %s", cloudConfig)
	}

	return &Webhook{
		WebhookID:   row.WebhookID.Bytes,
		VCSID:       row.VCSID.String,
		Secret:      row.Secret.String,
		Identifier:  row.Identifier.String,
		cloudConfig: cloudConfig,
	}, nil
}

// webhookDiff determines whether the webhook config on the vcs provider differs from
// what we expect the config to be.
//
// NOTE: we cannot determine whether secret has changed because cloud APIs tend
// not to expose it
func webhookDiff(vcs *VCSWebhook, db *Webhook, hostname string) bool {
	if !reflect.DeepEqual(vcs.Events, WebhookEvents) {
		return true
	}

	endpoint := webhookEndpoint(hostname, db.WebhookID.String())
	return vcs.Endpoint != endpoint
}

// webhookEndpoint returns the URL to the webhook's endpoint. Host is the
// externally-facing hostname[:port] of otfd.
func webhookEndpoint(host, id string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   host,
		Path:   path.Join("/webhooks/vcs", id),
	}).String()
}
