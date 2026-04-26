package execution

import "fmt"

type Mode string

const (
	RemoteMode Mode = "remote"
	LocalMode  Mode = "local"
	AgentMode  Mode = "agent"
)

func IsValidMode(mode Mode) error {
	switch mode {
	case RemoteMode, LocalMode, AgentMode:
		return nil
	default:
		return fmt.Errorf("invalid execution mode: %s", mode)
	}
}
