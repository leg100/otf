package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

type cvUpdater struct {
	q  *DBQuerier
	id string
}

func newConfigurationVersionUpdater(tx pgx.Tx, id string) *cvUpdater {
	return &cvUpdater{
		q:  NewQuerier(tx),
		id: id,
	}
}

func (u *cvUpdater) UpdateStatus(ctx context.Context, status otf.ConfigurationStatus) (otf.ConfigurationVersionStatusTimestamp, error) {
	_, err := u.q.UpdateConfigurationVersionStatus(ctx, otf.String(string(status)), &u.id)
	if err != nil {
		return otf.ConfigurationVersionStatusTimestamp{}, nil
	}

	ts, err := u.q.InsertConfigurationVersionStatusTimestamp(ctx, &u.id, otf.String(string(status)))
	if err != nil {
		return otf.ConfigurationVersionStatusTimestamp{}, nil
	}

	return otf.ConfigurationVersionStatusTimestamp{
		Status:    otf.ConfigurationStatus(*ts.Status),
		Timestamp: ts.Timestamp,
	}, nil
}

func (u *cvUpdater) SaveConfig(ctx context.Context, config []byte) error {
	_, err := u.q.UpdateConfigurationVersionConfig(ctx, config, &u.id)
	return err
}
