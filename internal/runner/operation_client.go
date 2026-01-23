package runner

// OperationClientCreator creates an OperationClient using the given job token.
type OperationClientCreator func(jobToken string) OperationClient

// OperationClient is a client for an operation to interact with services on
// behalf of its job.
type OperationClient struct {
	Runs       runClient
	Workspaces workspaceClient
	Variables  variablesClient
	State      stateClient
	Configs    configClient
	Server     hostnameClient
	Jobs       operationJobsClient
}
