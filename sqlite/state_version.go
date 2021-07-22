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

// CreateStateVersion persists a StateVersion to the DB.
func (s StateVersionService) Create(sv *ots.StateVersion) (*ots.StateVersion, error) {
	if result := s.DB.Omit("Workspaces").Create(sv); result.Error != nil {
		return nil, result.Error
	}

	return sv, nil
}

// UpdateStateVersion persists an updated StateVersion to the DB. The existing run is fetched from
// the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned StateVersion includes any changes, including a
// new UpdatedAt value.
func (s StateVersionService) Update(id string, fn func(*ots.StateVersion) error) (*ots.StateVersion, error) {
	var sv *ots.StateVersion

	err := s.DB.Transaction(func(tx *gorm.DB) (err error) {
		// Get existing model obj from DB
		sv, err = getStateVersion(tx, ots.StateVersionGetOptions{ID: &id})
		if err != nil {
			return err
		}

		// Update domain obj using client-supplied fn
		if err := fn(sv); err != nil {
			return err
		}

		if result := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(sv); result.Error != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return sv, nil
}

func (s StateVersionService) List(opts tfe.StateVersionListOptions) (*ots.StateVersionList, error) {
	var models []ots.StateVersion
	var count int64

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		ws, err := getWorkspace(tx, ots.WorkspaceSpecifier{Name: opts.Workspace, OrganizationName: opts.Organization})
		if err != nil {
			return err
		}

		query := tx.Where("workspace_id = ?", ws.InternalID)

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
		Items:      stateVersionListToPointerList(models),
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (s StateVersionService) Get(opts ots.StateVersionGetOptions) (*ots.StateVersion, error) {
	sv, err := getStateVersion(s.DB, opts)
	if err != nil {
		return nil, err
	}
	return sv, nil
}

func getStateVersion(db *gorm.DB, opts ots.StateVersionGetOptions) (*ots.StateVersion, error) {
	var sv ots.StateVersion

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

	if result := query.First(&sv); result.Error != nil {
		return nil, result.Error
	}

	return &sv, nil
}

func stateVersionListToPointerList(svl []ots.StateVersion) (pl []*ots.StateVersion) {
	for i := range svl {
		pl = append(pl, &svl[i])
	}
	return
}
