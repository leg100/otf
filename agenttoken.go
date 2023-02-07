package otf

type AgentToken interface {
	Token() string

	Subject
}

type CreateAgentTokenOptions struct {
	Organization string `schema:"organization_name,required"`
	Description  string `schema:"description,required"`
}
