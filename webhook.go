package otf

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/leg100/otf/sql/pggen"
)

// WebhookEvents are those events webhooks should subscribe to.
const WebhookEvents = []VCSEventType{
	VCSPushEventType,
	VCSPullEventType,
}

type Webhook struct {
	WebhookID  uuid.UUID // otf's ID
	VCSID      *string   // vcs provider's webhook ID
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
	SyncWebhook(ctx context.Context, hook *Webhook, idUpdater func(*Webhook) (string, error)) (*Webhook, error)
	GetWebhookSecret(ctx context.Context, id uuid.UUID) (string, error)
	// DeleteWebhook deletes the webhook from the store.
	DeleteWebhook(ctx context.Context, id uuid.UUID) error
}

func NewWebhook(identifier, httpURL string) (*Webhook, error) {
	secret, err := GenerateToken()
	if err != nil {
		return nil, err
	}
	return &Webhook{
		WebhookID:  uuid.New(),
		Secret:     secret,
		Identifier: identifier,
		HTTPURL:    httpURL,
	}, nil
}

// WebhookEndpoint returns the URL to the webhook's endpoint. Host is the
// externally-facing hostname[:port] of otfd.
func WebhookEndpoint(id, host, cloud string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   host,
		Path:   path.Join("/webhooks/vcs/", id, cloud),
	}).String()
}

func UnmarshalWebhookRow(row pggen.FindOrInsertWebhookRow) *Webhook {
	return &Webhook{
		WebhookID:  row.WebhookID.Bytes,
		Secret:     row.Secret.String,
		Identifier: row.Identifier.String,
		HTTPURL:    row.HTTPURL.String,
	}
}

type WebhookCreatorOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	HTTPURL    string `schema:"http_url,required"`   // complete HTTP/S URL for repo
	ProviderID string `schema:"vcs_provider_id,required"`
	Branch     string `schema:"branch,required"`
}
