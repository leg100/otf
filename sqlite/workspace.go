package sqlite

import (
	"fmt"

	"github.com/leg100/ots"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ ots.WorkspaceRepository = (*WorkspaceDB)(nil)

type WorkspaceDB struct {
	*gorm.DB
}

func NewWorkspaceDB(db *gorm.DB) *WorkspaceDB {
	db.AutoMigrate(&ots.Workspace{})

	return &WorkspaceDB{
		DB: db,
	}
}

// CreateWorkspace persists a Workspace to the DB. The returned Workspace is adorned with
// additional metadata, i.e. CreatedAt, UpdatedAt, etc.
func (db WorkspaceDB) Create(ws *ots.Workspace) (*ots.Workspace, error) {
	if result := db.Omit("Organization").Create(ws); result.Error != nil {
		return nil, result.Error
	}

	return ws, nil
}

// UpdateWorkspace persists an updated Workspace to the DB. The existing run is fetched from
// the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Workspace includes any changes, including a
// new UpdatedAt value.
func (db WorkspaceDB) Update(spec ots.WorkspaceSpecifier, fn func(*ots.Workspace) error) (*ots.Workspace, error) {
	var ws *ots.Workspace

	err := db.Transaction(func(tx *gorm.DB) (err error) {
		// Get existing model obj from DB
		ws, err = getWorkspace(tx, spec)
		if err != nil {
			return err
		}

		// Update domain obj using client-supplied fn
		if err := fn(ws); err != nil {
			return err
		}

		if result := tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(ws); result.Error != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert back to domain obj
	return ws, nil
}

func (db WorkspaceDB) List(organizationName string, opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	var models []ots.Workspace
	var count int64

	err := db.Transaction(func(tx *gorm.DB) error {
		org, err := getOrganization(tx, organizationName)
		if err != nil {
			return err
		}

		query := tx.Where("organization_id = ?", org.InternalID)

		if opts.Prefix != nil {
			query = query.Where("name LIKE ?", fmt.Sprintf("%s%%", *opts.Prefix))
		}

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

	return &ots.WorkspaceList{
		Items:      workspaceListToPointerList(models),
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (db WorkspaceDB) Get(spec ots.WorkspaceSpecifier) (*ots.Workspace, error) {
	ws, err := getWorkspace(db.DB, spec)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (db WorkspaceDB) Delete(spec ots.WorkspaceSpecifier) error {
	query := db.DB

	switch {
	case spec.ID != nil:
		// Delete workspace by ID
		query = query.Where("external_id = ?", *spec.ID)
	case spec.Name != nil && spec.OrganizationName != nil:
		// Delete workspace by name and organization name
		org, err := getOrganization(db.DB, *spec.OrganizationName)
		if err != nil {
			return err
		}

		query = query.Where("organization_id = ? AND workspaces.name = ?", org.InternalID, spec.Name)
	default:
		return ots.ErrInvalidWorkspaceDeleteOptions
	}

	if result := query.Delete(&ots.Workspace{}); result.Error != nil {
		return result.Error
	}

	return nil
}

func getWorkspace(db *gorm.DB, spec ots.WorkspaceSpecifier) (*ots.Workspace, error) {
	var model ots.Workspace

	query := db.Preload(clause.Associations)

	switch {
	case spec.ID != nil:
		// Get workspace by ID
		query = query.Where("external_id = ?", *spec.ID)
	case spec.Name != nil && spec.OrganizationName != nil:
		// Get workspace by name and organization name
		query = query.Joins("JOIN organizations ON organizations.id = workspaces.organization_id").
			Where("workspaces.name = ? AND organizations.name = ?", spec.Name, spec.OrganizationName)
	default:
		return nil, ots.ErrInvalidWorkspaceGetOptions
	}

	if result := query.First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}

func workspaceListToPointerList(workspaces []ots.Workspace) (pl []*ots.Workspace) {
	for i := range workspaces {
		pl = append(pl, &workspaces[i])
	}
	return
}
