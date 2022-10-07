package html

import (
	"context"

	"github.com/leg100/otf"
	"golang.org/x/oauth2"
)

type Cloud interface {
	NewDirectoryClient(context.Context, DirectoryClientOptions) (DirectoryClient, error)
	CloudConfig
}

type DirectoryClientOptions struct {
	Token *oauth2.Token

	*oauth2.Config
}

type DirectoryClient interface {
	GetUser(ctx context.Context) (string, error)
	ListTeams(ctx context.Context) ([]*otf.Team, error)
	ListOrganizations(ctx context.Context) ([]*otf.Organization, error)
}
