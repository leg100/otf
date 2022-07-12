package dto

// Apply represents a Terraform Enterprise apply.
type Apply struct {
	ID               string                 `jsonapi:"primary,applies"`
	LogReadURL       string                 `jsonapi:"attr,log-read-url"`
	Status           string                 `jsonapi:"attr,status"`
	StatusTimestamps *PhaseStatusTimestamps `jsonapi:"attr,status-timestamps"`

	ResourceReport
}
