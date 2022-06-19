package otf

import "context"

type ReportService interface {
	// CreatePlanReport produces a report of planned changes and persists it to
	// the database.
	CreatePlanReport(ctx context.Context, planID string) (ResourceReport, error)
	// CreateApplyReport produces a report of applied changes and persists it to
	// the database.
	CreateApplyReport(ctx context.Context, applyID string) (ResourceReport, error)
}
