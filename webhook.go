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

	cloud.EventHandler        // handles incoming vcs events
	cloudID            string // cloud's webhook ID
}

func (h *Webhook) VCSID() string { return h.cloudID }

func (h *Webhook) HandleEvent(w http.ResponseWriter, r *http.Request) cloud.VCSEvent {
	return h.EventHandler.HandleEvent(w, r, cloud.HandleEventOptions{
		WebhookID: h.id,
		Secret:    h.secret,
	})
}

// UnsynchronisedWebhook is a VCS repo webhook configuration that is yet to be
// synchronised to a cloud e.g. github.
type UnsynchronisedWebhook struct {
	id         uuid.UUID // otf's webhook ID
	secret     string    // secret token
	identifier string    // identifier is <repo_owner>/<repo_name>
	cloud      string    // cloud name
}

func NewUnsynchronisedWebhook(opts NewUnsynchronisedWebhookOptions) (*UnsynchronisedWebhook, error) {
	secret, err := GenerateToken()
	if err != nil {
		return nil, err
	}
	return &UnsynchronisedWebhook{
		id:         uuid.New(),
		secret:     secret,
		identifier: opts.Identifier,
		cloud:      opts.Cloud,
	}, nil
}

type NewUnsynchronisedWebhookOptions struct {
	Identifier string
	Cloud      string
}

func (h *UnsynchronisedWebhook) ID() uuid.UUID      { return h.id }
func (h *UnsynchronisedWebhook) Owner() string      { return strings.Split(h.identifier, "/")[0] }
func (h *UnsynchronisedWebhook) Repo() string       { return strings.Split(h.identifier, "/")[1] }
func (h *UnsynchronisedWebhook) Identifier() string { return h.identifier }
func (h *UnsynchronisedWebhook) Secret() string     { return h.secret }
func (h *UnsynchronisedWebhook) Cloud() string      { return h.cloud }

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
	CreateWebhook(ctx context.Context, opts CreateWebhookOptions) (*Webhook, error)
	GetWebhook(ctx context.Context, webhookID uuid.UUID) (*Webhook, error)
	DeleteWebhook(ctx context.Context, providerID string, webhookID uuid.UUID) error
}

type WebhookStore interface {
	// SynchroniseWebhook performs the two-step task of synchronising a webhook:
	// (1) The hook is created in the DB without a cloud ID
	// (2) Callback is invoked. If a hook already exists in (1) it is passed in
	// as an argument. The callback is expected to return a cloud ID which is
	// persisted to the DB.
	//
	// Finally, the synchronised webhook is returned, complete with cloud ID.
	SynchroniseWebhook(context.Context, *UnsynchronisedWebhook, func(*Webhook) (string, error)) (*Webhook, error)
	// GetWebhook retrieves a webhook by its ID
	GetWebhook(ctx context.Context, webhookID uuid.UUID) (*Webhook, error)
	// DeleteWebhook deletes the webhook from the store. If the webhook is still
	// connected (e.g. to a module or a workspace) then ErrWebhookConnected is
	// returned.
	DeleteWebhook(ctx context.Context, webhookID uuid.UUID) (*Webhook, error)
}

type CreateWebhookOptions struct {
	Identifier string `schema:"identifier,required"` // repo id: <owner>/<repo>
	ProviderID string `schema:"vcs_provider_id,required"`
	Cloud      string // cloud providing webhook
}

// WebhookSynchroniser synchronises a webhook, ensuring its config is present
// and identical in both OTF and on the cloud.
type WebhookSynchroniser struct {
	VCSProviderService
	HostnameService
	DB
}

func (s *WebhookSynchroniser) Synchronise(ctx context.Context, providerID string, unsynced *UnsynchronisedWebhook) (*Webhook, error) {
	client, err := s.GetVCSClient(ctx, providerID)
	if err != nil {
		return nil, err
	}

	return s.SynchroniseWebhook(ctx, unsynced, func(existing *Webhook) (string, error) {
		if existing == nil {
			return client.CreateWebhook(ctx, cloud.CreateWebhookOptions{
				Identifier: unsynced.identifier,
				Secret:     unsynced.secret,
				Events:     DefaultWebhookEvents,
				Endpoint:   unsynced.Endpoint(s.Hostname()),
			})
		}

		// existing hook; check it exists in cloud and create/update
		// accordingly
		cloudHook, err := client.GetWebhook(ctx, cloud.GetWebhookOptions{
			Identifier: unsynced.identifier,
			ID:         existing.cloudID,
		})
		if errors.Is(err, ErrResourceNotFound) {
			// hook not found in cloud; create it
			return client.CreateWebhook(ctx, cloud.CreateWebhookOptions{
				Identifier: unsynced.identifier,
				Secret:     existing.secret,
				Events:     DefaultWebhookEvents,
				Endpoint:   existing.Endpoint(s.Hostname()),
			})
		}

		// hook found in both DB and on cloud; check for differences and update
		// accordingly
		if reflect.DeepEqual(DefaultWebhookEvents, cloudHook.Events) &&
			existing.Endpoint(s.Hostname()) == cloudHook.Endpoint {
			// no differences
			return existing.cloudID, nil
		}

		// differences found; update hook on cloud
		err = client.UpdateWebhook(ctx, cloud.UpdateWebhookOptions{
			ID: cloudHook.ID,
			CreateWebhookOptions: cloud.CreateWebhookOptions{
				Identifier: unsynced.identifier,
				Secret:     existing.secret,
				Events:     DefaultWebhookEvents,
				Endpoint:   existing.Endpoint(s.Hostname()),
			},
		})
		return existing.cloudID, err
	})
}

// WebhookDeleter deletes a webhook, deleting it from both DB and cloud.
type WebhookDeleter struct {
	VCSProviderService
	DB
}

func (d *WebhookDeleter) Delete(ctx context.Context, providerID string, webhookID uuid.UUID) error {
	return d.Tx(ctx, func(db DB) error {
		hook, err := db.DeleteWebhook(ctx, webhookID)
		if err == ErrWebhookConnected {
			return nil
		} else if err != nil {
			return err
		}

		client, err := d.GetVCSClient(ctx, providerID)
		if err != nil {
			return err
		}
		return client.DeleteWebhook(ctx, cloud.DeleteWebhookOptions{
			Identifier: hook.Identifier(),
			ID:         hook.VCSID(),
		})
	})
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
			id:         row.WebhookID.Bytes,
			secret:     row.Secret.String,
			identifier: row.Identifier.String,
			cloud:      row.Cloud.String,
		},
		EventHandler: cloudConfig,
		cloudID:      row.VCSID.String,
	}, nil
}
