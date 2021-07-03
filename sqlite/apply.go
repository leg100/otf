package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

var _ ots.ApplyService = (*ApplyService)(nil)

type ApplyModel struct {
	gorm.Model

	ExternalID string
}

type ApplyService struct {
	*gorm.DB
}

func NewApplyService(db *gorm.DB) *ApplyService {
	db.AutoMigrate(&ApplyModel{})

	return &ApplyService{
		DB: db,
	}
}

func NewApplyFromModel(model *ApplyModel) *tfe.Apply {
	return &tfe.Apply{
		ID: model.ExternalID,
	}
}

func (ApplyModel) TableName() string {
	return "plans"
}

func (s ApplyService) GetApply(id string) (*tfe.Apply, error) {
	model, err := getApplyByID(s.DB, id)
	if err != nil {
		return nil, err
	}
	return NewApplyFromModel(model), nil
}

func createApply(db *gorm.DB) (*ApplyModel, error) {
	model := ApplyModel{
		ExternalID: ots.NewApplyID(),
	}

	if result := db.Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func getApplyByID(db *gorm.DB, id string) (*ApplyModel, error) {
	var model ApplyModel

	if result := db.Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
