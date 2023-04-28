package jsonapi

// Apply is a terraform apply
type Apply struct {
	ID               string                 `jsonapi:"primary,applies"`
	LogReadURL       string                 `jsonapi:"attribute" json:"log-read-url"`
	Status           string                 `jsonapi:"attribute" json:"status"`
	StatusTimestamps *PhaseStatusTimestamps `jsonapi:"attribute" json:"status-timestamps"`

	ResourceReport
}
