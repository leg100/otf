package main

import (
	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

func RunCommand(factory http.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runs",
		Short: "Runs management",
	}

	cmd.AddCommand(RunDownloadCommand(factory))

	return cmd
}
