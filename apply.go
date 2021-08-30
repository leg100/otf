package ots

import (
	tfe "github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

const (
	MaxApplyLogsLimit = 65536
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

	Logs []byte
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

type ApplyLogOptions struct {
	// The maximum number of bytes of logs to return to the client
	Limit int `schema:"limit"`

	// The start position in the logs from which to send to the client
	Offset int `schema:"offset"`
}

func newApply() *Apply {
	return &Apply{
		ID:               GenerateID("apply"),
		StatusTimestamps: &tfe.ApplyStatusTimestamps{},
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
