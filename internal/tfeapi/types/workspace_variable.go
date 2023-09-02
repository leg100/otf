package types

// WorkspaceVariable is a workspace variable.
type WorkspaceVariable struct {
	*Variable

	// Relations
	Workspace *Workspace `jsonapi:"relationship" json:"configurable"`
}
