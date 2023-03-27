package configversion

import (
	"context"

	"github.com/leg100/otf/sql"
)

type cvUploader struct {
	db *pgdb
	id string
}

func newConfigUploader(db *pgdb, id string) *cvUploader {
	return &cvUploader{
		db: db,
		id: id,
	}
}

func (u *cvUploader) SetErrored(ctx context.Context) error {
	// TODO: add status timestamp
	_, err := u.db.UpdateConfigurationVersionErroredByID(ctx, sql.String(u.id))
	if err != nil {
		return err
	}
	return nil
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (ConfigurationStatus, error) {
	// TODO: add status timestamp
	_, err := u.db.UpdateConfigurationVersionConfigByID(ctx, config, sql.String(u.id))
	if err != nil {
		return ConfigurationErrored, err
	}
	return ConfigurationUploaded, nil
}
