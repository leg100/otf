package main

import (
	"fmt"

	"github.com/leg100/otf/http"
	"github.com/spf13/cobra"
)

const (
	DummyToken = "dummy"
)

func LoginCommand(store KVStore) *cobra.Command {
	var address string

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

	cmd.Flags().StringVar(&address, "address", http.DefaultAddress, "Address of OTF instance")

	return cmd
}
