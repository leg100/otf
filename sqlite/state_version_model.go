package sqlite

import (
	"github.com/leg100/otf"
	"gorm.io/gorm"
)

// StateVersion models a row in a state versions table.
type StateVersion struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	Serial       int64
	VCSCommitSHA string
	VCSCommitURL string

	BlobID string

	// State version belongs to a workspace
	WorkspaceID uint
	Workspace   *Workspace

	// Run that created this state version. Optional.
	// Run     *Run

	// StateVersion has many StateVersionOutput
	Outputs []*StateVersionOutput
}

type StateVersionList []StateVersion

func (model *StateVersion) ToDomain() *otf.StateVersion {
	domain := otf.StateVersion{
		ID:           model.ExternalID,
		Model:        model.Model,
		Serial:       model.Serial,
		VCSCommitSHA: model.VCSCommitSHA,
		VCSCommitURL: model.VCSCommitURL,
		BlobID:       model.BlobID,
	}

	for _, out := range model.Outputs {
		domain.Outputs = append(domain.Outputs, out.ToDomain())
	}

	if model.Workspace != nil {
		domain.Workspace = model.Workspace.ToDomain()
	}

	return &domain
}

// FromDomain updates state version model fields with a state version domain
// object's fields
func (model *StateVersion) FromDomain(domain *otf.StateVersion) {
	model.Model = domain.Model
	model.ExternalID = domain.ID
	model.Serial = domain.Serial
	model.VCSCommitSHA = domain.VCSCommitSHA
	model.VCSCommitURL = domain.VCSCommitURL
	model.BlobID = domain.BlobID

	for _, out := range domain.Outputs {
		model.Outputs = append(model.Outputs, NewStateVersionOutputFromDomain(out))
	}

	if domain.Workspace != nil {
		model.Workspace = &Workspace{}
		model.Workspace.FromDomain(domain.Workspace)
		model.WorkspaceID = domain.Workspace.Model.ID
	}
}

func (l StateVersionList) ToDomain() (dl []*otf.StateVersion) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
