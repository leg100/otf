package agent

import (
	"github.com/go-logr/logr"
	"github.com/leg100/ots"
)

type NewPlanRunnerFn func(
	*ots.Run,
	ots.ConfigurationVersionService,
	ots.StateVersionService,
	ots.RunService,
	logr.Logger) *ots.Runner

func NewPlanRunner(run *ots.Run,
	cvs ots.ConfigurationVersionService,
	svs ots.StateVersionService,
	rs ots.RunService,
	log logr.Logger) *ots.Runner {

	return ots.NewRunner(
		[]ots.Step{
			DownloadConfigStep(run, cvs),
			DeleteBackendStep,
			DownloadStateStep(run, svs, log),
			UpdatePlanStatusStep(run, rs),
			InitStep,
			PlanStep,
			JSONPlanStep,
			UploadPlanStep(run, rs),
			UploadJSONPlanStep(run, rs),
			FinishPlanStep(run, rs, log),
		},
	)
}
