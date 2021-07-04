package sqlite

import (
	"log"
	"strings"
	"time"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ots.RunService = (*RunService)(nil)

type RunModel struct {
	gorm.Model

	ExternalID string
	Refresh    bool
	Message    string
	Status     tfe.RunStatus

	ReplaceAddrs string
	TargetAddrs  string

	ForceCancelAvailableAt time.Time

	tfe.RunActions          `gorm:"embedded;embeddedPrefix:action_"`
	tfe.RunPermissions      `gorm:"embedded;embeddedPrefix:permission_"`
	tfe.RunStatusTimestamps `gorm:"embedded;embeddedPrefix:timestamp_"`

	WorkspaceID uint
	Workspace   WorkspaceModel

	ConfigurationVersionID uint
	ConfigurationVersion   ConfigurationVersionModel

	PlanID uint
	Plan   PlanModel

	ApplyID uint
	Apply   ApplyModel
}

type RunService struct {
	*gorm.DB
}

func NewRunService(db *gorm.DB) *RunService {
	db.AutoMigrate(&RunModel{})

	return &RunService{
		DB: db,
	}
}

func NewRunListFromModels(models []RunModel, opts tfe.ListOptions, totalCount int) *tfe.RunList {
	var items []*tfe.Run
	for _, m := range models {
		items = append(items, NewRunFromModel(&m))
	}

	return &tfe.RunList{
		Items:      items,
		Pagination: ots.NewPagination(opts, totalCount),
	}
}

func NewRunFromModel(model *RunModel) *tfe.Run {
	return &tfe.Run{
		ID:                     model.ExternalID,
		Refresh:                model.Refresh,
		Message:                model.Message,
		ReplaceAddrs:           strings.Split(model.ReplaceAddrs, ","),
		TargetAddrs:            strings.Split(model.TargetAddrs, ","),
		ForceCancelAvailableAt: model.ForceCancelAvailableAt,
		Status:                 model.Status,
		Actions:                &model.RunActions,
		Permissions:            &model.RunPermissions,
		StatusTimestamps:       &model.RunStatusTimestamps,
		Plan:                   NewPlanFromModel(&model.Plan),
		Apply:                  NewApplyFromModel(&model.Apply),
		CreatedBy: &tfe.User{
			ID:       ots.DefaultUserID,
			Username: ots.DefaultUsername,
		},
		Workspace:            NewWorkspaceFromModel(&model.Workspace),
		ConfigurationVersion: NewConfigurationVersionFromModel(&model.ConfigurationVersion),
	}
}

func (RunModel) TableName() string {
	return "runs"
}

func (s RunService) CreateRun(opts *tfe.RunCreateOptions) (*tfe.Run, error) {
	ws, err := getWorkspaceByID(s.DB, opts.Workspace.ID)
	if err != nil {
		return nil, err
	}

	// If CV ID not provided then get workspace's latest CV
	var cv *ConfigurationVersionModel
	if opts.ConfigurationVersion != nil {
		cv, err = getConfigurationVersionByID(s.DB, opts.ConfigurationVersion.ID)
		if err != nil {
			return nil, err
		}
	} else {
		cv, err = getMostRecentConfigurationVersion(s.DB, ws.ID)
		if err != nil {
			return nil, err
		}
	}

	// TODO: wrap in TX
	plan, err := createPlan(s.DB)
	if err != nil {
		return nil, err
	}

	// TODO: wrap in TX
	apply, err := createApply(s.DB)
	if err != nil {
		return nil, err
	}

	model := RunModel{
		ExternalID:   ots.NewRunID(),
		Refresh:      ots.DefaultRefresh,
		ReplaceAddrs: strings.Join(opts.ReplaceAddrs, ","),
		TargetAddrs:  strings.Join(opts.TargetAddrs, ","),
		RunPermissions: tfe.RunPermissions{
			CanApply:       true,
			CanCancel:      true,
			CanDiscard:     true,
			CanForceCancel: true,
		},
		RunStatusTimestamps: tfe.RunStatusTimestamps{
			PlanQueueableAt: time.Now(),
		},
		// Simulate run having been immediately planned
		Status:                 tfe.RunPlanned,
		ConfigurationVersionID: cv.ID,
		ConfigurationVersion:   *cv,
		WorkspaceID:            ws.ID,
		Plan:                   *plan,
		Apply:                  *apply,
	}

	if opts.Message != nil {
		model.Message = *opts.Message
	}

	if opts.Refresh != nil {
		model.Refresh = *opts.Refresh
	}

	if result := s.DB.Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewRunFromModel(&model), nil
}

func (s RunService) ApplyRun(id string, opts *tfe.RunApplyOptions) error {
	model, err := getRunByID(s.DB, id)
	if err != nil {
		return err
	}

	model.Status = tfe.RunApplying
	if result := s.DB.Save(&model); result.Error != nil {
		return err
	}

	// Simulate run being applied...
	go func() {
		time.Sleep(time.Second)
		model.Status = tfe.RunApplied
		if result := s.DB.Save(&model); result.Error != nil {
			log.Printf("unable to update status on run: %s", result.Error.Error())
		}
	}()

	return nil
}

func (s RunService) ListRuns(workspaceID string, opts tfe.RunListOptions) (*tfe.RunList, error) {
	var models []RunModel
	var count int64

	ws, err := getWorkspaceByID(s.DB, workspaceID)
	if err != nil {
		return nil, err
	}

	query := s.DB.Preload(clause.Associations).Where("workspace_id = ?", ws.ID)

	if result := query.Model(models).Count(&count); result.Error != nil {
		return nil, result.Error
	}

	if result := query.Scopes(paginate(opts.ListOptions)).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	if err != nil {
		return nil, err
	}

	return NewRunListFromModels(models, opts.ListOptions, int(count)), nil
}

func (s RunService) GetRun(id string) (*tfe.Run, error) {
	model, err := getRunByID(s.DB, id)
	if err != nil {
		return nil, err
	}
	return NewRunFromModel(model), nil
}

func (s RunService) GetQueuedRuns(opts tfe.RunListOptions) (*tfe.RunList, error) {
	var models []RunModel
	var count int64

	if result := s.DB.Where("status = ?", tfe.RunPlanQueued).Limit(opts.PageSize).Offset((opts.PageNumber - 1) * opts.PageSize).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	var items []*tfe.Run
	for _, m := range models {
		items = append(items, NewRunFromModel(&m))
	}

	return &tfe.RunList{
		Items:      items,
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (s RunService) DiscardRun(id string, opts *tfe.RunDiscardOptions) error {
	model, err := getRunByID(s.DB, id)
	if err != nil {
		return err
	}

	switch tfe.RunStatus(model.Status) {
	case tfe.RunPending, tfe.RunPolicyChecked, tfe.RunPolicyOverride:
	default:
		return ots.ErrRunDiscardNotAllowed
	}

	model.Status = tfe.RunDiscarded
	if result := s.DB.Save(&model); result.Error != nil {
		return err
	}

	return nil
}

func (s RunService) CancelRun(id string, opts *tfe.RunCancelOptions) error {
	model, err := getRunByID(s.DB, id)
	if err != nil {
		return err
	}

	switch tfe.RunStatus(model.Status) {
	case tfe.RunPlanQueued, tfe.RunPending, tfe.RunPlanning, tfe.RunApplying:
	default:
		return ots.ErrRunCancelNotAllowed
	}

	model.ForceCancelAvailableAt = time.Now()
	model.Status = tfe.RunCanceled

	if result := s.DB.Save(&model); result.Error != nil {
		return err
	}

	return nil
}

func (s RunService) ForceCancelRun(id string, opts *tfe.RunForceCancelOptions) error {
	model, err := getRunByID(s.DB, id)
	if err != nil {
		return err
	}

	switch tfe.RunStatus(model.Status) {
	case tfe.RunPlanning, tfe.RunApplying:
	default:
		return ots.ErrRunForceCancelNotAllowed
	}

	model.Status = tfe.RunCanceled
	if result := s.DB.Save(&model); result.Error != nil {
		return err
	}

	return nil
}

func getRunByID(db *gorm.DB, id string) (*RunModel, error) {
	var model RunModel

	if result := db.Preload(clause.Associations).Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
