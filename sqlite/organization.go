package sqlite

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"gorm.io/gorm"
)

var _ otf.OrganizationStore = (*OrganizationDB)(nil)

// Organization models a row in a organizations table.
type OrganizationDB struct {
	*sqlx.DB
}

func NewOrganizationDB(db *sqlx.DB) *OrganizationDB {
	return &OrganizationDB{
		DB: db,
	}
}

// Create persists a Organization to the DB.
func (db OrganizationDB) Create(org *otf.Organization) (*otf.Organization, error) {
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()

	clauses := map[string]interface{}{
		"created_at":  org.CreatedAt,
		"updated_at":  org.UpdatedAt,
		"external_id": org.ID,
		"name":        org.Name,
		"email":       org.Email,
	}

	_, err := sq.Insert("organizations").SetMap(clauses).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}

	return org, nil
}

// Update persists an updated Organization to the DB. The existing org is
// fetched from the DB, the supplied func is invoked on the org, and the updated
// org is persisted back to the DB.
func (db OrganizationDB) Update(name string, fn func(*otf.Organization) error) (*otf.Organization, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	org, err := getOrganizationByName(tx, name)
	if err != nil {
		return nil, err
	}

	before := otf.Organization{}
	copier.Copy(&before, org)

	// Update obj using client-supplied fn
	if err := fn(org); err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	setIfChanged(before.Name, org.Name, updates, "name")
	setIfChanged(before.Email, org.Email, updates, "email")

	if len(updates) == 0 {
		return org, nil
	}

	updates["updated_at"] = time.Now()

	_, err = sq.Update("organizations").SetMap(updates).Where("name = ?", name).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}

	return org, tx.Commit()
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	query := sq.Select("*").From("organizations")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	ol := otf.OrganizationList{}
	var count int

	rows, err := db.Queryx(sql, args)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		org, err := scanOrganization(rows)
		if err != nil {
			return nil, err
		}

		ol.Items = append(ol.Items, org)
		count++
	}

	ol.Pagination = otf.NewPagination(opts.ListOptions, count)

	return &ol, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	return getOrganizationByName(db.DB, name)
}

func (db OrganizationDB) Delete(name string) error {
	_, err := sq.Delete("organizations").Where("name = ?", name).RunWith(db).Exec()
	return err
}

func getOrganizationByName(db sqlx.Queryer, name string) (*otf.Organization, error) {
	return getOrganization(db, "name = ?", name)
}

func getOrganizationByID(db sqlx.Queryer, id uint) (*otf.Organization, error) {
	return getOrganization(db, "id = ?", id)
}

func getOrganization(db sqlx.Queryer, pred interface{}, args ...interface{}) (*otf.Organization, error) {
	query := sq.Select("*").From("organizations").Where(pred, args)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	return scanOrganization(db.QueryRowx(sql, args))
}

func scanOrganization(scannable StructScannable) (*otf.Organization, error) {
	type result struct {
		metadata

		Name  string
		Email string
	}

	res := result{}
	if err := scannable.StructScan(res); err != nil {
		return nil, err
	}

	org := otf.Organization{
		ID: res.ExternalID,
		Model: gorm.Model{
			ID:        res.ID,
			CreatedAt: res.CreatedAt,
			UpdatedAt: res.UpdatedAt,
		},
		Name:  res.Name,
		Email: res.Email,
	}

	return &org, nil
}
