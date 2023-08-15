package variable

type (
	// WorkspaceVariable is a workspace-scoped variable.
	WorkspaceVariable struct {
		*Variable

		WorkspaceID string
	}
)
