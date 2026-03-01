package runner

import (
	"context"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sshkey"
)

// OperationClientCreator creates an OperationClient using the given job token.
type OperationClientCreator func(jobToken string) OperationClient

// sshKeyClient fetches SSH key data for an operation.
type sshKeyClient interface {
	GetSSHKey(ctx context.Context, id resource.TfeID) (*sshkey.SSHKey, error)
}

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
