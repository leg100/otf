package sqlite

import (
	"github.com/leg100/otf"
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

func (model *StateVersionOutput) ToDomain() *otf.StateVersionOutput {
	domain := otf.StateVersionOutput{
		ID:        model.ExternalID,
		Name:      model.Name,
		Sensitive: model.Sensitive,
		Type:      model.Type,
		Value:     model.Value,
	}

	return &domain
}

// NewStateVersionOutputFromDomain constructs a model obj from a domain obj
func NewStateVersionOutputFromDomain(domain *otf.StateVersionOutput) *StateVersionOutput {
	return &StateVersionOutput{
		ExternalID: domain.ID,
		Name:       domain.Name,
		Sensitive:  domain.Sensitive,
		Type:       domain.Type,
		Value:      domain.Value,
	}
}

func (l StateVersionOutputList) ToDomain() (dl []*otf.StateVersionOutput) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
