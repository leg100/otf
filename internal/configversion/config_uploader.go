package configversion

import (
	"context"

	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type cvUploader struct {
	q  *sqlc.Queries
	id string
}

func newConfigUploader(q *sqlc.Queries, id string) *cvUploader {
	return &cvUploader{
		q:  q,
		id: id,
	}
}

func (u *cvUploader) SetErrored(ctx context.Context) error {
	// TODO: add status timestamp
	_, err := u.q.UpdateConfigurationVersionErroredByID(ctx, sql.String(u.id))
	if err != nil {
		return err
	}
	return nil
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (ConfigurationStatus, error) {
	// TODO: add status timestamp
	_, err := u.q.UpdateConfigurationVersionConfigByID(ctx, sqlc.UpdateConfigurationVersionConfigByIDParams{
		ID:     sql.String(u.id),
		Config: config,
	})
	if err != nil {
		return ConfigurationErrored, err
	}
	return ConfigurationUploaded, nil
}
