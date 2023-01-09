package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http"
)

var _ KVStore = (CredentialsStore)("")

const (
	CredentialsPath = ".terraform.d/credentials.tfrc.json"
)

// CredentialsStore is a JSON file in a user's home dir that stores tokens for
// one or more TFE-type hosts
type CredentialsStore string

type CredentialsConfig struct {
	Credentials map[string]TokenConfig `json:"credentials"`
}

type TokenConfig struct {
	Token string `json:"token"`
}

// NewCredentialsStore is a contructor for CredentialsStore
func NewCredentialsStore() (CredentialsStore, error) {
	// Construct full path to creds config
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, CredentialsPath)

	return CredentialsStore(path), nil
}

// Load retrieves the token for hostname
func (c CredentialsStore) Load(hostname string) (string, error) {
	hostname, err := http.SanitizeHostname(hostname)
	if err != nil {
		return "", err
	}

	// TF v1.2.0 and later supports environment variables:
	//
	// https://developer.hashicorp.com/terraform/cli/config/config-file#environment-variable-credentials
	//
	// They take precendence over reading from file.
	if token, ok := os.LookupEnv(otf.HostnameCredentialEnv(hostname)); ok {
		return token, nil
	}

	config, err := c.read()
	if err != nil {
		return "", fmt.Errorf("reading credentials config: %w", err)
	}

	tokenConfig, ok := config.Credentials[hostname]
	if !ok {
		return "", fmt.Errorf("credentials for %s not found in %s", hostname, c)
	}

	return tokenConfig.Token, nil
}

// Save saves the token for the given hostname to the store, overwriting any
// existing tokens for the hostname.
func (c CredentialsStore) Save(hostname, token string) error {
	hostname, err := http.SanitizeHostname(hostname)
	if err != nil {
		return err
	}

	config, err := c.read()
	if err != nil {
		return err
	}

	config.Credentials[hostname] = TokenConfig{
		Token: token,
	}

	if err := c.write(config); err != nil {
		return err
	}

	return nil
}

func (c CredentialsStore) read() (*CredentialsConfig, error) {
	// Construct credentials config obj
	config := CredentialsConfig{Credentials: make(map[string]TokenConfig)}

	// Read any existing file contents
	data, err := os.ReadFile(string(c))
	if err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &config, nil
}

func (c CredentialsStore) write(config *CredentialsConfig) error {
	data, err := json.MarshalIndent(&config, "", "  ")
	if err != nil {
		return err
	}

	// Ensure all parent directories of config file exist
	if err := os.MkdirAll(filepath.Dir(string(c)), 0o775); err != nil {
		return err
	}

	if err := os.WriteFile(string(c), data, 0o600); err != nil {
		return err
	}

	return nil
}

// LoadCredentials is passed as an arg to http.NewClient to instruct it to load
// the auth token from the credentials store.
func LoadCredentials(cfg *http.Config) error {
	creds, err := NewCredentialsStore()
	if err != nil {
		return err
	}

	cfg.Token, err = creds.Load(cfg.Address)
	if err != nil {
		return err
	}

	return nil
}
