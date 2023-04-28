package jsonapi

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID               string                 `jsonapi:"primary,plans"`
	HasChanges       bool                   `jsonapi:"attribute" json:"has-changes"`
	LogReadURL       string                 `jsonapi:"attribute" json:"log-read-url"`
	Status           string                 `jsonapi:"attribute" json:"status"`
	StatusTimestamps *PhaseStatusTimestamps `jsonapi:"attribute" json:"status-timestamps"`

	ResourceReport
}

type ResourceReport struct {
	Additions    *int `json:"resource-additions"`
	Changes      *int `json:"resource-changes"`
	Destructions *int `json:"resource-destructions"`
}
