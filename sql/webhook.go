package sql

import (
	"context"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// SyncWebhook synchronises a webhook, creating it if it doesn't exist already,
// or if it does exist it'll check for differences and update if necessary.
func (db *DB) SyncWebhook(ctx context.Context, opts otf.SyncWebhookOptions) (*otf.Webhook, error) {
	var hook *otf.Webhook
	err := db.tx(ctx, func(tx *DB) error {
		// Prevent any modifications to table because we're checking first whether webhook with given
		// url exists and if it does not then we're creating a webhook with that
		// url, and we do not want another process to do the same thing
		// in parallel and create a webhook in the intervening period...
		_, err := tx.Exec(ctx, "LOCK webhooks IN EXCLUSIVE MODE")
		if err != nil {
			return errors.Wrap(err, "locking webhook table")
		}
		result, err := tx.FindWebhookByRepo(ctx, String(opts.Identifier), String(opts.Cloud))
		if err != nil {
			err = databaseError(err)
			if errors.Is(err, otf.ErrResourceNotFound) {
				// create webhook
				hook, err = opts.CreateWebhookFunc(ctx, otf.WebhookCreatorOptions{
					ProviderID: opts.ProviderID,
					Identifier: opts.Identifier,
					Cloud:      opts.Cloud,
				})
				if err != nil {
					return err
				}
				// and persist
				_, err = tx.InsertWebhook(ctx, pggen.InsertWebhookParams{
					WebhookID:  UUID(hook.WebhookID),
					VCSID:      String(hook.VCSID),
					Secret:     String(hook.Secret),
					Identifier: String(hook.Identifier),
					Cloud:      String(hook.CloudName()),
				})
				if err != nil {
					return databaseError(err)
				}
				return nil
			}
			return err
		} else {
			hook, err = db.UnmarshalWebhookRow(otf.WebhookRow(result))
			if err != nil {
				return err
			}

			id, err := opts.UpdateWebhookFunc(ctx, otf.WebhookUpdaterOptions{
				ProviderID: opts.ProviderID,
				Webhook:    hook,
			})
			if err != nil {
				return err
			}
			// Update VCS' ID if has changed.
			if hook.VCSID != id {
				_, err = tx.UpdateWebhookVCSID(ctx, String(id), UUID(hook.WebhookID))
				if err != nil {
					return databaseError(err)
				}
				hook.VCSID = id
			}
			return nil
		}
	})
	return hook, err
}

func (db *DB) GetWebhook(ctx context.Context, id uuid.UUID) (*otf.Webhook, error) {
	result, err := db.FindWebhookByID(ctx, UUID(id))
	if err != nil {
		return nil, databaseError(err)
	}
	hook, err := db.UnmarshalWebhookRow(otf.WebhookRow(result))
	if err != nil {
		return nil, err
	}
	return hook, nil
}

func (db *DB) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	_, err := db.Querier.DeleteWebhook(ctx, UUID(id))
	return databaseError(err)
}
