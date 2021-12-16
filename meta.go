package otf

type DB interface {
	GetOrganizationStore() OrganizationStore
	GetWorkspaceStore() WorkspaceStore
	GetStateVersionStore() StateVersionStore
	GetConfigurationVersionStore() ConfigurationVersionStore
	GetRunStore() RunStore
}

type MetaService interface {
	GetOrganizationService() OrganizationService
	GetWorkspaceService() WorkspaceService
	GetStateVersionService() StateVersionService
	GetConfigurationVersionService() ConfigurationVersionService
	GetRunService() RunService
	GetPlanService() PlanService
	GetApplyService() ApplyService
	GetEventService() EventService
	//GetCacheService() *CacheService
}

func NewMetaService() MetaService {
	return nil
}
