package otf

import "fmt"

const (
	CreateAction ChangeAction = "create"
	UpdateAction ChangeAction = "update"
	DeleteAction ChangeAction = "delete"

	// PlanFormatBinary is the binary representation of the plan file
	PlanFormatBinary = "bin"
	// PlanFormatJSON is the JSON representation of the plan file
	PlanFormatJSON = "json"
)

// PlanFile represents the schema of a plan file
type PlanFile struct {
	ResourcesChanges []ResourceChange `json:"resource_changes"`
}

// PlanFormat is the format of the plan file
type PlanFormat string

func (f *PlanFormat) CacheKey(id string) string {
	return fmt.Sprintf("%s.%v", id, f)
}

func (f *PlanFormat) SQLColumn() string {
	return fmt.Sprintf("plan_%v", f)
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
func (pf *PlanFile) Changes() (tally Resources) {
	for _, rc := range pf.ResourcesChanges {
		for _, action := range rc.Change.Actions {
			switch action {
			case CreateAction:
				tally.ResourceAdditions++
			case UpdateAction:
				tally.ResourceChanges++
			case DeleteAction:
				tally.ResourceDestructions++
			}
		}
	}

	return
}
