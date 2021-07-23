package sqlite

import (
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// StateVersion models a row in a runs table.
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

func (r *StateVersion) ToDomain() *ots.StateVersion {
	domain := ots.StateVersion{
		ID:           r.ExternalID,
		Serial:       r.Serial,
		VCSCommitSHA: r.VCSCommitSHA,
		VCSCommitURL: r.VCSCommitURL,
		BlobID:       r.BlobID,
		Workspace:    r.Workspace.ToDomain(),
	}

	return &domain
}

// FromDomain updates run model fields with a run domain object's fields
func (r *StateVersion) FromDomain(domain *ots.StateVersion) {
	r.ExternalID = domain.ID
	r.Serial = domain.Serial
	r.VCSCommitSHA = domain.VCSCommitSHA
	r.VCSCommitURL = domain.VCSCommitURL
	r.BlobID = domain.BlobID
	r.Workspace.FromDomain(domain.Workspace)
}

func (l StateVersionList) ToDomain() (dl []*ots.StateVersion) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
