package dto

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID                   string                 `jsonapi:"primary,plans"`
	HasChanges           bool                   `jsonapi:"attr,has-changes"`
	LogReadURL           string                 `jsonapi:"attr,log-read-url"`
	ResourceAdditions    int                    `jsonapi:"attr,resource-additions"`
	ResourceChanges      int                    `jsonapi:"attr,resource-changes"`
	ResourceDestructions int                    `jsonapi:"attr,resource-destructions"`
	Status               string                 `jsonapi:"attr,status"`
	StatusTimestamps     *PhaseStatusTimestamps `jsonapi:"attr,status-timestamps"`
}
