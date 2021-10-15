package otf

const (
	CreateAction ChangeAction = "create"
	UpdateAction ChangeAction = "update"
	DeleteAction ChangeAction = "delete"
)

// PlanFileOptions represents the options for retrieving the plan file for a
// run.
type PlanFileOptions struct {
	// Format of plan file. Valid values are json and binary.
	Format string `schema:"format"`
}

// PlanFile represents the schema of a plan file
type PlanFile struct {
	ResourcesChanges []ResourceChange `json:"resource_changes"`
}

// ResourceChange represents a proposed change to a resource in a plan file
type ResourceChange struct {
	Change Change
}

// Change represents the type of change being made
type Change struct {
	Actions []ChangeAction
}

type ChangeAction string

// Changes provides a tally of the types of changes proposed in the plan file.
func (pf *PlanFile) Changes() (adds int, updates int, deletes int) {
	for _, rc := range pf.ResourcesChanges {
		for _, action := range rc.Change.Actions {
			switch action {
			case CreateAction:
				adds++
			case UpdateAction:
				updates++
			case DeleteAction:
				deletes++
			}
		}
	}

	return
}
