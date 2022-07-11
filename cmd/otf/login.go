package main

import (
	"fmt"
	"io"
	"net/url"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func LoginCommand(store KVStore, address *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to OTF",
		RunE: func(cmd *cobra.Command, args []string) error {
			u := url.URL{
				Scheme: "https://",
				Host:   *address,
				Path:   "/profile/tokens",
			}

			// disable browser lib printing to stdout
			browser.Stdout = io.Discard
			if err := browser.OpenURL(u.String()); err != nil {
				return err
			}

			fmt.Println("Opened browser for you to create a new token in the web app")
			fmt.Println()

			fmt.Printf("Enter token: ")
			var token string
			tokenLen, err := fmt.Scanln(&token)
			if err != nil {
				return err
			}
			if tokenLen != 32 {
				return fmt.Errorf("token must be 32 characters long")
			}

			if err := store.Save(*address, token); err != nil {
				return err
			}

			fmt.Printf("Successfully added credentials for %s to %s\n", *address, store)

			return nil
		},
	}

	return cmd
}
