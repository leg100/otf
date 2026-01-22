package runner

type OperationClient struct {
	OperationClientUseToken

	Runs       runClient
	Workspaces workspaceClient
	Variables  variablesClient
	State      stateClient
	Configs    configClient
	Server     hostnameClient
	Jobs       operationJobsClient
}

type OperationClientUseToken interface {
	UseToken(string)
}
