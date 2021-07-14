package ots

import (
	"fmt"
	"time"

	"github.com/leg100/go-tfe"
	"gorm.io/gorm"
)

type StateVersionOutput struct {
	ExternalID string `gorm:"uniqueIndex"`
	InternalID uint   `gorm:"primaryKey;column:id"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name      string
	Sensitive bool
	Type      string
	Value     string

	// StateVersionOutput belongs to State Version
	StateVersionID uint
}

type StateVersionOutputList []*StateVersionOutput

func (svo *StateVersionOutput) DTO() interface{} {
	return &tfe.StateVersionOutput{
		ID:        svo.ExternalID,
		Name:      svo.Name,
		Sensitive: svo.Sensitive,
		Type:      svo.Type,
		Value:     svo.Value,
	}
}

func (svol StateVersionOutputList) DTO() interface{} {
	var dtol []*tfe.StateVersionOutput
	for _, item := range svol {
		dtol = append(dtol, item.DTO().(*tfe.StateVersionOutput))
	}

	return dtol
}

func NewStateVersionOutputID() string {
	return fmt.Sprintf("wsout-%s", GenerateRandomString(16))
}
