package configversion

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

type cvUploader struct {
	q  *sqlc.Queries
	id resource.ID
}

func newConfigUploader(q *sqlc.Queries, id resource.ID) *cvUploader {
	return &cvUploader{
		q:  q,
		id: id,
	}
}

func (u *cvUploader) SetErrored(ctx context.Context) error {
	// TODO: add status timestamp
	_, err := u.q.UpdateConfigurationVersionErroredByID(ctx, u.id)
	if err != nil {
		return err
	}
	return nil
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (ConfigurationStatus, error) {
	// TODO: add status timestamp
	_, err := u.q.UpdateConfigurationVersionConfigByID(ctx, sqlc.UpdateConfigurationVersionConfigByIDParams{
		ID:     u.id,
		Config: config,
	})
	if err != nil {
		return ConfigurationErrored, err
	}
	return ConfigurationUploaded, nil
}
