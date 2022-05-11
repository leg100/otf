package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var _ otf.RunStatusUpdater = (*runStatusUpdater)(nil)

type runStatusUpdater struct {
	q   *DBQuerier
	run *otf.Run
}

func newRunStatusUpdater(tx pgx.Tx, run *otf.Run) *runStatusUpdater {
	return &runStatusUpdater{
		q:   NewQuerier(tx),
		run: run,
	}
}

func (u *runStatusUpdater) UpdateRunStatus(ctx context.Context, run *otf.Run, status otf.RunStatus) error {
	result, err := u.q.UpdateRunStatus(ctx, otf.String(string(status)), &u.run.ID)
	if err != nil {
		return err
	}
	addResultToRun(u.run, result)

	ts, err := u.q.InsertRunStatusTimestamp(ctx, &u.run.ID, otf.String(string(status)))
	if err != nil {
		return err
	}
	u.run.StatusTimestamps = append(u.run.StatusTimestamps, convertRunStatusTimestamp(ts))

	return nil
}

func (u *runStatusUpdater) UpdatePlanStatus(ctx context.Context, plan *otf.Plan, status otf.PlanStatus) error {
	result, err := u.q.UpdatePlanStatus(ctx, otf.String(string(status)), &u.run.ID)
	if err != nil {
		return nil
	}
	addResultToPlan(plan, result)

	ts, err := u.q.InsertPlanStatusTimestamp(ctx, &u.run.ID, otf.String(string(status)))
	if err != nil {
		return nil
	}
	plan.StatusTimestamps = append(plan.StatusTimestamps, convertPlanStatusTimestamp(ts))

	return nil
}

func (u *runStatusUpdater) UpdateApplyStatus(ctx context.Context, status otf.ApplyStatus) error {
	result, err := u.q.UpdateApplyStatus(ctx, otf.String(string(status)), &u.run.ID)
	if err != nil {
		return nil
	}
	addResultToApply(u.run.Apply, result)

	ts, err := u.q.InsertApplyStatusTimestamp(ctx, &u.run.ID, otf.String(string(status)))
	if err != nil {
		return nil
	}
	u.run.Apply.StatusTimestamps = append(u.run.Apply.StatusTimestamps, convertApplyStatusTimestamp(ts))

	return nil
}
