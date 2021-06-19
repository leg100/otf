package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const (
	CredentialsPath = ".terraform.d/credentials.tfrc.json"
	DummyToken      = "dummy"
)

var (
	ErrMissingHostname = errors.New("--hostname must be set to the hostname of the OTS server")
)

type CredentialsConfig struct {
	Credentials map[string]TokenConfig `json:"credentials"`
}

type TokenConfig struct {
	Token string `json:"token"`
}

func LoginCommand(dirs Directories) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to OTS",
		RunE: func(cmd *cobra.Command, args []string) error {
			if hostname == "" {
				return ErrMissingHostname
			}

			// Construct full path to creds config
			home, err := dirs.UserHomeDir()
			if err != nil {
				return err
			}
			path := filepath.Join(home, CredentialsPath)

			// Construct credentials config obj
			config := CredentialsConfig{Credentials: make(map[string]TokenConfig)}

			// Read any existing file contents
			data, err := os.ReadFile(path)
			if err == nil {
				if err := json.Unmarshal(data, &config); err != nil {
					return err
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}

			// Update file with dummy token for OTS instance
			config.Credentials[hostname] = TokenConfig{
				Token: DummyToken,
			}
			data, err = json.MarshalIndent(&config, "", "  ")
			if err != nil {
				return err
			}

			// Ensure all parent directories of config file exist
			if err := os.MkdirAll(filepath.Dir(path), 0775); err != nil {
				return err
			}

			if err := os.WriteFile(path, data, 0600); err != nil {
				return err
			}

			fmt.Printf("Successfully added credentials to %s\n", path)

			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", os.Getenv("OTS_HOSTNAME"), "Name of server deployment")

	return cmd
}
