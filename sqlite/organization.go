package sqlite

import (
	"github.com/leg100/otf"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ otf.OrganizationStore = (*OrganizationDB)(nil)

type OrganizationDB struct {
	*gorm.DB
}

func NewOrganizationDB(db *gorm.DB) *OrganizationDB {
	return &OrganizationDB{
		DB: db,
	}
}

// Create persists a Organization to the DB.
func (db OrganizationDB) Create(domain *otf.Organization) (*otf.Organization, error) {
	model := &Organization{}
	model.FromDomain(domain)

	if result := db.DB.Create(model); result.Error != nil {
		return nil, result.Error
	}

	return model.ToDomain(), nil
}

// Update persists an updated Organization to the DB. The existing org is
// fetched from the DB, the supplied func is invoked on the org, and the updated
// org is persisted back to the DB.
func (db OrganizationDB) Update(name string, fn func(*otf.Organization) error) (*otf.Organization, error) {
	var model *Organization

	err := db.Transaction(func(tx *gorm.DB) (err error) {
		// Get existing model obj from DB
		model, err = getOrganizationByName(tx, name)
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

	return model.ToDomain(), nil
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	var count int64
	var models OrganizationList

	err := db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Model(models).Count(&count); result.Error != nil {
			return result.Error
		}

		if result := tx.Scopes(paginate(opts.ListOptions)).Find(&models); result.Error != nil {
			return result.Error
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &otf.OrganizationList{
		Items:      models.ToDomain(),
		Pagination: otf.NewPagination(opts.ListOptions, int(count)),
	}, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	org, err := getOrganizationByName(db.DB, name)
	if err != nil {
		return nil, err
	}
	return org.ToDomain(), nil
}

func (db OrganizationDB) Delete(name string) error {
	if result := db.Where("name = ?", name).Delete(&Organization{}); result.Error != nil {
		return result.Error
	}

	return nil
}

func getOrganizationByName(db *gorm.DB, name string) (*Organization, error) {
	var model Organization

	if result := db.Preload(clause.Associations).Where("name = ?", name).First(&model); result.Error != nil {
		return nil, result.Error
	}

	return &model, nil
}
