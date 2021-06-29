package sqlite

import (
	"fmt"
	"strings"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ots.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceModel struct {
	gorm.Model

	Name                string
	ExternalID          string
	AllowDestroyPlan    bool
	AutoApply           bool
	Description         string
	ExecutionMode       string
	FileTriggersEnabled bool
	Operations          bool
	QueueAllRuns        bool
	SpeculativeEnabled  bool
	GlobalRemoteState   bool
	Locked              bool
	SourceName          string
	SourceURL           string
	TerraformVersion    string
	TriggerPrefixes     string
	WorkingDirectory    string

	OrganizationID uint
	Organization   OrganizationModel
}

type WorkspaceService struct {
	*gorm.DB
}

func NewWorkspaceService(db *gorm.DB) *WorkspaceService {
	db.AutoMigrate(&WorkspaceModel{})

	return &WorkspaceService{
		DB: db,
	}
}

func NewWorkspaceFromModel(model *WorkspaceModel) *ots.Workspace {
	return &ots.Workspace{
		Name:                model.Name,
		ID:                  model.ExternalID,
		Description:         model.Description,
		FileTriggersEnabled: model.FileTriggersEnabled,
		AutoApply:           model.AutoApply,
		Operations:          model.Operations,
		QueueAllRuns:        model.QueueAllRuns,
		SpeculativeEnabled:  model.SpeculativeEnabled,
		TerraformVersion:    model.TerraformVersion,
		WorkingDirectory:    model.WorkingDirectory,
		Locked:              model.Locked,
		TriggerPrefixes:     strings.Split(model.TriggerPrefixes, ","),
		Permissions:         &ots.WorkspacePermissions{},
		Organization:        NewOrganizationFromModel(&model.Organization),
	}
}

func (WorkspaceModel) TableName() string {
	return "workspaces"
}

func (s WorkspaceService) CreateWorkspace(orgName string, opts *tfe.WorkspaceCreateOptions) (*ots.Workspace, error) {
	org, err := getOrganizationByName(s.DB, orgName)
	if err != nil {
		return nil, err
	}

	model := WorkspaceModel{
		Name:                *opts.Name,
		ExternalID:          ots.NewWorkspaceID(),
		OrganizationID:      org.ID,
		ExecutionMode:       "local", // Only local execution mode is supported
		FileTriggersEnabled: ots.DefaultFileTriggersEnabled,
		GlobalRemoteState:   true, // Only global remote state is supported
		TerraformVersion:    ots.DefaultTerraformVersion,
		Organization:        *org,
	}

	if opts.AllowDestroyPlan != nil {
		model.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		model.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		model.Description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		model.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.Operations != nil {
		model.Operations = *opts.Operations
	}
	if opts.QueueAllRuns != nil {
		model.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SpeculativeEnabled != nil {
		model.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.TerraformVersion != nil {
		model.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		model.TriggerPrefixes = strings.Join(opts.TriggerPrefixes, ",")
	}
	if opts.WorkingDirectory != nil {
		model.WorkingDirectory = *opts.WorkingDirectory
	}

	if result := s.DB.Omit(clause.Associations).Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(&model), nil
}

func (s WorkspaceService) UpdateWorkspace(name, orgName string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	var model WorkspaceModel

	org, err := getOrganizationByName(s.DB, orgName)
	if err != nil {
		return nil, err
	}

	update := make(map[string]interface{})
	if opts.Name != nil {
		update["name"] = *opts.Name
	}
	if opts.AllowDestroyPlan != nil {
		update["allow_destroy_plan"] = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		update["auto_apply"] = *opts.AutoApply
	}
	if opts.Description != nil {
		update["description"] = *opts.Description
	}
	if opts.TerraformVersion != nil {
		update["terraform_version"] = *opts.TerraformVersion
	}

	if result := s.DB.Where("name = ? AND organization_id = ?", name, org.ID).First(&model); result.Error != nil {
		return nil, result.Error
	}

	if result := s.DB.Model(&model).Updates(update); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(&model), nil
}

func (s WorkspaceService) UpdateWorkspaceByID(id string, opts *tfe.WorkspaceUpdateOptions) (*ots.Workspace, error) {
	var model WorkspaceModel

	update := make(map[string]interface{})
	if opts.Name != nil {
		update["name"] = *opts.Name
	}

	if result := s.DB.Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	if result := s.DB.Model(&model).Updates(update); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(&model), nil
}

func (s WorkspaceService) ListWorkspaces(orgName string, opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	var models []WorkspaceModel
	var count int64

	_, err := getOrganizationByName(s.DB, orgName)
	if err != nil {
		return nil, err
	}

	query := s.DB.Preload(clause.Associations)

	if opts.Search != nil {
		query = query.Where("name LIKE ?", fmt.Sprintf("%s%%", *opts.Search))
	}

	if result := query.Model(&models).Count(&count); result.Error != nil {
		return nil, result.Error
	}

	if result := query.Limit(opts.PageSize).Offset((opts.PageNumber - 1) * opts.PageSize).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	var items []*ots.Workspace
	for _, m := range models {
		items = append(items, NewWorkspaceFromModel(&m))
	}

	return &ots.WorkspaceList{
		Items:      items,
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (s WorkspaceService) GetWorkspace(name, orgName string) (*ots.Workspace, error) {
	model, err := getWorkspaceByName(s.DB, name, orgName)
	if err != nil {
		return nil, err
	}

	return NewWorkspaceFromModel(model), nil
}

func (s WorkspaceService) GetWorkspaceByID(id string) (*ots.Workspace, error) {
	var model WorkspaceModel

	if result := s.DB.Preload(clause.Associations).Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(&model), nil
}

func (s WorkspaceService) DeleteWorkspace(name, orgName string) error {
	var model WorkspaceModel

	_, err := getOrganizationByName(s.DB, orgName)
	if err != nil {
		return err
	}

	if result := s.DB.Where("name = ?", name).First(&model); result.Error != nil {
		return result.Error
	}

	if result := s.DB.Delete(&model); result.Error != nil {
		return result.Error
	}

	return nil
}

func (s WorkspaceService) DeleteWorkspaceByID(id string) error {
	var model WorkspaceModel

	if result := s.DB.Where("external_id = ?", id).First(&model); result.Error != nil {
		return result.Error
	}

	if result := s.DB.Delete(&model); result.Error != nil {
		return result.Error
	}

	return nil
}

func (s WorkspaceService) LockWorkspace(id string, opts ots.WorkspaceLockOptions) (*ots.Workspace, error) {
	return toggleWorkspaceLock(s.DB, id, true)
}

func (s WorkspaceService) UnlockWorkspace(id string) (*ots.Workspace, error) {
	return toggleWorkspaceLock(s.DB, id, false)
}

func toggleWorkspaceLock(db *gorm.DB, id string, lock bool) (*ots.Workspace, error) {
	var model WorkspaceModel

	if result := db.Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	if lock && model.Locked {
		return nil, ots.ErrWorkspaceAlreadyLocked
	}
	if !lock && !model.Locked {
		return nil, ots.ErrWorkspaceAlreadyUnlocked
	}

	if result := db.Model(&model).Update("locked", lock); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(&model), nil
}

func getWorkspaceByID(db *gorm.DB, id string) (*WorkspaceModel, error) {
	var model WorkspaceModel

	if result := db.Preload(clause.Associations).Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func getWorkspaceByName(db *gorm.DB, name, orgName string) (*WorkspaceModel, error) {
	var model WorkspaceModel

	org, err := getOrganizationByName(db, orgName)
	if err != nil {
		return nil, err
	}

	if result := db.Preload(clause.Associations).Where("name = ? AND organization_id = ?", name, org.ID).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
