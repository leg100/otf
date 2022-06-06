package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

type db struct {
	*pgxpool.Pool

	organizationStore         otf.OrganizationStore
	workspaceStore            otf.WorkspaceStore
	stateVersionStore         otf.StateVersionStore
	configurationVersionStore otf.ConfigurationVersionStore
	runStore                  otf.RunStore
	planLogStore              otf.PlanLogStore
	applyLogStore             otf.ApplyLogStore
	userStore                 otf.UserStore
	sessionStore              otf.SessionStore
	tokenStore                otf.TokenStore
}

func New(logger logr.Logger, path string, cache *bigcache.BigCache, sessionExpiry time.Duration) (otf.DB, error) {
	conn, err := pgxpool.Connect(context.Background(), path)
	if err != nil {
		return nil, err
	}

	if err := migrate(logger, path); err != nil {
		return nil, err
	}

	db := db{
		Pool:                      conn,
		organizationStore:         NewOrganizationDB(conn),
		workspaceStore:            NewWorkspaceDB(conn),
		stateVersionStore:         NewStateVersionDB(conn),
		configurationVersionStore: NewConfigurationVersionDB(conn),
		runStore:                  newRunDB(conn),
		planLogStore:              NewPlanLogDB(conn),
		applyLogStore:             NewApplyLogDB(conn),
		userStore:                 NewUserDB(conn),
		sessionStore:              NewSessionDB(conn, sessionExpiry),
		tokenStore:                NewTokenDB(conn),
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

func (db db) Close() error                             { db.Pool.Close(); return nil }
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
func (db db) SessionStore() otf.SessionStore   { return db.sessionStore }
func (db db) TokenStore() otf.TokenStore       { return db.tokenStore }
