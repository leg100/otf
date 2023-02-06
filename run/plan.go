package run

const (
	LocalStateFilename = "terraform.tfstate"
	PlanFilename       = "plan.out"
	JSONPlanFilename   = "plan.out.json"
	LockFilename       = ".terraform.lock.hcl"
)

// Plan is the plan phase of a run
type Plan struct {
	// report of planned resource changes
	*ResourceReport

	runID string
	*phaseStatus
}

func (p *Plan) ID() string       { return p.runID }
func (p *Plan) Phase() PhaseType { return PlanPhase }

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	return p.ResourceReport != nil && p.ResourceReport.HasChanges()
}

func newPlan(run *Run) *Plan {
	return &Plan{
		runID:       run.id,
		phaseStatus: newPhaseStatus(),
	}
}
