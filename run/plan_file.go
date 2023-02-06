package run

import (
	"encoding/json"
	"fmt"
)

const (
	CreateAction ChangeAction = "create"
	UpdateAction ChangeAction = "update"
	DeleteAction ChangeAction = "delete"

	// PlanFormatBinary is the binary representation of the plan file
	PlanFormatBinary = "bin"
	// PlanFormatJSON is the JSON representation of the plan file
	PlanFormatJSON = "json"
)

// PlanFormat is the format of the plan file
type PlanFormat string

func (f PlanFormat) CacheKey(id string) string {
	return fmt.Sprintf("%s.%s", id, f)
}

func (f PlanFormat) SQLColumn() string {
	return fmt.Sprintf("plan_%s", f)
}

// PlanFile represents the schema of a plan file
type PlanFile struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
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
func (pf *PlanFile) Changes() (tally ResourceReport) {
	for _, rc := range pf.ResourceChanges {
		for _, action := range rc.Change.Actions {
			switch action {
			case CreateAction:
				tally.Additions++
			case UpdateAction:
				tally.Changes++
			case DeleteAction:
				tally.Destructions++
			}
		}
	}

	return
}

// CompilePlanReport compiles a report of planned changes from a JSON
// representation of a plan file.
func CompilePlanReport(planJSON []byte) (ResourceReport, error) {
	planFile := PlanFile{}
	if err := json.Unmarshal(planJSON, &planFile); err != nil {
		return ResourceReport{}, err
	}

	// Parse plan output
	return planFile.Changes(), nil
}
