package sql

import (
	"context"

	"github.com/leg100/otf"
)

type cvUploader struct {
	db *DB
	id string
}

func newConfigUploader(db *DB, id string) *cvUploader {
	return &cvUploader{
		db: db,
		id: id,
	}
}

func (u *cvUploader) SetErrored(ctx context.Context) error {
	_, err := u.db.UpdateConfigurationVersionErroredByID(ctx, String(u.id))
	if err != nil {
		return err
	}
	return nil
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (otf.ConfigurationStatus, error) {
	_, err := u.db.UpdateConfigurationVersionConfigByID(ctx, config, String(u.id))
	if err != nil {
		return otf.ConfigurationErrored, err
	}
	return otf.ConfigurationUploaded, nil
}
