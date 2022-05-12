package otf

import (
	"context"
)

var _ RunStatusUpdater = (*fakeUpdater)(nil)

type fakeUpdater struct{}

func (u *fakeUpdater) UpdateApplyStatus(ctx context.Context, status ApplyStatus) (*Apply, error) {
	return &Apply{Status: status}, nil
}

func (u *fakeUpdater) UpdatePlanStatus(ctx context.Context, status PlanStatus) (*Plan, error) {
	return &Plan{Status: status}, nil
}

func (u *fakeUpdater) UpdateRunStatus(ctx context.Context, status RunStatus) (*Run, error) {
	return &Run{Status: status}, nil
}
