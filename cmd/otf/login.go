package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	DummyToken = "dummy"
)

func LoginCommand(store KVStore, address string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to OTF",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := store.Save(address, DummyToken); err != nil {
				return err
			}

			fmt.Printf("Successfully added credentials for %s to %s\n", address, store)

			return nil
		},
	}

	return cmd
}
