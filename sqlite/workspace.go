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

	Name                       string
	ExternalID                 string
	AllowDestroyPlan           bool
	AutoApply                  bool
	Description                string
	ExecutionMode              string
	FileTriggersEnabled        bool
	Operations                 bool
	QueueAllRuns               bool
	SpeculativeEnabled         bool
	GlobalRemoteState          bool
	Locked                     bool
	SourceName                 string
	SourceURL                  string
	StructuredRunOutputEnabled bool
	TerraformVersion           string
	TriggerPrefixes            string
	WorkingDirectory           string

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

func NewWorkspaceListFromModels(models []WorkspaceModel, opts tfe.ListOptions, totalCount int) *tfe.WorkspaceList {
	var items []*tfe.Workspace
	for _, m := range models {
		items = append(items, NewWorkspaceFromModel(&m))
	}

	return &tfe.WorkspaceList{
		Items:      items,
		Pagination: ots.NewPagination(opts, totalCount),
	}
}

func NewWorkspaceFromModel(model *WorkspaceModel) *tfe.Workspace {
	return &tfe.Workspace{
		Name: model.Name,
		ID:   model.ExternalID,
		Actions: &tfe.WorkspaceActions{
			IsDestroyable: true,
		},
		CreatedAt:                  model.CreatedAt,
		Description:                model.Description,
		FileTriggersEnabled:        model.FileTriggersEnabled,
		AutoApply:                  model.AutoApply,
		AllowDestroyPlan:           model.AllowDestroyPlan,
		Operations:                 model.Operations,
		QueueAllRuns:               model.QueueAllRuns,
		SourceName:                 model.SourceName,
		SourceURL:                  model.SourceURL,
		SpeculativeEnabled:         model.SpeculativeEnabled,
		StructuredRunOutputEnabled: model.StructuredRunOutputEnabled,
		TerraformVersion:           model.TerraformVersion,
		WorkingDirectory:           model.WorkingDirectory,
		Locked:                     model.Locked,
		TriggerPrefixes:            strings.Split(model.TriggerPrefixes, ","),
		Permissions: &tfe.WorkspacePermissions{
			CanDestroy: true,
		},
		Organization: NewOrganizationFromModel(&model.Organization),
	}
}

func (WorkspaceModel) TableName() string {
	return "workspaces"
}

