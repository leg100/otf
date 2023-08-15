package variable

type (
	VariableSet struct {
		ID          string
		Key         string
		Value       string
		Description string
		Category    VariableCategory
		Sensitive   bool
		HCL         bool
		WorkspaceID string

		// OTF doesn't use this internally but the go-tfe integration tests
		// expect it to be a random value that changes on every update.
		VersionID string
	}
	CreateVariableSetOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool
	}
	UpdateVariableSetOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool
	}
)
