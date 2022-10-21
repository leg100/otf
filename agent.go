package otf

// LocalAgent identifies the built-in agent for authentication purposes.
// (Built-in agent is distinct from 'proper' agents deployed outside of otfd).
type LocalAgent struct{}

// CanAccessSite - local agent needs to retrieve runs across site
func (*LocalAgent) CanAccessSite(action Action) bool { return true }

// CanAccessOrganization - unlike proper agents, the local agent can access any
// organization.
func (*LocalAgent) CanAccessOrganization(Action, string) bool { return true }

// CanAccessWorkspace - unlike proper agents, the local agent can acess any
// workspace
func (*LocalAgent) CanAccessWorkspace(Action, *WorkspacePolicy) bool { return true }

func (*LocalAgent) String() string { return "local-agent" }
func (*LocalAgent) ID() string     { return "local-agent" }
