package sql

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

type cvUploader struct {
	q  *pggen.DBQuerier
	id string
}

func newConfigUploader(tx pgx.Tx, id string) *cvUploader {
	return &cvUploader{
		q:  pggen.NewQuerier(tx),
		id: id,
	}
}

func (u *cvUploader) SetErrored(ctx context.Context) error {
	_, err := u.q.UpdateConfigurationVersionErroredByID(ctx, pgtype.Text{String: u.id, Status: pgtype.Present})
	if err != nil {
		return err
	}
	return nil
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (otf.ConfigurationStatus, error) {
	_, err := u.q.UpdateConfigurationVersionConfigByID(ctx, config, pgtype.Text{String: u.id, Status: pgtype.Present})
	if err != nil {
		return otf.ConfigurationErrored, err
	}
	return otf.ConfigurationUploaded, nil
}
