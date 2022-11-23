package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) GetOrCreateWebhook(ctx context.Context, hook *otf.Webhook) (*otf.Webhook, error) {
	result, err := db.FindOrInsertWebhook(ctx, pggen.FindOrInsertWebhookParams{
		WebhookID:  UUID(hook.WebhookID),
		Secret:     String(hook.Secret),
		Identifier: String(hook.Identifier),
		HTTPURL:    String(hook.HTTPURL),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalWebhookRow(result), nil
}

func (db *DB) GetWebhookSecret(ctx context.Context, id uuid.UUID) (string, error) {
	secret, err := db.FindWebhookSecret(ctx, UUID(id))
	return secret.String, databaseError(err)
}

func (db *DB) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	_, err := db.Querier.DeleteWebhook(ctx, UUID(id))
	return databaseError(err)
}
