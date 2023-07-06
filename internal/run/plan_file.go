package run

import (
	"encoding/json"
)

const (
	CreateAction ChangeAction = "create"
	UpdateAction ChangeAction = "update"
	DeleteAction ChangeAction = "delete"
)

type (
	// PlanFile represents the schema of a plan file
	PlanFile struct {
		ResourceChanges []ResourceChange  `json:"resource_changes"`
		OutputChanges   map[string]Change `json:"output_changes"`
	}

	// PlanFileOptions are options for the plan file API
	PlanFileOptions struct {
		Format PlanFormat `schema:"format,required"`
	}

	// ResourceChange represents a proposed change to a resource in a plan file
	ResourceChange struct {
		Change Change
	}

	// Change represents the type of change being made
	Change struct {
		Actions []ChangeAction
	}

	ChangeAction string
)

// Summarize provides a tally of the types of changes proposed in the plan file.
func (pf *PlanFile) Summarize() (resource, output Report) {
	for _, rc := range pf.ResourceChanges {
		for _, action := range rc.Change.Actions {
			switch action {
			case CreateAction:
				resource.Additions++
			case UpdateAction:
				resource.Changes++
			case DeleteAction:
				resource.Destructions++
			}
		}
	}
	for _, rc := range pf.OutputChanges {
		for _, action := range rc.Actions {
			switch action {
			case CreateAction:
				output.Additions++
			case UpdateAction:
				output.Changes++
			case DeleteAction:
				output.Destructions++
			}
		}
	}

	return
}

// CompilePlanReports compiles reports of planned changes from a JSON
// representation of a plan file: one report for planned *resources*, and
// another for planned *outputs*.
func CompilePlanReports(planJSON []byte) (resources Report, outputs Report, err error) {
	planFile := PlanFile{}
	if err := json.Unmarshal(planJSON, &planFile); err != nil {
		return Report{}, Report{}, err
	}

	resources, outputs = planFile.Summarize()
	return resources, outputs, nil
}
