package sql

import (
	"context"

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
	// TODO: add status timestamp
	_, err := u.q.UpdateConfigurationVersionErroredByID(ctx, String(u.id))
	if err != nil {
		return err
	}
	return nil
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (otf.ConfigurationStatus, error) {
	// TODO: add status timestamp
	_, err := u.q.UpdateConfigurationVersionConfigByID(ctx, config, String(u.id))
	if err != nil {
		return otf.ConfigurationErrored, err
	}
	return otf.ConfigurationUploaded, nil
}
