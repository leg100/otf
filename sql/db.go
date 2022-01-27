package sql

import (
	"fmt"
	"time"

	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/iancoleman/strcase"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

type db struct {
	*sqlx.DB

	organizationStore         otf.OrganizationStore
	workspaceStore            otf.WorkspaceStore
	stateVersionStore         otf.StateVersionStore
	configurationVersionStore otf.ConfigurationVersionStore
	runStore                  otf.RunStore
	planLogStore              otf.PlanLogStore
	applyLogStore             otf.ApplyLogStore
	userStore                 otf.UserStore
}

func New(logger logr.Logger, path string, cache *bigcache.BigCache, sessionExpiry time.Duration) (otf.DB, error) {
	sqlxdb, err := sqlx.Connect("postgres", path)
	if err != nil {
		return nil, err
	}

	// Map struct field names from CamelCase to snake_case.
	sqlxdb.MapperFunc(strcase.ToSnake)

	if err := migrate(logger, sqlxdb.DB); err != nil {
		return nil, err
	}

	db := db{
		DB:                        sqlxdb,
		organizationStore:         NewOrganizationDB(sqlxdb),
		workspaceStore:            NewWorkspaceDB(sqlxdb),
		stateVersionStore:         NewStateVersionDB(sqlxdb),
		configurationVersionStore: NewConfigurationVersionDB(sqlxdb),
		runStore:                  NewRunDB(sqlxdb),
		planLogStore:              NewPlanLogDB(sqlxdb),
		applyLogStore:             NewApplyLogDB(sqlxdb),
		userStore:                 NewUserDB(sqlxdb, sessionExpiry),
	}

	if cache != nil {
		db.planLogStore, err = inmem.NewChunkProxy(cache, db.planLogStore)
		if err != nil {
			return nil, fmt.Errorf("unable to instantiate plan log store: %w", err)
		}

		db.applyLogStore, err = inmem.NewChunkProxy(cache, db.applyLogStore)
		if err != nil {
			return nil, fmt.Errorf("unable to instantiate apply log store: %w", err)
		}
	}

	return db, nil
}

func (db db) Handle() *sqlx.DB                         { return db.DB }
func (db db) Close() error                             { return db.DB.Close() }
func (db db) OrganizationStore() otf.OrganizationStore { return db.organizationStore }
func (db db) WorkspaceStore() otf.WorkspaceStore       { return db.workspaceStore }
func (db db) StateVersionStore() otf.StateVersionStore { return db.stateVersionStore }
func (db db) ConfigurationVersionStore() otf.ConfigurationVersionStore {
	return db.configurationVersionStore
}
func (db db) RunStore() otf.RunStore           { return db.runStore }
func (db db) PlanLogStore() otf.PlanLogStore   { return db.planLogStore }
func (db db) ApplyLogStore() otf.ApplyLogStore { return db.applyLogStore }
func (db db) UserStore() otf.UserStore         { return db.userStore }
