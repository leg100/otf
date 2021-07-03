package sqlite

import (
	"encoding/base64"
	"fmt"

	"github.com/leg100/go-tfe"
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
	StateVersionOutputs []StateVersionOutputModel `gorm:"foreignKey:StateVersionID"`
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

func NewStateVersionFromModel(model *StateVersionModel) *tfe.StateVersion {
	return &tfe.StateVersion{
		ID:          model.ExternalID,
		Serial:      model.Serial,
		DownloadURL: fmt.Sprintf("/state-versions/%s/download", model.ExternalID),
		Outputs:     NewStateVersionOutputsFromModels(model.StateVersionOutputs),
		CreatedAt:   model.CreatedAt,
	}
}

func (StateVersionModel) TableName() string {
	return "state_versions"
}

func (s StateVersionService) CreateStateVersion(workspaceID string, opts *tfe.StateVersionCreateOptions) (*tfe.StateVersion, error) {
	workspace, err := getWorkspaceByID(s.DB, workspaceID)
	if err != nil {
		return nil, err
	}

	model := StateVersionModel{
		Serial:      *opts.Serial,
		ExternalID:  ots.NewStateVersionID(),
		State:       *opts.State,
		Workspace:   *workspace,
		WorkspaceID: workspace.ID,
	}

	decoded, err := base64.StdEncoding.DecodeString(*opts.State)
	if err != nil {
		return nil, err
	}

	state, err := ots.Parse(decoded)
	if err != nil {
		return nil, err
	}

	for k, v := range state.Outputs {
		model.StateVersionOutputs = append(model.StateVersionOutputs, StateVersionOutputModel{
			ExternalID: ots.NewStateVersionOutputID(),
			Name:       k,
			Type:       v.Type,
			Value:      v.Value,
		})
	}

	if result := s.DB.Omit("Workspaces").Create(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewStateVersionFromModel(&model), nil
}

func (s StateVersionService) ListStateVersions(orgName, workspaceName string, opts tfe.StateVersionListOptions) (*tfe.StateVersionList, error) {
	var models []StateVersionModel
	var count int64

	workspace, err := getWorkspaceByName(s.DB, workspaceName, orgName)
	if err != nil {
		return nil, err
	}

	query := s.DB.
		Preload(clause.Associations).
		Where("workspace_id = ?", workspace.ID)

	if result := query.Model(&models).Count(&count); result.Error != nil {
		return nil, result.Error
	}

	if result := query.Limit(opts.PageSize).Offset((opts.PageNumber - 1) * opts.PageSize).Find(&models); result.Error != nil {
		return nil, result.Error
	}

	var items []*tfe.StateVersion
	for _, m := range models {
		items = append(items, NewStateVersionFromModel(&m))
	}

	return &tfe.StateVersionList{
		Items:      items,
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (s StateVersionService) GetStateVersion(id string) (*tfe.StateVersion, error) {
	sv, err := getStateVersionByID(s.DB, id)
	if err != nil {
		return nil, err
	}
	return NewStateVersionFromModel(sv), err
}

func (s StateVersionService) CurrentStateVersion(workspaceID string) (*tfe.StateVersion, error) {
	var model StateVersionModel

	workspace, err := getWorkspaceByID(s.DB, workspaceID)
	if err != nil {
		return nil, err
	}

	if result := s.DB.Preload(clause.Associations).Where("workspace_id = ?", workspace.ID).Order("serial desc, created_at desc").First(&model); result.Error != nil {
		return nil, result.Error
	}

	return NewStateVersionFromModel(&model), nil
}

func (s StateVersionService) DownloadStateVersion(id string) ([]byte, error) {
	sv, err := getStateVersionByID(s.DB, id)
	if err != nil {
		return nil, err
	}

	return base64.StdEncoding.DecodeString(sv.State)
}

func getStateVersionByID(db *gorm.DB, id string) (*StateVersionModel, error) {
	var model StateVersionModel

	if result := db.Preload(clause.Associations).Where("external_id = ?", id).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
