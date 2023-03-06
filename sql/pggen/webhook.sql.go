// Code generated by pggen. DO NOT EDIT.

package pggen

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const insertWebhookSQL = `INSERT INTO webhooks (
    webhook_id,
    vcs_id,
    secret,
    identifier,
    cloud
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *
;`

type InsertWebhookParams struct {
	WebhookID  pgtype.UUID
	VCSID      pgtype.Text
	Secret     pgtype.Text
	Identifier pgtype.Text
	Cloud      pgtype.Text
}

type InsertWebhookRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
}

// InsertWebhook implements Querier.InsertWebhook.
func (q *DBQuerier) InsertWebhook(ctx context.Context, params InsertWebhookParams) (InsertWebhookRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "InsertWebhook")
	row := q.conn.QueryRow(ctx, insertWebhookSQL, params.WebhookID, params.VCSID, params.Secret, params.Identifier, params.Cloud)
	var item InsertWebhookRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("query InsertWebhook: %w", err)
	}
	return item, nil
}

// InsertWebhookBatch implements Querier.InsertWebhookBatch.
func (q *DBQuerier) InsertWebhookBatch(batch genericBatch, params InsertWebhookParams) {
	batch.Queue(insertWebhookSQL, params.WebhookID, params.VCSID, params.Secret, params.Identifier, params.Cloud)
}

// InsertWebhookScan implements Querier.InsertWebhookScan.
func (q *DBQuerier) InsertWebhookScan(results pgx.BatchResults) (InsertWebhookRow, error) {
	row := results.QueryRow()
	var item InsertWebhookRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("scan InsertWebhookBatch row: %w", err)
	}
	return item, nil
}

const updateWebhookVCSIDSQL = `UPDATE webhooks
SET vcs_id = $1
WHERE webhook_id = $2
RETURNING *
;`

type UpdateWebhookVCSIDRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
}

// UpdateWebhookVCSID implements Querier.UpdateWebhookVCSID.
func (q *DBQuerier) UpdateWebhookVCSID(ctx context.Context, vcsID pgtype.Text, webhookID pgtype.UUID) (UpdateWebhookVCSIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "UpdateWebhookVCSID")
	row := q.conn.QueryRow(ctx, updateWebhookVCSIDSQL, vcsID, webhookID)
	var item UpdateWebhookVCSIDRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("query UpdateWebhookVCSID: %w", err)
	}
	return item, nil
}

// UpdateWebhookVCSIDBatch implements Querier.UpdateWebhookVCSIDBatch.
func (q *DBQuerier) UpdateWebhookVCSIDBatch(batch genericBatch, vcsID pgtype.Text, webhookID pgtype.UUID) {
	batch.Queue(updateWebhookVCSIDSQL, vcsID, webhookID)
}

// UpdateWebhookVCSIDScan implements Querier.UpdateWebhookVCSIDScan.
func (q *DBQuerier) UpdateWebhookVCSIDScan(results pgx.BatchResults) (UpdateWebhookVCSIDRow, error) {
	row := results.QueryRow()
	var item UpdateWebhookVCSIDRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("scan UpdateWebhookVCSIDBatch row: %w", err)
	}
	return item, nil
}

const findWebhookByIDSQL = `SELECT *
FROM webhooks
WHERE webhook_id = $1;`

type FindWebhookByIDRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
}

// FindWebhookByID implements Querier.FindWebhookByID.
func (q *DBQuerier) FindWebhookByID(ctx context.Context, webhookID pgtype.UUID) (FindWebhookByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWebhookByID")
	row := q.conn.QueryRow(ctx, findWebhookByIDSQL, webhookID)
	var item FindWebhookByIDRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("query FindWebhookByID: %w", err)
	}
	return item, nil
}

// FindWebhookByIDBatch implements Querier.FindWebhookByIDBatch.
func (q *DBQuerier) FindWebhookByIDBatch(batch genericBatch, webhookID pgtype.UUID) {
	batch.Queue(findWebhookByIDSQL, webhookID)
}

