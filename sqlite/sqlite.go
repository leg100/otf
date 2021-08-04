package sqlite

import (
	"github.com/leg100/go-tfe"
	gormzerolog "github.com/leg100/gorm-zerolog"
	"github.com/leg100/ots"
	"github.com/rs/zerolog"
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

type Option func(*gorm.Config)

func WithZeroLogger(logger *zerolog.Logger) Option {
	return func(cfg *gorm.Config) {
		cfg.Logger = &gormzerolog.Logger{Zlog: *logger}
	}
}

func New(path string, opts ...Option) (*gorm.DB, error) {
	cfg := &gorm.Config{}
	for _, o := range opts {
		o(cfg)
	}

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
