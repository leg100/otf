package sqlite

import (
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

// StateVersionOutput models a row in a state versions table.
type StateVersionOutput struct {
	gorm.Model

	ExternalID string `gorm:"uniqueIndex"`

	Name      string
	Sensitive bool
	Type      string
	Value     string

	// StateVersionOutput belongs to State Version
	StateVersionID uint
}

// StateVersionOutputList is a list of run models
type StateVersionOutputList []StateVersionOutput

func (svo *StateVersionOutput) ToDomain() *ots.StateVersionOutput {
	domain := ots.StateVersionOutput{
		ID:        svo.ExternalID,
		Name:      svo.Name,
		Sensitive: svo.Sensitive,
		Type:      svo.Type,
		Value:     svo.Value,
	}

	return &domain
}

// NewStateVersionOutputFromDomain constructs a model obj from a domain obj
func NewStateVersionOutputFromDomain(domain *ots.StateVersionOutput) *StateVersionOutput {
	return &StateVersionOutput{
		ExternalID: domain.ID,
		Name:       domain.Name,
		Sensitive:  domain.Sensitive,
		Type:       domain.Type,
		Value:      domain.Value,
	}
}

func (l StateVersionOutputList) ToDomain() (dl []*ots.StateVersionOutput) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
