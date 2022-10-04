package html

import (
	"errors"

	"github.com/spf13/pflag"
)

var (
	ErrOAuthCredentialsUnspecified = errors.New("no oauth credentials have been specified")
	ErrOAuthCredentialsIncomplete  = errors.New("must specify both client ID and client secret")
)

type OAuthCredentials struct {
	prefix       string
	clientID     string
	clientSecret string
}

func NewOAuthCredentials(prefix string) *OAuthCredentials {
	return &OAuthCredentials{prefix: prefix}
}

func (a *OAuthCredentials) Valid() error {
	if a.clientID != "" && a.clientSecret != "" {
		return nil
	}
	if a.clientID == "" && a.clientSecret == "" {
		return ErrOAuthCredentialsUnspecified
	}
	return ErrOAuthCredentialsIncomplete
}

// TODO: rename to something like ConfigureViaFlags
func (a *OAuthCredentials) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&a.clientID, a.clientIDFlag(), "", a.prefix+" Client ID")
	flags.StringVar(&a.clientSecret, a.clientSecretFlag(), "", a.prefix+" Client Secret")
}

func (a *OAuthCredentials) ClientID() string     { return a.clientID }
func (a *OAuthCredentials) ClientSecret() string { return a.clientSecret }

func (a *OAuthCredentials) clientIDFlag() string {
	return a.prefix + "-client-id"
}

func (a *OAuthCredentials) clientSecretFlag() string {
	return a.prefix + "-client-secret"
}
