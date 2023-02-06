package otf

type AgentToken interface {
	Token() string

	Subject
}
