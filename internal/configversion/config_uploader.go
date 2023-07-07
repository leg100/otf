package configversion

import (
	"context"

	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type cvUploader struct {
	q  pggen.Querier
	id string
}

func newConfigUploader(q pggen.Querier, id string) *cvUploader {
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
	_, err := u.q.UpdateConfigurationVersionConfigByID(ctx, config, sql.String(u.id))
	if err != nil {
		return ConfigurationErrored, err
	}
	return ConfigurationUploaded, nil
}
