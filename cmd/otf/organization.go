package main

import (
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func OrganizationCommand(factory http.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Organization management",
	}

	cmd.AddCommand(OrganizationNewCommand(factory))

	return cmd
}
