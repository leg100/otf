package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) SynchroniseWebhook(ctx context.Context, unsynced *otf.UnsynchronisedWebhook, cb func(*otf.Webhook) (string, error)) (*otf.Webhook, error) {
	hook := &otf.Webhook{
		UnsynchronisedWebhook: unsynced,
	}
	err := db.tx(ctx, func(db *DB) error {
		upsertResult, err := db.UpsertWebhook(ctx, pggen.UpsertWebhookParams{
			WebhookID:  UUID(unsynced.ID()),
			Secret:     String(unsynced.Secret()),
			Identifier: String(unsynced.Identifier()),
			Cloud:      String(unsynced.Cloud()),
		})
		if err != nil {
			return err
		}
		if upsertResult.VCSID.Status == pgtype.Present {
			hook, err = db.UnmarshalWebhookRow(otf.WebhookRow(upsertResult))
			if err != nil {
				return err
			}
		}
		cloudID, err := cb(hook)
		if err != nil {
			return err
		}
		updateResult, err := db.UpdateWebhookVCSID(ctx, String(cloudID), UUID(hook.ID()))
		if err != nil {
			return err
		}
		hook, err = db.UnmarshalWebhookRow(otf.WebhookRow(updateResult))
		if err != nil {
			return err
		}
		return err
	})
	return hook, databaseError(err)
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
func (db *DB) DeleteWebhook(ctx context.Context, id uuid.UUID) (*otf.Webhook, error) {
	var row pggen.DisconnectWebhookRow
	err := db.tx(ctx, func(db *DB) (err error) {
		row, err = db.DisconnectWebhook(ctx, UUID(id))
		if err != nil {
			return err
		}
		if row.Connected > 0 {
			return nil
		}
		_, err = db.Querier.DeleteWebhook(ctx, UUID(id))
		return err
	})
	if err != nil {
		return nil, databaseError(err)
	}
	if row.Connected > 0 {
		return nil, otf.ErrWebhookConnected
	}
	return db.UnmarshalWebhookRow(otf.WebhookRow(row))
}
