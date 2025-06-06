package vcs

import (
	"context"

	"github.com/a-h/templ"
)

const (
	ForgejoKind     Kind = "forgejo"
	GithubTokenKind Kind = "github-token"
	GithubAppKind   Kind = "github-app"
	GitlabKind      Kind = "gitlab"
)

// Kind of vcs hosting provider
type Kind string

func KindPtr(k Kind) *Kind { return &k }

type ProviderKind struct {
	// Kind distinguishes this kind from other kinds
	Kind Kind
	// Name provides a human meaningful identification of this provider kind.
	Name string
	// Hostname is the hostname of the VCS host, not including scheme or path.
	Hostname string
	// Icon renders a icon distinguishing the VCS host kind.
	Icon templ.Component
	// TokenKind provides info about the token the provider expects. Mutually
	// exclusive with InstallationKind.
	TokenKind *TokenKind
	// InstallationKind provides info about installations for this ProviderKind.
	// Mutually exclusive with TokenKind.
	InstallationKind InstallationKind
	// NewClient constructs a client implementation.
	NewClient func(context.Context, Config) (Client, error)
}

type TokenKind struct {
	// TokenDescription renders a helpful description of what is expected of the
	// token, e.g. what permissions it should possess.
	Description templ.Component
}

type InstallationKind interface {
	// ListInstallations retrieves a list of installations.
	ListInstallations(context.Context) (ListInstallationsResult, error)
	// GetInstallation retrieves an installation by its ID.
	GetInstallation(context.Context, int64) (Installation, error)
}

type ListInstallationsResult struct {
	// InstallationLink is a link to the site where a user can create an
	// installation.
	InstallationLink templ.SafeURL
	// Results is a map of IDs of existing installations keyed by a human
	// meaningful name.
	Results []Installation
}

type Installation struct {
	ID           int64
	AppID        int64
	Username     *string
	Organization *string
}

func (i Installation) String() string {
	if i.Organization != nil {
		return "org/" + *i.Organization
	}
	return "user/" + *i.Username
}
