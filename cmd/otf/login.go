package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	DummyToken = "dummy"
)

func LoginCommand(dirs Directories) *cobra.Command {
	var address string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to OTF",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewCredentialsStore(dirs)
			if err != nil {
				return err
			}

			if err := store.Save(address, DummyToken); err != nil {
				return err
			}

			fmt.Printf("Successfully added credentials for %s to %s\n", address, store)

			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", DefaultAddress, "Address of OTF instance")

	return cmd
}
