package agent

import (
	"github.com/go-logr/logr"
	"github.com/leg100/ots"
)

type NewApplyRunnerFn func(
	*ots.Run,
	ots.ConfigurationVersionService,
	ots.StateVersionService,
	ots.RunService,
	logr.Logger) *ots.Runner

func NewApplyRunner(run *ots.Run,
	cvs ots.ConfigurationVersionService,
	svs ots.StateVersionService,
	rs ots.RunService,
	runLogger ots.RunLogger,
	log logr.Logger) *ots.Runner {

	return ots.NewRunner(
		[]ots.Step{
			DownloadConfigStep(run, cvs),
			DeleteBackendStep,
			DownloadPlanFileStep(run, rs),
			DownloadStateStep(run, svs, log),
			UpdateApplyStatusStep(run, rs),
			InitStep,
			ApplyStep,
			UploadStateStep(run, svs),
			FinishApplyStep(run, rs, log),
		},
		runLogger,
		log,
		run.ID,
	)
}
