package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateUnsynchronisedWebhook attempts to persist an unsynchronised webhook. If a
// synchronised webhook already exists then it's returned; otherwise nil is
// returned.
func (db *DB) CreateUnsynchronisedWebhook(ctx context.Context, hook *otf.UnsynchronisedWebhook) (*otf.Webhook, error) {
	row, err := db.InsertWebhook(ctx, pggen.InsertWebhookParams{
		WebhookID:  UUID(hook.ID()),
		Secret:     String(hook.Secret()),
		Identifier: String(hook.Identifier()),
		Cloud:      String(hook.CloudName()),
	})
	if err != nil {
		return nil, databaseError(err)
	}
	if row.VCSID.Status == pgtype.Null {
		// unsynchronised hook persisted, so nothing to return.
		return nil, nil
	}
	return db.UnmarshalWebhookRow(otf.WebhookRow(row))
}

func (db *DB) SynchroniseWebhook(ctx context.Context, webhookID uuid.UUID, cloudID string) (*otf.Webhook, error) {
	result, err := db.UpdateWebhookVCSID(ctx, String(cloudID), UUID(webhookID))
	if err != nil {
		return nil, databaseError(err)
	}
	hook, err := db.UnmarshalWebhookRow(otf.WebhookRow(result))
	if err != nil {
		return nil, err
	}
	return hook, nil
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

// DeleteWebhook deletes the webhook from the database but not before it
// decrements the number of 'connections' to the webhook and only if the number
// is zero does it delete the webhook; otherwise it returns
// ErrWebhookConnected.
func (db *DB) DeleteWebhook(ctx context.Context, id uuid.UUID) error {
	var isConnected bool

	err := db.tx(ctx, func(db *DB) error {
		connections, err := db.DisconnectWebhook(ctx, UUID(id))
		if err != nil {
			return err
		}
		if connections == 0 {
			_, err := db.Querier.DeleteWebhook(ctx, UUID(id))
			return err
		}
		isConnected = true
		return nil
	})
	if err != nil {
		return databaseError(err)
	}
	if isConnected {
		return otf.ErrWebhookConnected
	}
	return nil
}
