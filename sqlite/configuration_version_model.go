package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/otf"
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

	// BlobID is the ID of the binary object containing the configuration
	BlobID string

	// Configuration Version belongs to a Workspace
	WorkspaceID uint
	Workspace   *Workspace
}

// ConfigurationVersionList is a list of run models
type ConfigurationVersionList []ConfigurationVersion

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (model *ConfigurationVersion) Update(fn func(*otf.ConfigurationVersion) error) error {
	// model -> domain
	domain := model.ToDomain()

	// invoke user fn
	if err := fn(domain); err != nil {
		return err
	}

	// domain -> model
	model.FromDomain(domain)

	return nil
}

func (model *ConfigurationVersion) ToDomain() *otf.ConfigurationVersion {
	domain := otf.ConfigurationVersion{
		ID:            model.ExternalID,
		Model:         model.Model,
		AutoQueueRuns: model.AutoQueueRuns,
		Error:         model.Error,
		ErrorMessage:  model.ErrorMessage,
		Speculative:   model.Speculative,
		Source:        model.Source,
		Status:        model.Status,
		BlobID:        model.BlobID,
	}

	if model.Workspace != nil {
		domain.Workspace = model.Workspace.ToDomain()
	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (model *ConfigurationVersion) FromDomain(domain *otf.ConfigurationVersion) {
	model.ExternalID = domain.ID
	model.Model = domain.Model
	model.AutoQueueRuns = domain.AutoQueueRuns
	model.Error = domain.Error
	model.ErrorMessage = domain.ErrorMessage
	model.Speculative = domain.Speculative
	model.Source = domain.Source
	model.Status = domain.Status
	model.BlobID = domain.BlobID

	if domain.Workspace != nil {
		model.Workspace = &Workspace{}
		model.Workspace.FromDomain(domain.Workspace)
		model.WorkspaceID = domain.Workspace.Model.ID
	}
}

func (l ConfigurationVersionList) ToDomain() (dl []*otf.ConfigurationVersion) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
