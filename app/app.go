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
	planService                 otf.PlanService
	applyService                otf.ApplyService
	eventService                otf.EventService
	userService                 otf.UserService
	jobService                  otf.JobService
}

func NewApplication(logger logr.Logger, db *sql.DB, cache *bigcache.BigCache) (*Application, error) {
	// Setup event broker
	eventService := inmem.NewEventService(logger)

	// Setup services
	orgService := NewOrganizationService(db, logger, eventService)
	workspaceService, err := NewWorkspaceService(db, logger, orgService, eventService)
	if err != nil {
		return nil, err
	}
	stateVersionService := NewStateVersionService(db, logger, cache)
	configurationVersionService := NewConfigurationVersionService(db, logger, cache)
	logService, err := newLogService(db, logger, cache)
	if err != nil {
		return nil, err
	}
	runService := NewRunService(db, logger, workspaceService, configurationVersionService, eventService, logService, cache)
	jobService := newJobService(db, logger, eventService, logService, runService)
	planService := NewPlanService(db, logger, runService)
	applyService := NewApplyService(db, logger, jobService)
	userService := NewUserService(logger, db)

	return &Application{
		organizationService:         orgService,
		workspaceService:            workspaceService,
		stateVersionService:         stateVersionService,
		configurationVersionService: configurationVersionService,
		runService:                  runService,
		planService:                 planService,
		applyService:                applyService,
		eventService:                eventService,
		userService:                 userService,
		jobService:                  jobService,
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

func (app *Application) UserService() otf.UserService {
	return app.userService
}

func (app *Application) JobService() otf.JobService {
	return app.jobService
}
