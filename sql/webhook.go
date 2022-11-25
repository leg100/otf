package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) SyncWebhook(ctx context.Context, hook *otf.Webhook, idUpdater func(*otf.Webhook) (string, error)) (*otf.Webhook, error) {
	err := db.tx(ctx, func(tx *DB) error {
		// Try to retrieve existing hook first from store and if it doesn't exist,
		// create it.
		result, err := db.FindOrInsertWebhook(ctx, pggen.FindOrInsertWebhookParams{
			WebhookID:  UUID(hook.WebhookID),
			Secret:     String(hook.Secret),
			Identifier: String(hook.Identifier),
			HTTPURL:    String(hook.HTTPURL),
		})
		if err != nil {
			return databaseError(err)
		}
		// If hook already exists then that has precedence over what the
		// caller provided
		hook, err = otf.UnmarshalWebhookRow(result), nil
		if err != nil {
			return databaseError(err)
		}

		id, err := idUpdater(hook)
		if err != nil {
			return err
		}

		// Store vcs provider's webhook ID if it's yet to be stored, or if it differs
		// from what is currently in the store.
		if hook.VCSID == nil || *hook.VCSID != id {
			_, err = db.UpdateWebhookVCSID(ctx, String(id), UUID(hook.WebhookID))
			if err != nil {
				return databaseError(err)
			}
		}

		return nil
	})
	return hook, databaseError(err)
}

func (db *DB) GetWebhookSecret(ctx context.Context, id uuid.UUID) (string, error) {
	secret, err := db.FindWebhookSecret(ctx, UUID(id))
	return secret.String, databaseError(err)
}

func (db *DB) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	_, err := db.Querier.DeleteWebhook(ctx, UUID(id))
	return databaseError(err)
}
