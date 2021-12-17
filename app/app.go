/*
Package app implements services, co-ordinating between the layers of the project.
*/
package app

import (
	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
)

type Application struct {
	organizationService         otf.OrganizationService
	workspaceService            otf.WorkspaceService
	stateVersionService         otf.StateVersionService
	configurationVersionService otf.ConfigurationVersionService
	runService                  otf.RunService
	planService                 otf.PlanService
	applyService                otf.ApplyService
	eventService                otf.EventService
}

type Config struct {
	inmem.CacheConfig
}

func NewApplication(logger logr.Logger, db otf.DB, cache *bigcache.BigCache) (*Application, error) {
	// Setup event broker
	eventService := inmem.NewEventService(logger)

	// Setup services
	orgService := NewOrganizationService(db.OrganizationStore(), logger, eventService)
	workspaceService := NewWorkspaceService(db.WorkspaceStore(), logger, orgService, eventService)
	stateVersionService := NewStateVersionService(db.StateVersionStore(), logger, workspaceService, cache)
	configurationVersionService := NewConfigurationVersionService(db.ConfigurationVersionStore(), logger, workspaceService, cache)
	runService := NewRunService(db.RunStore(), logger, workspaceService, configurationVersionService, eventService, db.PlanLogStore(), db.ApplyLogStore(), cache)
	planService := NewPlanService(db.RunStore(), db.PlanLogStore(), logger, eventService, cache)
	applyService := NewApplyService(db.RunStore(), db.ApplyLogStore(), logger, eventService, cache)

	return &Application{
		organizationService:         orgService,
		workspaceService:            workspaceService,
		stateVersionService:         stateVersionService,
		configurationVersionService: configurationVersionService,
		runService:                  runService,
		planService:                 planService,
		applyService:                applyService,
		eventService:                eventService,
	}, nil
}

func (app *Application) OrganizationService() otf.OrganizationService {
	return app.organizationService
}

func (app *Application) WorkspaceService() otf.WorkspaceService {
	return app.workspaceService
}

func (app *Application) StateVersionService() otf.StateVersionService {
	return app.stateVersionService
}

func (app *Application) ConfigurationVersionService() otf.ConfigurationVersionService {
	return app.configurationVersionService
}

func (app *Application) RunService() otf.RunService {
	return app.runService
}

func (app *Application) PlanService() otf.PlanService {
	return app.planService
}

func (app *Application) ApplyService() otf.ApplyService {
	return app.applyService
}

func (app *Application) EventService() otf.EventService {
	return app.eventService
}
