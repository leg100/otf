package variable

type (
	WorkspaceVariable struct {
		*Variable
		WorkspaceID string
	}

	WorkspaceVariables []*Variable
)
