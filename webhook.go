package otf

import (
	"context"
	"errors"
	"net/url"
	"path"
	"reflect"
	"strings"

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
	HTTPURL    string    // HTTPURL is the web url for the repo
}

func (h *Webhook) ID() string    { return h.WebhookID.String() }
func (h *Webhook) Owner() string { return strings.Split(h.Identifier, "/")[0] }
func (h *Webhook) Repo() string  { return strings.Split(h.Identifier, "/")[1] }

type WebhookStore interface {
	// SyncWebhook ensures webhook configuration is present and
	// equal in both the store and the VCS provider. The idUpdater is called
	// after the webhook is created in the store (or retrieved if it already
	// exists) and it should be used to ensure the webhook exists and is up to
	// date in the VCS provider before returning the ID the provider uses to
	// identify the webhook. SyncWebhook will then update the store with the ID if
	// it differs from its present value in the store.
	SyncWebhook(ctx context.Context, opts SyncWebhookOptions) (*Webhook, error)
	GetWebhookSecret(ctx context.Context, id uuid.UUID) (string, error)
	// DeleteWebhook deletes the webhook from the store.
	DeleteWebhook(ctx context.Context, id uuid.UUID) error
}

type SyncWebhookOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	HTTPURL    string `schema:"http_url,required"`   // complete HTTP/S URL for repo
	ProviderID string `schema:"vcs_provider_id,required"`
	OTFHost    string // otf host

	CreateWebhookFunc func(context.Context, WebhookCreatorOptions) (*Webhook, error)
	UpdateWebhookFunc func(context.Context, WebhookUpdaterOptions) (string, error)
}

type WebhookCreatorOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	HTTPURL    string `schema:"http_url,required"`   // complete HTTP/S URL for repo
	ProviderID string `schema:"vcs_provider_id,required"`
	OTFHost    string // otf host
}

type WebhookCreator struct {
	VCSProviderService // for creating webhook on vcs provider
	WebhookStore       // for persisting webhook to db
}

func (wc *WebhookCreator) Create(ctx context.Context, opts WebhookCreatorOptions) (*Webhook, error) {
	secret, err := GenerateToken()
	if err != nil {
		return nil, err
	}
	webhookID := uuid.New()
	endpoint := webhookEndpoint(opts.OTFHost, webhookID.String())

	// create webhook on vcs provider
	id, err := wc.VCSProviderService.CreateWebhook(ctx, opts.ProviderID, CreateWebhookOptions{
		Identifier: opts.Identifier,
		Secret:     secret,
		Events:     WebhookEvents,
		OTFHost:    endpoint,
	})
	if err != nil {
		return nil, err
	}
	// now persist it to the db
	return &Webhook{
		WebhookID:  uuid.New(),
		VCSID:      id,
		Secret:     secret,
		Identifier: opts.Identifier,
		HTTPURL:    opts.HTTPURL,
	}, nil
}

type WebhookUpdater struct {
	VCSProviderService // for creating webhook on vcs provider
}

type WebhookUpdaterOptions struct {
	ProviderID string `schema:"vcs_provider_id,required"`
	OTFHost    string

	*Webhook
}

func (wc *WebhookUpdater) Update(ctx context.Context, opts WebhookUpdaterOptions) (string, error) {
	// first retrieve webhook from vcs provider
	vcsHook, err := wc.GetWebhook(ctx, opts.ProviderID, GetWebhookOptions{
		Identifier: opts.Identifier,
		ID:         opts.VCSID,
	})
	if errors.Is(err, ErrResourceNotFound) {
		// webhook not found, need to create it
		return wc.CreateWebhook(ctx, opts.ProviderID, CreateWebhookOptions{
			Identifier: opts.Identifier,
			Secret:     opts.Secret,
			Events:     WebhookEvents,
			OTFHost:    webhookEndpoint(opts.OTFHost, opts.WebhookID.String()),
		})
	} else if err != nil {
		return "", err
	}

	// webhook exists; check if it needs updating
	if webhookDiff(vcsHook, opts.Webhook, opts.OTFHost) {
		err := wc.UpdateWebhook(ctx, opts.ProviderID, UpdateWebhookOptions{
			ID: vcsHook.ID,
			CreateWebhookOptions: CreateWebhookOptions{
				Identifier: opts.Identifier,
				Secret:     opts.Secret,
				OTFHost:    opts.OTFHost,
			},
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
	HTTPURL    pgtype.Text `json:"http_url"`
}

func UnmarshalWebhookRow(row WebhookRow) *Webhook {
	return &Webhook{
		WebhookID:  row.WebhookID.Bytes,
		VCSID:      row.VCSID.String,
		Secret:     row.Secret.String,
		Identifier: row.Identifier.String,
		HTTPURL:    row.HTTPURL.String,
	}
}

// webhookDiff determines whether the webhook config on the vcs provider differs from
// what we expect the config to be.
//
// NOTE: we cannot determine whether secret has changed because cloud APIs tend
// not to expose it
func webhookDiff(vcs *VCSWebhook, db *Webhook, otfHost string) bool {
	if !reflect.DeepEqual(vcs.Events, WebhookEvents) {
		return true
	}

	endpoint := webhookEndpoint(otfHost, db.WebhookID.String())
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
