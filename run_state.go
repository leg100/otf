package otf

const (
	RunApplied            RunStatus = "applied"
	RunApplyQueued        RunStatus = "apply_queued"
	RunApplying           RunStatus = "applying"
	RunCanceled           RunStatus = "canceled"
	RunForceCanceled      RunStatus = "force_canceled"
	RunConfirmed          RunStatus = "confirmed"
	RunDiscarded          RunStatus = "discarded"
	RunErrored            RunStatus = "errored"
	RunPending            RunStatus = "pending"
	RunPlanQueued         RunStatus = "plan_queued"
	RunPlanned            RunStatus = "planned"
	RunPlannedAndFinished RunStatus = "planned_and_finished"
	RunPlanning           RunStatus = "planning"
)

type RunStatus string

func (r RunStatus) String() string { return string(r) }

type runState interface {
	String() string
	Start() error
	Finish(RunService) (*ResourceReport, error)
	Discard() error
	Apply() error
	Cancelable() bool
	Confirmable() bool
	Discardable() bool
	Done() bool
}
