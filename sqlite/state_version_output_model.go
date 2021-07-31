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

// Update updates the model with the supplied fn. The fn operates on the domain
// obj, so Update handles converting to and from a domain obj.
func (r *StateVersionOutput) Update(fn func(*ots.StateVersionOutput) error) error {
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

// FromDomain updates run model fields with a run domain object's fields
func (r *StateVersionOutput) FromDomain(domain *ots.StateVersionOutput) {
	r.ExternalID = domain.ID
	r.Name = domain.Name
	r.Sensitive = domain.Sensitive
	r.Type = domain.Type
	r.Value = domain.Value
}

func (l StateVersionOutputList) ToDomain() (dl []*ots.StateVersionOutput) {
	for _, i := range l {
		dl = append(dl, i.ToDomain())
	}
	return
}
