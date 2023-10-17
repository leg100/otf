// Package vcsprovider is responsible for VCS providers
package vcsprovider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/leg100/otf/internal/vcs"
)

type (
	// VCSProvider provides authenticated access to a VCS.
	VCSProvider struct {
		ID           string
		Name         string
		CreatedAt    time.Time
		Organization string // name of OTF organization
		Hostname     string // hostname of github/gitlab etc

		Kind  vcs.Kind // github/gitlab etc. Not necessary if GithubApp is non-nil.
		Token *string  // personal access token.

		GithubApp *github.InstallCredentials // mutually exclusive with Token.

		skipTLSVerification bool // toggle skipping verification of VCS host's TLS cert.
	}

	// factory produces VCS providers
	factory struct {
		github.GithubAppService

		githubHostname      string
		gitlabHostname      string
		skipTLSVerification bool // toggle skipping verification of VCS host's TLS cert.
	}

	CreateOptions struct {
		Organization string
		Name         string
		Kind         *vcs.Kind

		// Specify either token or github app install ID
		Token              *string
		GithubAppInstallID *int64
	}

	UpdateOptions struct {
		Token *string
		Name  string
	}
)

func (f *factory) newProvider(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	var (
		creds *github.InstallCredentials
		err   error
	)
	if opts.GithubAppInstallID != nil {
		creds, err = f.GetInstallCredentials(ctx, *opts.GithubAppInstallID)
		if err != nil {
			return nil, err
		}
	}
	return f.newWithGithubCredentials(ctx, opts, creds)
}

func (f *factory) newWithGithubCredentials(ctx context.Context, opts CreateOptions, creds *github.InstallCredentials) (*VCSProvider, error) {
	provider := &VCSProvider{
		ID:                  internal.NewID("vcs"),
		Name:                opts.Name,
		CreatedAt:           internal.CurrentTimestamp(),
		Organization:        opts.Organization,
		skipTLSVerification: f.skipTLSVerification,
	}
	if opts.Token != nil {
		if opts.Kind == nil {
			return nil, errors.New("must specify both token and kind")
		}
		provider.Kind = *opts.Kind
		switch provider.Kind {
		case vcs.GithubKind:
			provider.Hostname = f.githubHostname
		case vcs.GitlabKind:
			provider.Hostname = f.gitlabHostname
		default:
			return nil, errors.New("no hostname found for vcs kind")
		}
		if err := provider.setToken(*opts.Token); err != nil {
			return nil, err
		}
	} else if creds != nil {
		provider.GithubApp = creds
		provider.Kind = vcs.GithubKind
		provider.Hostname = f.githubHostname
	} else {
		return nil, errors.New("must specify either token or github app installation ID")
	}
	return provider, nil
}

func (f *factory) fromDB(ctx context.Context, opts CreateOptions, creds *github.InstallCredentials, id string, createdAt time.Time) (*VCSProvider, error) {
	provider, err := f.newWithGithubCredentials(ctx, opts, creds)
	if err != nil {
		return nil, err
	}
	provider.ID = id
	provider.CreatedAt = createdAt
	return provider, nil
}

// String provides a human meaningful description of the vcs provider, using the
// name if set; otherwise a name is constructed using both the underlying cloud
// kind and the auth kind.
func (t *VCSProvider) String() string {
	if t.Name != "" {
		return t.Name
	}
	s := string(t.Kind)
	if t.Token != nil {
		s += " (token)"
	}
	if t.GithubApp != nil {
		s += " (app)"
	}
	return s
}

func (t *VCSProvider) NewClient() (vcs.Client, error) {
	if t.GithubApp != nil {
		return github.NewClient(github.ClientOptions{
			Hostname:            t.Hostname,
			InstallCredentials:  t.GithubApp,
			SkipTLSVerification: t.skipTLSVerification,
		})
	} else if t.Token != nil {
		opts := vcs.NewTokenClientOptions{
			Hostname:            t.Hostname,
			Token:               *t.Token,
			SkipTLSVerification: t.skipTLSVerification,
		}
		switch t.Kind {
		case vcs.GithubKind:
			return github.NewTokenClient(opts)
		case vcs.GitlabKind:
			return gitlab.NewTokenClient(opts)
		default:
			return nil, fmt.Errorf("unknown kind: %s", t.Kind)
		}
	} else {
		return nil, fmt.Errorf("missing credentials")
	}
}

func (t *VCSProvider) Update(opts UpdateOptions) error {
	t.Name = opts.Name
	if opts.Token != nil {
		if err := t.setToken(*opts.Token); err != nil {
			return err
		}
	}
	return nil
}

// LogValue implements slog.LogValuer.
func (t *VCSProvider) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", t.ID),
		slog.String("organization", t.Organization),
		slog.String("name", t.String()),
		slog.String("kind", string(t.Kind)),
	)
}

func (t *VCSProvider) setToken(token string) error {
	if token == "" {
		return fmt.Errorf("token: %w", internal.ErrEmptyValue)
	}
	t.Token = &token
	return nil
}
