package sqlite

import (
	"fmt"
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
)

var _ ots.PlanService = (*PlanService)(nil)

type PlanModel struct {
	gorm.Model

	ExternalID string

	ResourceAdditions    int
	ResourceChanges      int
	ResourceDestructions int
	Status               string

	Logs []byte

	tfe.PlanStatusTimestamps `gorm:"embedded,embeddedPrefix:timestamp_"`
}

type PlanService struct {
	*gorm.DB
}

func NewPlanService(db *gorm.DB) *PlanService {
	db.AutoMigrate(&PlanModel{})

	return &PlanService{
		DB: db,
	}
}

func NewPlanFromModel(model *PlanModel) *tfe.Plan {
	return &tfe.Plan{
		ID:                   model.ExternalID,
		ResourceAdditions:    model.ResourceAdditions,
		ResourceChanges:      model.ResourceChanges,
		ResourceDestructions: model.ResourceDestructions,
		Status:               tfe.PlanStatus(model.Status),
		StatusTimestamps:     &model.PlanStatusTimestamps,
	}
}

func (PlanModel) TableName() string {
	return "plans"
}

func (s PlanService) GetPlan(id string) (*tfe.Plan, error) {
	model, err := getPlanByID(s.DB, id)
	if err != nil {
		return nil, err
	}
	return NewPlanFromModel(model), nil
}

func (s PlanService) UpdatePlanStatus(id string, status tfe.PlanStatus) (*tfe.Plan, error) {
	model, err := getPlanByID(s.DB, id)
	if err != nil {
		return nil, err
	}

	update := make(map[string]interface{})

	switch status {
	case tfe.PlanRunning:
		update["status"] = string(tfe.PlanRunning)
	case tfe.PlanErrored:
		update["status"] = string(tfe.PlanErrored)
		update["timestamp_errored_at"] = time.Now
	}

	if result := s.DB.Model(&model).Updates(update); result.Error != nil {
		return nil, result.Error
	}

	return NewPlanFromModel(model), nil
}

func (s PlanService) FinishPlan(id string, opts *ots.PlanFinishOptions) (*tfe.Plan, error) {
	model, err := getPlanByID(s.DB, id)
	if err != nil {
		return nil, err
	}

	update := map[string]interface{}{
		"status":                string(tfe.PlanFinished),
		"timestamp_finished_at": time.Now(),
		"resource_additions":    opts.ResourceAdditions,
		"resource_changes":      opts.ResourceChanges,
		"resource_destructions": opts.ResourceDestructions,
	}

	if result := s.DB.Model(&model).Updates(update); result.Error != nil {
		return nil, result.Error
	}

	return NewPlanFromModel(model), nil
}

func (s PlanService) GetPlanLogs(id string, opts ots.PlanLogOptions) ([]byte, error) {
	model, err := getPlanByID(s.DB, id)
	if err != nil {
		return nil, err
	}

	if opts.Offset > len(model.Logs) {
		return nil, fmt.Errorf("offset too high")
	}
	if opts.Limit > ots.MaxPlanLogsLimit {
		opts.Limit = ots.MaxPlanLogsLimit
	}
	if (opts.Offset + opts.Limit) > len(model.Logs) {
		opts.Limit = len(model.Logs) - opts.Offset
	}

	return model.Logs[opts.Offset:opts.Limit], nil
}

func (s PlanService) UploadPlanLogs(id string, logs []byte) error {
	model, err := getPlanByID(s.DB, id)
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"logs": logs,
	}

	if result := s.DB.Model(&model).Updates(update); result.Error != nil {
		return result.Error
	}

	return nil
}

func createPlan(db *gorm.DB) (*PlanModel, error) {
	model := PlanModel{
		ExternalID: ots.NewPlanID(),
	}

	if result := db.Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func getPlanByID(db *gorm.DB, id string) (*PlanModel, error) {
	var model PlanModel

	if result := db.Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
