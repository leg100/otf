/*
Package app implements services, co-ordinating between the layers of the project.
*/
package app

import (
	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
)

var (
	_ otf.Application = (*Application)(nil)
)

type Application struct {
	organizationService         otf.OrganizationService
	workspaceService            otf.WorkspaceService
	stateVersionService         otf.StateVersionService
	configurationVersionService otf.ConfigurationVersionService
	runService                  otf.RunService
	eventService                otf.EventService
	userService                 otf.UserService
}

func NewApplication(logger logr.Logger, db *sql.DB, cache *bigcache.BigCache) (*Application, error) {
	// Setup event broker
	eventService := inmem.NewEventService(logger)

	// Setup services
	orgService, err := NewOrganizationService(db, logger, eventService)
	if err != nil {
		return nil, err
	}
	workspaceService, err := NewWorkspaceService(db, logger, orgService, eventService)
	if err != nil {
		return nil, err
	}
	stateVersionService := NewStateVersionService(db, logger, cache)
	configurationVersionService := NewConfigurationVersionService(db, logger, cache)
	runService, err := NewRunService(db, logger, workspaceService, configurationVersionService, eventService, cache)
	if err != nil {
		return nil, err
	}
	userService := NewUserService(logger, db)

	return &Application{
		organizationService:         orgService,
		workspaceService:            workspaceService,
		stateVersionService:         stateVersionService,
		configurationVersionService: configurationVersionService,
		runService:                  runService,
		eventService:                eventService,
		userService:                 userService,
	}, nil
}

func (app *Application) OrganizationService() otf.OrganizationService { return app.organizationService }
func (app *Application) WorkspaceService() otf.WorkspaceService       { return app.workspaceService }
func (app *Application) StateVersionService() otf.StateVersionService { return app.stateVersionService }
func (app *Application) ConfigurationVersionService() otf.ConfigurationVersionService {
	return app.configurationVersionService
}
func (app *Application) RunService() otf.RunService     { return app.runService }
func (app *Application) EventService() otf.EventService { return app.eventService }
func (app *Application) UserService() otf.UserService   { return app.userService }
