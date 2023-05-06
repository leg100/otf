package repo

import (
	"context"
	"reflect"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

// defaultEvents are the VCS events that hooks subscribe to.
var defaultEvents = []cloud.VCSEventType{
	cloud.VCSPushEventType,
	cloud.VCSPullEventType,
}

// hook is a webhook for a VCS repo
type hook struct {
	id      uuid.UUID // internal otf ID
	cloudID *string   // cloud's hook ID, populated following synchronisation

	secret     string // secret token
	identifier string // repo identifier: <repo_owner>/<repo_name>
	cloud      string // cloud name
	endpoint   string // otf URL that receives events

	cloud.EventHandler // handles incoming vcs events
}

// sync synchronises a hook with the cloud
func (h *hook) sync(ctx context.Context, client cloud.Client) error {
	if h.cloudID == nil {
		cloudID, err := client.CreateWebhook(ctx, h.createOpts())
		if err != nil {
			return err
		}
		h.cloudID = &cloudID
		return nil
	}

	// existing hook in DB; check it exists in cloud and create/update
	// accordingly
	cloudHook, err := client.GetWebhook(ctx, cloud.GetWebhookOptions{
		Repo: h.identifier,
		ID:   *h.cloudID,
	})
	if errors.Is(err, otf.ErrResourceNotFound) {
		// hook not found in cloud; create it
		cloudID, err := client.CreateWebhook(ctx, h.createOpts())
		if err != nil {
			return err
		}
		h.cloudID = &cloudID
		return nil
	} else if err != nil {
		return errors.Wrap(err, "retrieving config from cloud")
	}

	// hook found in both DB and on cloud; check for differences and update
	// accordingly
	if reflect.DeepEqual(defaultEvents, cloudHook.Events) &&
		h.endpoint == cloudHook.Endpoint {
		// no differences
		return nil
	}

	// differences found; update hook on cloud
	err = client.UpdateWebhook(ctx, cloud.UpdateWebhookOptions{
		ID:                   cloudHook.ID,
		CreateWebhookOptions: h.createOpts(),
	})
	return err
}

func (h *hook) createOpts() cloud.CreateWebhookOptions {
	return cloud.CreateWebhookOptions{
		Repo:     h.identifier,
		Secret:   h.secret,
		Events:   defaultEvents,
		Endpoint: h.endpoint,
	}
}
