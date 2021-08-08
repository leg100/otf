package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ots.StateVersionStore = (*StateVersionService)(nil)

type StateVersionService struct {
	*gorm.DB
}

func NewStateVersionDB(db *gorm.DB) *StateVersionService {
	return &StateVersionService{
		DB: db,
	}
}

// Create persists a StateVersion to the DB.
func (s StateVersionService) Create(sv *ots.StateVersion) (*ots.StateVersion, error) {
	model := &StateVersion{}
	model.FromDomain(sv)

	if result := s.DB.Omit("Workspace").Create(model); result.Error != nil {
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

func (s StateVersionService) List(opts tfe.StateVersionListOptions) (*ots.StateVersionList, error) {
	var models StateVersionList
	var count int64

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		ws, err := getWorkspace(tx, ots.WorkspaceSpecifier{Name: opts.Workspace, OrganizationName: opts.Organization})
		if err != nil {
			return err
		}

		query := tx.Where("workspace_id = ?", ws.ID)

		if result := query.Model(&models).Count(&count); result.Error != nil {
			return result.Error
		}

		if result := query.Preload(clause.Associations).Scopes(paginate(opts.ListOptions)).Find(&models); result.Error != nil {
			return result.Error
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ots.StateVersionList{
		Items:      models.ToDomain(),
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (s StateVersionService) Get(opts ots.StateVersionGetOptions) (*ots.StateVersion, error) {
	sv, err := getStateVersion(s.DB, opts)
	if err != nil {
		return nil, err
	}
	return sv.ToDomain(), nil
}

func getStateVersion(db *gorm.DB, opts ots.StateVersionGetOptions) (*StateVersion, error) {
	var model StateVersion

	query := db.Preload(clause.Associations)

	switch {
	case opts.ID != nil:
		// Get state version by ID
		query = query.Where("external_id = ?", *opts.ID)
	case opts.WorkspaceID != nil:
		// Get most recent state version belonging to workspace
		query = query.Joins("JOIN workspaces ON workspaces.id = state_versions.workspace_id").
			Order("state_versions.serial desc, state_versions.created_at desc").
			Where("workspaces.external_id = ?", *opts.WorkspaceID)
	default:
		return nil, ots.ErrInvalidStateVersionGetOptions
	}

	if result := query.First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
