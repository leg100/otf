package otf

import (
	"context"
	"time"
)

var _ RunStatusUpdater = (*fakeUpdater)(nil)

type fakeUpdater struct{}

func (u *fakeUpdater) UpdateApplyStatus(ctx context.Context, status ApplyStatus) (ApplyStatusTimestamp, error) {
	return ApplyStatusTimestamp{Status: status, Timestamp: time.Now()}, nil
}

func (u *fakeUpdater) UpdatePlanStatus(ctx context.Context, status PlanStatus) (PlanStatusTimestamp, error) {
	return PlanStatusTimestamp{Status: status, Timestamp: time.Now()}, nil
}

func (u *fakeUpdater) UpdateRunStatus(ctx context.Context, status RunStatus) (RunStatusTimestamp, error) {
	return RunStatusTimestamp{Status: status, Timestamp: time.Now()}, nil
}
