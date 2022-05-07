package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

type runStatusUpdater struct {
	q  *DBQuerier
	id string
}

func newRunStatusUpdater(tx pgx.Tx, id string) *runStatusUpdater {
	return &runStatusUpdater{
		q:  NewQuerier(tx),
		id: id,
	}
}

func (u *runStatusUpdater) UpdateRunStatus(ctx context.Context, status otf.RunStatus) (otf.RunStatusTimestamp, error) {
	_, err := u.q.UpdateRunStatus(ctx, otf.String(string(status)), &u.id)
	if err != nil {
		return otf.RunStatusTimestamp{}, nil
	}

	ts, err := u.q.InsertRunStatusTimestamp(ctx, &u.id, otf.String(string(status)))
	if err != nil {
		return otf.RunStatusTimestamp{}, nil
	}

	return otf.RunStatusTimestamp{
		Status:    otf.RunStatus(*ts.Status),
		Timestamp: ts.Timestamp,
	}, nil
}

func (u *runStatusUpdater) UpdatePlanStatus(ctx context.Context, status otf.PlanStatus) (otf.PlanStatusTimestamp, error) {
	_, err := u.q.UpdatePlanStatus(ctx, otf.String(string(status)), &u.id)
	if err != nil {
		return otf.PlanStatusTimestamp{}, nil
	}

	ts, err := u.q.InsertPlanStatusTimestamp(ctx, &u.id, otf.String(string(status)))
	if err != nil {
		return otf.PlanStatusTimestamp{}, nil
	}

	return otf.PlanStatusTimestamp{
		Status:    otf.PlanStatus(*ts.Status),
		Timestamp: ts.Timestamp,
	}, nil
}

func (u *runStatusUpdater) UpdateApplyStatus(ctx context.Context, status otf.ApplyStatus) (otf.ApplyStatusTimestamp, error) {
	_, err := u.q.UpdateApplyStatus(ctx, otf.String(string(status)), &u.id)
	if err != nil {
		return otf.ApplyStatusTimestamp{}, nil
	}

	ts, err := u.q.InsertApplyStatusTimestamp(ctx, &u.id, otf.String(string(status)))
	if err != nil {
		return otf.ApplyStatusTimestamp{}, nil
	}

	return otf.ApplyStatusTimestamp{
		Status:    otf.ApplyStatus(*ts.Status),
		Timestamp: ts.Timestamp,
	}, nil
}
