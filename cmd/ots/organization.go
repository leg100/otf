package main

import (
	"os"

	"github.com/spf13/cobra"
)

func OrganizationCommand() *cobra.Command {
	cfg := clientConfig{}

	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Organization management",
	}
	cmd.Flags().StringVar(&cfg.Address, "address", DefaultAddress, "Address of OTS server")
	cmd.Flags().StringVar(&cfg.Token, "token", os.Getenv("OTS_TOKEN"), "Authentication token")

	cmd.AddCommand(OrganizationNewCommand(&cfg))

	return cmd
}
