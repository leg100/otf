package sql

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) SyncWebhook(ctx context.Context, opts otf.SyncWebhookOptions) (*otf.Webhook, error) {
	var hook *otf.Webhook
	err := db.tx(ctx, func(tx *DB) error {
		// Prevent any modifications to table because we're checking first whether webhook with given
		// url exists and if it does not then we're creating a webhook with that
		// url, and we do not want another process to do the same thing
		// in parallel and create a webhook in the intervening period...
		_, err := tx.Exec(ctx, "LOCK webhooks IN EXCLUSIVE MODE")
		if err != nil {
			return err
		}
		result, err := tx.FindWebhookByURL(ctx, String(opts.HTTPURL))
		if err != nil {
			err = databaseError(err)
			if errors.Is(err, otf.ErrResourceNotFound) {
				// create webhook
				hook, err = opts.CreateWebhookFunc(ctx, otf.WebhookCreatorOptions{
					ProviderID: opts.ProviderID,
					Identifier: opts.Identifier,
					HTTPURL:    opts.HTTPURL,
					OTFHost:    opts.OTFHost,
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
					HTTPURL:    String(hook.HTTPURL),
				})
				if err != nil {
					return databaseError(err)
				}
				return nil
			}
			return err
		} else {
			hook = otf.UnmarshalWebhookRow(otf.WebhookRow(result))

			id, err := opts.UpdateWebhookFunc(ctx, otf.WebhookUpdaterOptions{
				ProviderID: opts.ProviderID,
				OTFHost:    opts.OTFHost,
				Webhook:    hook,
			})
			if err != nil {
				return err
			}
			// Update VCS ID if has changed.
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

func (db *DB) GetWebhookSecret(ctx context.Context, id uuid.UUID) (string, error) {
	secret, err := db.FindWebhookSecret(ctx, UUID(id))
	return secret.String, databaseError(err)
}

func (db *DB) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	_, err := db.Querier.DeleteWebhook(ctx, UUID(id))
	return databaseError(err)
}
