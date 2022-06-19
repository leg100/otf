package otf

import "context"

type fakeReportService struct {
	report ResourceReport
}

func (f *fakeReportService) CreatePlanReport(ctx context.Context, planID string) (ResourceReport, error) {
	return f.report, nil
}

func (f *fakeReportService) CreateApplyReport(ctx context.Context, applyID string) (ResourceReport, error) {
	return f.report, nil
}
