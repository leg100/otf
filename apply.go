package otf

import (
	"fmt"

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

func (a *Apply) GetLogsBlobID() string {
	return a.LogsBlobID
}

func (a *Apply) Do(run *Run, exe *Executor) error {
	if err := exe.RunFunc(run.downloadPlanFile); err != nil {
		return err
	}

	if err := exe.RunCLI("sh", "-c", fmt.Sprintf("terraform apply -no-color %s | tee %s", PlanFilename, ApplyOutputFilename)); err != nil {
		return err
	}

	if err := exe.RunFunc(run.uploadState); err != nil {
		return err
	}

	return nil
}

// UpdateResources parses the output from terraform apply to determine the
// number and type of resource changes applied and updates the apply object
// accordingly.
func (a *Apply) UpdateResources(bs BlobStore) error {
	logs, err := bs.Get(a.LogsBlobID)
	if err != nil {
		return err
	}

	resources, err := parseApplyOutput(string(logs))
	if err != nil {
		return err
	}

	a.ResourceAdditions = resources.adds
	a.ResourceChanges = resources.changes
	a.ResourceDestructions = resources.deletions

	return nil
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
