package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:           "ots",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(LoginCommand(&SystemDirectories{}))

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
