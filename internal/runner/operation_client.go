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
	SSHKeys    sshKeyClient
}

func NewOperationClient(
	Runs runClient,
	Workspaces workspaceClient,
	Variables variablesClient,
	State stateClient,
	Configs configClient,
	Server hostnameClient,
	Jobs operationJobsClient,
	SSHKeys sshKeyClient,
) OperationClient {
	return OperationClient{
		Workspaces: Workspaces,
		Variables:  Variables,
		State:      State,
		Configs:    Configs,
		Runs:       Runs,
		Jobs:       Jobs,
		Server:     Server,
		SSHKeys:    SSHKeys,
	}
}
