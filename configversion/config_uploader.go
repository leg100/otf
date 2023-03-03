package configversion

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

type cvUploader struct {
	db *db
	id string
}

func newConfigUploader(db *db, id string) *cvUploader {
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

func (u *cvUploader) Upload(ctx context.Context, config []byte) (otf.ConfigurationStatus, error) {
	// TODO: add status timestamp
	_, err := u.db.UpdateConfigurationVersionConfigByID(ctx, config, sql.String(u.id))
	if err != nil {
		return otf.ConfigurationErrored, err
	}
	return otf.ConfigurationUploaded, nil
}
