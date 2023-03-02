package otf

// Plan is the plan phase of a run
type Plan struct {
	*ResourceReport // report of planned resource changes

	*Phase
}

// HasChanges determines whether plan has any changes (adds/changes/deletions).
func (p *Plan) HasChanges() bool {
	return p.ResourceReport != nil && p.ResourceReport.HasChanges()
}
