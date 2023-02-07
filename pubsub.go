package otf

import "context"

type PubSubDatabase interface {
	GetOrganizationByID(context.Context, string) (Organization, error)
	GetRun(context.Context, string) (Run, error)
	GetWorkspace(context.Context, string) (Workspace, error)
	GetChunkByID(context.Context, int) (Chunk, error)
}
