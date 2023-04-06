package tokens

const (
	userSessionKind     kind = "user_session"
	registrySessionKind kind = "registry_session"
	agentTokenKind      kind = "agent_token"
	userTokenKind       kind = "user_token"
)

// the kind of authentication token: user session, user token, agent token, etc
type kind string
