package html

import (
	"errors"

	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func (a *OAuthCredentials) Valid() error {
	if a.clientID != "" && a.clientSecret != "" {
		return nil
	}
	if a.clientID == "" && a.clientSecret == "" {
		return ErrOAuthCredentialsUnspecified
	}
	return ErrOAuthCredentialsIncomplete
}

func (a *OAuthCredentials) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&a.clientID, a.clientIDFlag(), "", a.prefixTitle()+" client ID")
	flags.StringVar(&a.clientSecret, a.clientSecretFlag(), "", a.prefixTitle()+" client secret")
}

func (a *OAuthCredentials) ClientID() string     { return a.clientID }
func (a *OAuthCredentials) ClientSecret() string { return a.clientSecret }

func (a *OAuthCredentials) clientIDFlag() string {
	return a.prefix + "-client-id"
}

func (a *OAuthCredentials) clientSecretFlag() string {
	return a.prefix + "-client-secret"
}

func (a *OAuthCredentials) prefixTitle() string {
	return cases.Title(language.English).String(a.prefix)
}