func (s WorkspaceService) CreateWorkspace(orgName string, opts *tfe.WorkspaceCreateOptions) (*tfe.Workspace, error) {
	org, err := getOrganizationByName(s.DB, orgName)
	if err != nil {
		return nil, err
	}

	ws := &WorkspaceModel{
		Name:                *opts.Name,
		ExternalID:          ots.NewWorkspaceID(),
		AllowDestroyPlan:    ots.DefaultAllowDestroyPlan,
		ExecutionMode:       "local", // Only local execution mode is supported
		FileTriggersEnabled: ots.DefaultFileTriggersEnabled,
		GlobalRemoteState:   true, // Only global remote state is supported
		TerraformVersion:    ots.DefaultTerraformVersion,
		Organization:        *org,
		OrganizationID:      org.ID,
	}

	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.Operations != nil {
		ws.Operations = *opts.Operations
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SourceName != nil {
		ws.SourceName = *opts.SourceName
	}
	if opts.SourceURL != nil {
		ws.SourceURL = *opts.SourceURL
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		ws.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = strings.Join(opts.TriggerPrefixes, ",")
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}

	if result := s.DB.Omit(clause.Associations).Create(ws); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(ws), nil
}

func (s WorkspaceService) UpdateWorkspace(name, orgName string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	ws, err := getWorkspaceByName(s.DB, name, orgName)
	if err != nil {
		return nil, err
	}

	return updateWorkspace(s.DB, ws, opts)
}

func (s WorkspaceService) UpdateWorkspaceByID(id string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	ws, err := getWorkspaceByID(s.DB, id)
	if err != nil {
		return nil, err
	}

	return updateWorkspace(s.DB, ws, opts)
}

func (s WorkspaceService) ListWorkspaces(orgName string, opts tfe.WorkspaceListOptions) (*tfe.WorkspaceList, error) {
	var models []WorkspaceModel
	var count int64

	org, err := getOrganizationByName(s.DB, orgName)
	if err != nil {
		return nil, err
	}

	query := s.DB.Where("organization_id = ?", org.ID)

	if opts.Search != nil {
		query = query.Where("name LIKE ?", fmt.Sprintf("%s%%", *opts.Search))
	}

	if result := query.Model(&models).Count(&count); result.Error != nil {
		return nil, result.Error
	}

	if result := query.Preload(clause.Associations).Scopes(paginate(opts.ListOptions)).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceListFromModels(models, opts.ListOptions, int(count)), nil
}

func (s WorkspaceService) GetWorkspace(name, orgName string) (*tfe.Workspace, error) {
	model, err := getWorkspaceByName(s.DB, name, orgName)
	if err != nil {
		return nil, err
	}
	return NewWorkspaceFromModel(model), err
}

func (s WorkspaceService) GetWorkspaceByID(id string) (*tfe.Workspace, error) {
	model, err := getWorkspaceByID(s.DB, id)
	if err != nil {
		return nil, err
	}
	return NewWorkspaceFromModel(model), err
}

func (s WorkspaceService) DeleteWorkspace(name, orgName string) error {
	model, err := getWorkspaceByName(s.DB, name, orgName)
	if err != nil {
		return err
	}

	if result := s.DB.Delete(model); result.Error != nil {
		return result.Error
	}

	return nil
}

func (s WorkspaceService) DeleteWorkspaceByID(id string) error {
	model, err := getWorkspaceByID(s.DB, id)
	if err != nil {
		return err
	}

	if result := s.DB.Delete(&model); result.Error != nil {
		return result.Error
	}

	return nil
}

func (s WorkspaceService) LockWorkspace(id string, opts tfe.WorkspaceLockOptions) (*tfe.Workspace, error) {
	return toggleWorkspaceLock(s.DB, id, true)
}

func (s WorkspaceService) UnlockWorkspace(id string) (*tfe.Workspace, error) {
	return toggleWorkspaceLock(s.DB, id, false)
}

func toggleWorkspaceLock(db *gorm.DB, id string, lock bool) (*tfe.Workspace, error) {
	model, err := getWorkspaceByID(db, id)
	if err != nil {
		return nil, err
	}

	if lock && model.Locked {
		return nil, ots.ErrWorkspaceAlreadyLocked
	}
	if !lock && !model.Locked {
		return nil, ots.ErrWorkspaceAlreadyUnlocked
	}

	model.Locked = lock

	if result := db.Save(model); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(model), nil
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

	if result := db.Preload(clause.Associations).Joins("JOIN organizations ON organizations.id = workspaces.organization_id").First(&model, "workspaces.name = ? AND organizations.name = ?", name, orgName); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func updateWorkspace(db *gorm.DB, ws *WorkspaceModel, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
	if opts.Name != nil {
		ws.Name = *opts.Name
	}

	if opts.AllowDestroyPlan != nil {
		ws.AllowDestroyPlan = *opts.AllowDestroyPlan
	}
	if opts.AutoApply != nil {
		ws.AutoApply = *opts.AutoApply
	}
	if opts.Description != nil {
		ws.Description = *opts.Description
	}
	if opts.FileTriggersEnabled != nil {
		ws.FileTriggersEnabled = *opts.FileTriggersEnabled
	}
	if opts.Operations != nil {
		ws.Operations = *opts.Operations
	}
	if opts.QueueAllRuns != nil {
		ws.QueueAllRuns = *opts.QueueAllRuns
	}
	if opts.SpeculativeEnabled != nil {
		ws.SpeculativeEnabled = *opts.SpeculativeEnabled
	}
	if opts.StructuredRunOutputEnabled != nil {
		ws.StructuredRunOutputEnabled = *opts.StructuredRunOutputEnabled
	}
	if opts.TerraformVersion != nil {
		ws.TerraformVersion = *opts.TerraformVersion
	}
	if opts.TriggerPrefixes != nil {
		ws.TriggerPrefixes = strings.Join(opts.TriggerPrefixes, ",")
	}
	if opts.WorkingDirectory != nil {
		ws.WorkingDirectory = *opts.WorkingDirectory
	}

	if result := db.Save(ws); result.Error != nil {
		return nil, result.Error
	}

	return NewWorkspaceFromModel(ws), nil
}
