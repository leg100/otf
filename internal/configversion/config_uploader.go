package configversion

import (
	"context"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type cvUploader struct {
	conn sql.Connection
	id   resource.ID
}

func (u *cvUploader) SetErrored(ctx context.Context) error {
	// TODO: add status timestamp
	_, err := q.UpdateConfigurationVersionErroredByID(ctx, u.conn, u.id)
	if err != nil {
		return err
	}
	return nil
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (ConfigurationStatus, error) {
	// TODO: add status timestamp
	_, err := q.UpdateConfigurationVersionConfigByID(ctx, u.conn, UpdateConfigurationVersionConfigByIDParams{
		ID:     u.id,
		Config: config,
	})
	if err != nil {
		return ConfigurationErrored, err
	}
	return ConfigurationUploaded, nil
}
