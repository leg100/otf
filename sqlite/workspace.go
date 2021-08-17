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
	return &WorkspaceDB{
		DB: db,
	}
}

// Create persists a Workspace to the DB. The returned Workspace is adorned with
// additional metadata, i.e. CreatedAt, UpdatedAt, etc.
func (db WorkspaceDB) Create(domain *ots.Workspace) (*ots.Workspace, error) {
	model := &Workspace{}
	model.FromDomain(domain)

	if result := db.Omit("Organization").Create(model); result.Error != nil {
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// Update persists an updated Workspace to the DB. The existing run is fetched
// from the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Workspace includes any changes,
// including a new UpdatedAt value.
func (db WorkspaceDB) Update(spec ots.WorkspaceSpecifier, fn func(*ots.Workspace) error) (*ots.Workspace, error) {
	var model *Workspace

	err := db.Transaction(func(tx *gorm.DB) (err error) {
		// Get existing model obj from DB
		model, err = getWorkspace(tx, spec)
		if err != nil {
			return err
		}

		// Update domain obj using client-supplied fn
		if err := model.Update(fn); err != nil {
			return err
		}

		if result := tx.Save(model); result.Error != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert back to domain obj
	return model.ToDomain(), nil
}

func (db WorkspaceDB) List(opts ots.WorkspaceListOptions) (*ots.WorkspaceList, error) {
	var models WorkspaceList
	var count int64

	err := db.Transaction(func(tx *gorm.DB) error {
		query := tx

		if opts.OrganizationName != nil {
			org, err := getOrganizationByName(tx, *opts.OrganizationName)
			if err != nil {
				return err
			}

			query = query.Where("organization_id = ?", org.Model.ID)
		}

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
		Items:      models.ToDomain(),
		Pagination: ots.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (db WorkspaceDB) Get(spec ots.WorkspaceSpecifier) (*ots.Workspace, error) {
	ws, err := getWorkspace(db.DB, spec)
	if err != nil {
		return nil, err
	}
	return ws.ToDomain(), nil
}

// Delete deletes a specific workspace, along with its associated records (runs
// etc).
func (db WorkspaceDB) Delete(spec ots.WorkspaceSpecifier) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		var ws Workspace
		var query *gorm.DB

		switch {
		case spec.ID != nil:
			// Get workspace by ID
			query = tx.Where("external_id = ?", *spec.ID)
		case spec.Name != nil && spec.OrganizationName != nil:
			// Get workspace by name and organization name
			org, err := getOrganizationByName(tx, *spec.OrganizationName)
			if err != nil {
				return err
			}

			query = tx.Where("organization_id = ? AND name = ?", org.ID, spec.Name)
		default:
			return ots.ErrInvalidWorkspaceSpecifier
		}

		// Retrieve workspace
		if result := query.First(&ws); result.Error != nil {
			return result.Error
		}

		// Delete workspace
		if result := query.Delete(&ws); result.Error != nil {
			return result.Error
		}

		// Delete associated runs
		if result := tx.Delete(&Run{}, "workspace_id = ?", ws.ID); result.Error != nil {
			return result.Error
		}

		// Delete associated state versions
		if result := tx.Delete(&StateVersion{}, "workspace_id = ?", ws.ID); result.Error != nil {
			return result.Error
		}

		// Delete associated configuration versions
		if result := tx.Delete(&ConfigurationVersion{}, "workspace_id = ?", ws.ID); result.Error != nil {
			return result.Error
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func getWorkspace(db *gorm.DB, spec ots.WorkspaceSpecifier) (*Workspace, error) {
	var model Workspace

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
		return nil, ots.ErrInvalidWorkspaceSpecifier
	}

	if result := query.First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
