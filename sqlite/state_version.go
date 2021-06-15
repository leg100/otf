package sqlite

import (
	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ots.StateVersionService = (*StateVersionService)(nil)

type StateVersionModel struct {
	gorm.Model

	Serial     int64
	ExternalID string
	State      string

	// State version belongs to a workspace
	WorkspaceID uint
	Workspace   WorkspaceModel

	// State version has many outputs
	StateVersionOutputs []StateVersionOutputModel
}

type StateVersionService struct {
	*gorm.DB
}

func NewStateVersionService(db *gorm.DB) *StateVersionService {
	db.AutoMigrate(&StateVersionModel{})

	return &StateVersionService{
		DB: db,
	}
}

func NewStateVersionFromModel(model *StateVersionModel) *ots.StateVersion {
	return &ots.StateVersion{
		ID:     model.ExternalID,
		Serial: model.Serial,
	}
}

func (StateVersionModel) TableName() string {
	return "state_versions"
}

func (s StateVersionService) CreateStateVersion(workspaceID string, opts *ots.StateVersionCreateOptions) (*ots.StateVersion, error) {
	workspace, err := getWorkspaceByID(s.DB, workspaceID)
	if err != nil {
		return nil, err
	}

	model := StateVersionModel{
		Serial:      *opts.Serial,
		ExternalID:  ots.NewStateVersionID(),
		WorkspaceID: workspace.ID,
	}

	if result := s.DB.Omit(clause.Associations).Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewStateVersionFromModel(&model), nil
}

func (s StateVersionService) ListStateVersions(orgName, workspaceName string, opts ots.StateVersionListOptions) (*ots.StateVersionList, error) {
	var models []StateVersionModel

	workspace, err := getWorkspaceByName(s.DB, workspaceName, orgName)
	if err != nil {
		return nil, err
	}

	query := s.DB
	query = query.Where("workspace_id = ?", workspace.ID)
	query = query.Limit(opts.PageSize).Offset((opts.PageNumber - 1) * opts.PageSize)

	if result := query.Find(&models); result.Error != nil {
		return nil, result.Error
	}

	workspaces := &ots.StateVersionList{
		StateVersionListOptions: ots.StateVersionListOptions{
			ListOptions: opts.ListOptions,
		},
	}
	for _, m := range models {
		workspaces.Items = append(workspaces.Items, NewStateVersionFromModel(&m))
	}

	return workspaces, nil
}

func (s StateVersionService) GetStateVersion(id string) (*ots.StateVersion, error) {
	var model StateVersionModel

	if result := s.DB.Preload(clause.Associations).Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewStateVersionFromModel(&model), nil
}

func (s StateVersionService) CurrentStateVersion(workspaceID string) (*ots.StateVersion, error) {
	var model StateVersionModel

	if result := s.DB.Preload(clause.Associations).Where("workspace_id = ?", workspaceID).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewStateVersionFromModel(&model), nil
}
