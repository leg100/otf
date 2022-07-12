package dto

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID               string                 `jsonapi:"primary,plans"`
	HasChanges       bool                   `jsonapi:"attr,has-changes"`
	LogReadURL       string                 `jsonapi:"attr,log-read-url"`
	Status           string                 `jsonapi:"attr,status"`
	StatusTimestamps *PhaseStatusTimestamps `jsonapi:"attr,status-timestamps"`

	ResourceReport
}

type ResourceReport struct {
	Additions    *int `json:"resource-additions"`
	Changes      *int `json:"resource-changes"`
	Destructions *int `json:"resource-destructions"`
}
