package main

import (
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func AgentCommand(factory http.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Agent management",
	}

	cmd.AddCommand(AgentTokenCommand(factory))

	return cmd
}

func AgentTokenCommand(factory http.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens",
		Short: "Agent token management",
	}

	cmd.AddCommand(AgentTokenNewCommand(factory))

	return cmd
}
