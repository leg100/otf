package otf

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/leg100/otf/cloud"
	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

var (

	// DefaultWebhookEvents are VCS events webhooks should subscribe to.
	DefaultWebhookEvents = []cloud.VCSEventType{
		cloud.VCSPushEventType,
		cloud.VCSPullEventType,
	}
	// ErrWebhookConnected is returned when an attempt is made to delete a
	// webhook but the webhook is still connected e.g. to a module or a
	// workspace.
	ErrWebhookConnected = errors.New("webhook is still connected")
)

// Webhook is a VCS repo webhook configuration present on both a cloud (e.g.
// github) as well as in OTF, i.e. it is synchronised.
type Webhook struct {
	*UnsynchronisedWebhook

	cloudID string // cloud's webhook ID
}

func (h *Webhook) VCSID() string { return h.cloudID }

func (h *Webhook) HandleEvent(w http.ResponseWriter, r *http.Request) cloud.VCSEvent {
	return h.cloudConfig.HandleEvent(w, r, cloud.HandleEventOptions{
		WebhookID: h.id,
		Secret:    h.secret,
	})
}

// UnsynchronisedWebhook is a VCS repo webhook configuration that is yet to be
// synchronised to a cloud e.g. github.
type UnsynchronisedWebhook struct {
	id          uuid.UUID    // otf's webhook ID
	secret      string       // secret token
	identifier  string       // identifier is <repo_owner>/<repo_name>
	cloudConfig cloud.Config // identifies cloud and provides VCS event handler
}

func NewUnsynchronisedWebhook(opts NewUnsynchronisedWebhookOptions) (*UnsynchronisedWebhook, error) {
	secret, err := GenerateToken()
	if err != nil {
		return nil, err
	}
	return &UnsynchronisedWebhook{
		id:          uuid.New(),
		secret:      secret,
		identifier:  opts.Identifier,
		cloudConfig: opts.CloudConfig,
	}, nil
}

type NewUnsynchronisedWebhookOptions struct {
	Identifier  string
	CloudConfig cloud.Config
}

func (h *UnsynchronisedWebhook) ID() uuid.UUID      { return h.id }
func (h *UnsynchronisedWebhook) Owner() string      { return strings.Split(h.identifier, "/")[0] }
func (h *UnsynchronisedWebhook) Repo() string       { return strings.Split(h.identifier, "/")[1] }
func (h *UnsynchronisedWebhook) Identifier() string { return h.identifier }
func (h *UnsynchronisedWebhook) Secret() string     { return h.secret }
func (h *UnsynchronisedWebhook) CloudName() string  { return h.cloudConfig.Name }

// Endpoint returns the webhook's endpoint, using the otf host (hostname:[port])
// to build the endpoint URL.
func (h *UnsynchronisedWebhook) Endpoint(host string) string {
	return (&url.URL{
		Scheme: "https",
		Host:   host,
		Path:   path.Join("/webhooks/vcs", h.id.String()),
	}).String()
}

type WebhookService interface {
	CreateWebhook(ctx context.Context, opts SynchroniseWebhookOptions) (*Webhook, error)
	GetWebhook(ctx context.Context, workspaceID string) (*Webhook, error)
	GetWebhookByName(ctx context.Context, organization, workspace string) (*Webhook, error)
	DeleteWebhook(ctx context.Context, workspaceID string) (*Webhook, error)
}

type WebhookStore interface {
	// CreateWebhook idempotently persists an unsynchronised webhook to the
	// store. If a (synchronised) webhook already exists then it is returned.
	CreateUnsynchronisedWebhook(context.Context, *UnsynchronisedWebhook) (*Webhook, error)
	// SynchroniseWebhook synchronises a webhook in the store, setting or updating its
	// cloud ID, and returning the synchronised webhook.
	SynchroniseWebhook(ctx context.Context, webhookID uuid.UUID, cloudID string) (*Webhook, error)
	// GetWebhook retrieves a webhook by its ID
	GetWebhook(ctx context.Context, webhookID uuid.UUID) (*Webhook, error)
	// DeleteWebhook deletes the webhook from the store. If the webhook is still
	// connected (e.g. to a module or a workspace) then ErrWebhookConnected is
	// returned.
	DeleteWebhook(ctx context.Context, webhookID uuid.UUID) error
}

type SynchroniseWebhookOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud providing webhook
}

// WebhookSynchroniser synchronises a webhook, ensuring its config is present
// and identical in both OTF and on the cloud.
type WebhookSynchroniser struct {
	Application
}

func (wc *WebhookSynchroniser) Synchronise(ctx context.Context, opts SynchroniseWebhookOptions) (*Webhook, error) {
	// lookup cloudConfig using cloud name
	cloudConfig, err := wc.GetCloudConfig(opts.Cloud)
	if err != nil {
		return nil, err
	}

	unsynced, err := NewUnsynchronisedWebhook(NewUnsynchronisedWebhookOptions{
		Identifier:  opts.Identifier,
		CloudConfig: cloudConfig,
	})
	if err != nil {
		return nil, err
	}

	client, err := wc.GetVCSClient(ctx, opts.ProviderID)
	if err != nil {
		return nil, err
	}

	// Wrap the process within a transaction to prevent hooks from being
	// synchronised concurrently, which could lead to disparities in
	// configuration.
	var hook *Webhook
	err = wc.Tx(ctx, func(app Application) (err error) {
		hook, err = app.DB().CreateUnsynchronisedWebhook(ctx, unsynced)
		if err != nil {
			return err
		}
		if hook == nil {
			// no existing hook; create in cloud and sync
			cloudID, err := client.CreateWebhook(ctx, cloud.CreateWebhookOptions{
				Identifier: opts.Identifier,
				Secret:     unsynced.secret,
				Events:     DefaultWebhookEvents,
				Endpoint:   unsynced.Endpoint(wc.Hostname()),
			})
			if err != nil {
				return err
			}
			hook, err = app.DB().SynchroniseWebhook(ctx, unsynced.id, cloudID)
			if err != nil {
				return err
			}
			// synchronised
			return nil
		}

		// existing hook; check it exists in cloud and create/update
		// accordingly
		cloudHook, err := client.GetWebhook(ctx, cloud.GetWebhookOptions{
			Identifier: opts.Identifier,
			ID:         hook.cloudID,
		})
		if errors.Is(err, ErrResourceNotFound) {
			// hook not found in cloud; create it
			cloudID, err := client.CreateWebhook(ctx, cloud.CreateWebhookOptions{
				Identifier: opts.Identifier,
				Secret:     hook.secret,
				Events:     DefaultWebhookEvents,
				Endpoint:   hook.Endpoint(wc.Hostname()),
			})
			if err != nil {
				return err
			}
			hook, err = app.DB().SynchroniseWebhook(ctx, hook.id, cloudID)
			if err != nil {
				return err
			}
			// synchronised
			return nil
		}

		// hook found in both DB and on cloud; check for differences and update
		// accordingly
		if reflect.DeepEqual(DefaultWebhookEvents, cloudHook.Events) &&
			hook.Endpoint(wc.Hostname()) == cloudHook.Endpoint {
			// synchronised
			return nil
		}

		// differences found; update hook on cloud
		err = client.UpdateWebhook(ctx, cloud.UpdateWebhookOptions{
			ID: cloudHook.ID,
			CreateWebhookOptions: cloud.CreateWebhookOptions{
				Identifier: opts.Identifier,
				Secret:     hook.secret,
				Events:     DefaultWebhookEvents,
				Endpoint:   hook.Endpoint(wc.Hostname()),
			},
		})
		if err != nil {
			return err
		}

		// synchronised
		return nil
	})
	return hook, err
}

type WebhookRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
	Connected  int         `json:"connected"`
}

func (u *Unmarshaler) UnmarshalWebhookRow(row WebhookRow) (*Webhook, error) {
	cloudConfig, err := u.GetCloudConfig(row.Cloud.String)
	if err != nil {
		return nil, fmt.Errorf("unknown cloud: %s", cloudConfig)
	}

	return &Webhook{
		UnsynchronisedWebhook: &UnsynchronisedWebhook{
			id:          row.WebhookID.Bytes,
			secret:      row.Secret.String,
			identifier:  row.Identifier.String,
			cloudConfig: cloudConfig,
		},
		cloudID: row.VCSID.String,
	}, nil
}
