package ots

import (
	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

type ApplyService interface {
	Get(id string) (*Apply, error)
}

type Apply struct {
	ID string

	gorm.Model

	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               tfe.ApplyStatus
	StatusTimestamps     *tfe.ApplyStatusTimestamps

	// Logs is the blob ID for the log output from a terraform apply
	LogsBlobID string
}

// ApplyFinishOptions represents the options for finishing an apply.
type ApplyFinishOptions struct {
	// Type is a public field utilized by JSON:API to set the resource type via
	// the field tag.  It is not a user-defined value and does not need to be
	// set.  https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,applies"`

	ResourceAdditions    int `jsonapi:"attr,resource-additions"`
	ResourceChanges      int `jsonapi:"attr,resource-changes"`
	ResourceDestructions int `jsonapi:"attr,resource-destructions"`
}

func newApply() *Apply {
	return &Apply{
		ID:               GenerateID("apply"),
		StatusTimestamps: &tfe.ApplyStatusTimestamps{},
		LogsBlobID:       NewBlobID(),
	}
}

func (a *Apply) UpdateStatus(status tfe.ApplyStatus) {
	a.Status = status
	a.setTimestamp(status)
}

func (a *Apply) setTimestamp(status tfe.ApplyStatus) {
	switch status {
	case tfe.ApplyCanceled:
		a.StatusTimestamps.CanceledAt = TimeNow()
	case tfe.ApplyErrored:
		a.StatusTimestamps.ErroredAt = TimeNow()
	case tfe.ApplyFinished:
		a.StatusTimestamps.FinishedAt = TimeNow()
	case tfe.ApplyQueued:
		a.StatusTimestamps.QueuedAt = TimeNow()
	case tfe.ApplyRunning:
		a.StatusTimestamps.StartedAt = TimeNow()
	}
}
