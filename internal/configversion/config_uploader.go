package configversion

import (
	"context"
)

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
