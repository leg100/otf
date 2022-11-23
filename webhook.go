package otf

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/leg100/otf/sql/pggen"
)

type Webhook struct {
	WebhookID  uuid.UUID
	Secret     string
	Identifier string // identifier is <repo_owner>/<repo_name>
	HTTPURL    string // HTTPURL is the web url for the repo
}

func (h *Webhook) ID() string    { return h.WebhookID.String() }
func (h *Webhook) Owner() string { return strings.Split(h.Identifier, "/")[0] }
func (h *Webhook) Repo() string  { return strings.Split(h.Identifier, "/")[1] }

type WebhookStore interface {
	GetOrCreateWebhook(ctx context.Context, hook *Webhook) (*Webhook, error)
	GetWebhookSecret(ctx context.Context, id uuid.UUID) (string, error)
	// DeleteWebhook deletes the webhook from the persistence store.
	DeleteWebhook(ctx context.Context, id uuid.UUID) error
}

// CreateWebhookOptions are options for creating a webhook on a cloud provider.
type CreateWebhookOptions struct {
	URL        string // otfd event handler endpoint
	Identifier string // Repository identifier
	Secret     string // Secret key for signing events
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
