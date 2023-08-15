package variable

type (
	// VariableSet is a set of variables
	VariableSet struct {
		ID           string
		Name         string
		Description  string
		Global       bool
		Variables    []*Variable
		Workspaces   []string // workspace IDs
		Organization string   // org name
	}
	CreateVariableSetOptions struct {
		Name        string
		Description string
		Global      bool
	}
	UpdateVariableSetOptions struct {
		Name        *string
		Description *string
		Global      *bool
	}
)
