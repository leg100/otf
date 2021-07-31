package sqlite

import (
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// StateVersion models a row in a state versions table.
type StateVersion struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	Serial       int64
	VCSCommitSHA string
	VCSCommitURL string

	State  string
	BlobID string

	// State version belongs to a workspace
	WorkspaceID uint
	Workspace   *Workspace

	// Run that created this state version. Optional.
	// Run     *Run

	Outputs []*StateVersionOutput
}

// StateVersionList is a list of run models
type StateVersionList []StateVersion

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (r *StateVersion) Update(fn func(*ots.StateVersion) error) error {
	// model -> domain
	domain := r.ToDomain()

	// invoke user fn
	if err := fn(domain); err != nil {
		return err
	}

	// domain -> model
	r.FromDomain(domain)

	return nil
}

func (model *StateVersion) ToDomain() *ots.StateVersion {
	domain := ots.StateVersion{
		ID:           model.ExternalID,
		Serial:       model.Serial,
		VCSCommitSHA: model.VCSCommitSHA,
		VCSCommitURL: model.VCSCommitURL,
		BlobID:       model.BlobID,
	}

	if model.Workspace != nil {
		domain.Workspace = model.Workspace.ToDomain()
	}

	return &domain
}

// FromDomain updates state version model fields with a state version domain
// object's fields
func (model *StateVersion) FromDomain(domain *ots.StateVersion) {
	model.ExternalID = domain.ID
	model.Serial = domain.Serial
	model.VCSCommitSHA = domain.VCSCommitSHA
	model.VCSCommitURL = domain.VCSCommitURL
	model.BlobID = domain.BlobID

	if domain.Workspace != nil {
		model.Workspace = &Workspace{}
		model.Workspace.FromDomain(domain.Workspace)
		model.WorkspaceID = domain.Workspace.Model.ID
	}
}

func (l StateVersionList) ToDomain() (dl []*ots.StateVersion) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
