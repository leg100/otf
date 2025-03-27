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
	_, err := u.conn.Exec(ctx, `
UPDATE configuration_versions
SET
    status = 'errored'
WHERE configuration_version_id = $1
RETURNING configuration_version_id
`, u.id)
	return err
}

func (u *cvUploader) Upload(ctx context.Context, config []byte) (ConfigurationStatus, error) {
	// TODO: add status timestamp
	_, err := u.conn.Exec(ctx, `
UPDATE configuration_versions
SET
    config = $1,
    status = 'uploaded'
WHERE configuration_version_id = $2
`, config, u.id)
	if err != nil {
		return ConfigurationErrored, err
	}
	return ConfigurationUploaded, nil
}
