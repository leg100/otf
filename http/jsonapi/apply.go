package jsonapi

// Apply is a terraform apply
type Apply struct {
	ID               string                 `jsonapi:"primary,applies"`
	LogReadURL       string                 `jsonapi:"attr,log-read-url"`
	Status           string                 `jsonapi:"attr,status"`
	StatusTimestamps *PhaseStatusTimestamps `jsonapi:"attr,status-timestamps"`

	ResourceReport
}
