package ots

import (
	"fmt"
	"time"

	tfe "github.com/leg100/go-tfe"
)

type ApplyService interface {
	Get(id string) (*Apply, error)
}

type Apply struct {
	ExternalID string `gorm:"uniqueIndex"`
	InternalID uint   `gorm:"primaryKey;column:id"`

	LogReadURL           string
	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               tfe.ApplyStatus
	StatusTimestamps     *tfe.ApplyStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Logs []byte

	RunID uint
}

func (a *Apply) DTO() interface{} {
	return &tfe.Apply{
		ID:                   a.ExternalID,
		LogReadURL:           a.LogReadURL,
		ResourceAdditions:    a.ResourceAdditions,
		ResourceChanges:      a.ResourceChanges,
		ResourceDestructions: a.ResourceDestructions,
		Status:               a.Status,
		StatusTimestamps:     a.StatusTimestamps,
	}
}

func NewApplyID() string {
	return fmt.Sprintf("apply-%s", GenerateRandomString(16))
}

func newApply() *Apply {
	return &Apply{
		ExternalID: NewApplyID(),
	}
}

// UpdateStatus updates the status of the apply. It'll also update the
// appropriate timestamp and set any other appropriate fields for the given
// status.
func (a *Apply) UpdateStatus(status tfe.ApplyStatus) {
	// Copy timestamps from apply
	timestamps := &tfe.ApplyStatusTimestamps{}
	if a.StatusTimestamps != nil {
		timestamps = a.StatusTimestamps
	}

	switch status {
	case tfe.ApplyRunning:
		timestamps.StartedAt = time.Now()
	case tfe.ApplyQueued:
		timestamps.QueuedAt = time.Now()
	case tfe.ApplyCanceled:
		timestamps.CanceledAt = time.Now()
	case tfe.ApplyErrored:
		timestamps.ErroredAt = time.Now()
	default:
		// Don't set a timestamp
		return
	}

	a.Status = status

	// Set timestamps on apply
	a.StatusTimestamps = timestamps
}
