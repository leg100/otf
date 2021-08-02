package sqlite

import (
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	driver "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var models = []interface{}{
	&Run{},
	&Apply{},
	&Plan{},
	&ConfigurationVersion{},
	&Organization{},
	&StateVersion{},
	&StateVersionOutput{},
	&Workspace{},
}

func New(path string, cfg *gorm.Config) (*gorm.DB, error) {
	db, err := gorm.Open(driver.Open(path), cfg)
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(models...); err != nil {
		return nil, err
	}

	return db, nil
}

// Gorm scopes: https://gorm.io/docs/advanced_query.html#Scopes

func paginate(opts tfe.ListOptions) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		ots.SanitizeListOptions(&opts)

		offset := (opts.PageNumber - 1) * opts.PageSize

		return db.Offset(offset).Limit(opts.PageSize)
	}
}
