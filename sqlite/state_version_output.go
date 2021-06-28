package sqlite

import (
	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

var _ ots.StateVersionOutputService = (*StateVersionOutputService)(nil)

type StateVersionOutputModel struct {
	gorm.Model

	ExternalID string
	Name       string
	Sensitive  bool
	Type       string
	Value      string

	StateVersionID uint
}

type StateVersionOutputService struct {
	*gorm.DB
}

func NewStateVersionOutputService(db *gorm.DB) *StateVersionOutputService {
	db.AutoMigrate(&StateVersionOutputModel{})

	return &StateVersionOutputService{
		DB: db,
	}
}

func NewStateVersionOutputFromModel(model *StateVersionOutputModel) *tfe.StateVersionOutput {
	return &tfe.StateVersionOutput{
		ID:        model.ExternalID,
		Name:      model.Name,
		Sensitive: model.Sensitive,
		Type:      model.Type,
		Value:     model.Value,
	}
}

func (StateVersionOutputModel) TableName() string {
	return "state_version_outputs"
}

func (s StateVersionOutputService) GetStateVersionOutput(id string) (*tfe.StateVersionOutput, error) {
	var model StateVersionOutputModel

	if result := s.DB.Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewStateVersionOutputFromModel(&model), nil
}
