package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// ConfigurationVersion models a row in a configuration_versions table.
type ConfigurationVersion struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	AutoQueueRuns    bool
	Error            string
	ErrorMessage     string
	Source           tfe.ConfigurationSource
	Speculative      bool
	Status           tfe.ConfigurationStatus
	StatusTimestamps *tfe.CVStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	Configuration []byte `gorm:"-"`
	BlobID        string

	// Configuration Version belongs to a Workspace
	WorkspaceID uint
	Workspace   *Workspace
}

// ConfigurationVersionList is a list of run models
type ConfigurationVersionList []ConfigurationVersion

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (cv *ConfigurationVersion) Update(fn func(*ots.ConfigurationVersion) error) error {
	// model -> domain
	domain := cv.ToDomain()

	// invoke user fn
	if err := fn(domain); err != nil {
		return err
	}

	// domain -> model
	cv.FromDomain(domain)

	return nil
}

func (cv *ConfigurationVersion) ToDomain() *ots.ConfigurationVersion {
	domain := ots.ConfigurationVersion{
		ID:            cv.ExternalID,
		Model:         cv.Model,
		AutoQueueRuns: cv.AutoQueueRuns,
		Error:         cv.Error,
		ErrorMessage:  cv.ErrorMessage,
		Speculative:   cv.Speculative,
		Source:        cv.Source,
		Status:        cv.Status,
		Workspace:     cv.Workspace.ToDomain(),
	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (cv *ConfigurationVersion) FromDomain(domain *ots.ConfigurationVersion) {
	cv.ExternalID = domain.ID
	cv.Model = domain.Model
	cv.AutoQueueRuns = domain.AutoQueueRuns
	cv.Error = domain.Error
	cv.ErrorMessage = domain.ErrorMessage
	cv.Speculative = domain.Speculative
	cv.Source = domain.Source
	cv.Status = domain.Status

	cv.Workspace = &Workspace{}
	cv.Workspace.FromDomain(domain.Workspace)
	cv.WorkspaceID = domain.Workspace.Model.ID
}

func (l ConfigurationVersionList) ToDomain() (dl []*ots.ConfigurationVersion) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