// FindWebhookByIDScan implements Querier.FindWebhookByIDScan.
func (q *DBQuerier) FindWebhookByIDScan(results pgx.BatchResults) (FindWebhookByIDRow, error) {
	row := results.QueryRow()
	var item FindWebhookByIDRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("scan FindWebhookByIDBatch row: %w", err)
	}
	return item, nil
}

const findWebhooksByRepoSQL = `SELECT *
FROM webhooks
WHERE identifier = $1
AND   cloud = $2;`

type FindWebhooksByRepoRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
}

// FindWebhooksByRepo implements Querier.FindWebhooksByRepo.
func (q *DBQuerier) FindWebhooksByRepo(ctx context.Context, identifier pgtype.Text, cloud pgtype.Text) ([]FindWebhooksByRepoRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "FindWebhooksByRepo")
	rows, err := q.conn.Query(ctx, findWebhooksByRepoSQL, identifier, cloud)
	if err != nil {
		return nil, fmt.Errorf("query FindWebhooksByRepo: %w", err)
	}
	defer rows.Close()
	items := []FindWebhooksByRepoRow{}
	for rows.Next() {
		var item FindWebhooksByRepoRow
		if err := rows.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
			return nil, fmt.Errorf("scan FindWebhooksByRepo row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindWebhooksByRepo rows: %w", err)
	}
	return items, err
}

// FindWebhooksByRepoBatch implements Querier.FindWebhooksByRepoBatch.
func (q *DBQuerier) FindWebhooksByRepoBatch(batch genericBatch, identifier pgtype.Text, cloud pgtype.Text) {
	batch.Queue(findWebhooksByRepoSQL, identifier, cloud)
}

// FindWebhooksByRepoScan implements Querier.FindWebhooksByRepoScan.
func (q *DBQuerier) FindWebhooksByRepoScan(results pgx.BatchResults) ([]FindWebhooksByRepoRow, error) {
	rows, err := results.Query()
	if err != nil {
		return nil, fmt.Errorf("query FindWebhooksByRepoBatch: %w", err)
	}
	defer rows.Close()
	items := []FindWebhooksByRepoRow{}
	for rows.Next() {
		var item FindWebhooksByRepoRow
		if err := rows.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
			return nil, fmt.Errorf("scan FindWebhooksByRepoBatch row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("close FindWebhooksByRepoBatch rows: %w", err)
	}
	return items, err
}

const deleteWebhookByIDSQL = `DELETE
FROM webhooks
WHERE webhook_id = $1
RETURNING *
;`

type DeleteWebhookByIDRow struct {
	WebhookID  pgtype.UUID `json:"webhook_id"`
	VCSID      pgtype.Text `json:"vcs_id"`
	Secret     pgtype.Text `json:"secret"`
	Identifier pgtype.Text `json:"identifier"`
	Cloud      pgtype.Text `json:"cloud"`
}

// DeleteWebhookByID implements Querier.DeleteWebhookByID.
func (q *DBQuerier) DeleteWebhookByID(ctx context.Context, webhookID pgtype.UUID) (DeleteWebhookByIDRow, error) {
	ctx = context.WithValue(ctx, "pggen_query_name", "DeleteWebhookByID")
	row := q.conn.QueryRow(ctx, deleteWebhookByIDSQL, webhookID)
	var item DeleteWebhookByIDRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("query DeleteWebhookByID: %w", err)
	}
	return item, nil
}

// DeleteWebhookByIDBatch implements Querier.DeleteWebhookByIDBatch.
func (q *DBQuerier) DeleteWebhookByIDBatch(batch genericBatch, webhookID pgtype.UUID) {
	batch.Queue(deleteWebhookByIDSQL, webhookID)
}

// DeleteWebhookByIDScan implements Querier.DeleteWebhookByIDScan.
func (q *DBQuerier) DeleteWebhookByIDScan(results pgx.BatchResults) (DeleteWebhookByIDRow, error) {
	row := results.QueryRow()
	var item DeleteWebhookByIDRow
	if err := row.Scan(&item.WebhookID, &item.VCSID, &item.Secret, &item.Identifier, &item.Cloud); err != nil {
		return item, fmt.Errorf("scan DeleteWebhookByIDBatch row: %w", err)
	}
	return item, nil
}
