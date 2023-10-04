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
	"github.com/leg100/otf/internal/vcs"
)

type (
	// VCSProvider provides authenticated access to a VCS.
	VCSProvider struct {
		ID           string
		Name         string
		CreatedAt    time.Time
		Organization string   // name of OTF organization
		Kind         vcs.Kind // github/gitlab etc
		Hostname     string   // hostname of github/gitlab etc

		Token     *string                    // personal access token.
		GithubApp *github.InstallCredentials // mutually exclusive with Token.
	}

	newOptions struct {
		CreateOptions

		// map of vcs kind to hostname
		cloudHostnames map[vcs.Kind]string

		// Must be specified if GithubAppInstallID is non-nil
		github.GithubAppService

		// These fields are only needed for re-creating a vcs provider from a DB
		// query
		ID        *string
		CreatedAt *time.Time
	}

	UpdateOptions struct {
		Token *string
		Name  string
	}
)

func newProvider(ctx context.Context, opts newOptions) (*VCSProvider, error) {
	provider := &VCSProvider{
		ID:           internal.NewID("vcs"),
		Name:         opts.Name,
		CreatedAt:    internal.CurrentTimestamp(),
		Organization: opts.Organization,
		Kind:         opts.Kind,
	}
	host, ok := opts.cloudHostnames[opts.Kind]
	if !ok {
		return nil, fmt.Errorf("cloud kind %s has no hostname", opts.Kind)
	}
	provider.Hostname = host
	if opts.ID != nil {
		provider.ID = *opts.ID
	}
	if opts.CreatedAt != nil {
		provider.CreatedAt = *opts.CreatedAt
	}
	if opts.Token != nil {
		if err := provider.setToken(*opts.Token); err != nil {
			return nil, err
		}
	} else if opts.GithubAppInstallID != nil {
		app, err := opts.GetGithubApp(ctx)
		if err != nil {
			return nil, err
		}
		provider.GithubApp = &github.InstallCredentials{
			ID: *opts.GithubAppInstallID,
			AppCredentials: github.AppCredentials{
				ID:         app.ID,
				PrivateKey: app.PrivateKey,
			},
		}
	} else {
		return nil, errors.New("must specify either token or github app installation ID")
	}
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
			InstallCredentials: t.GithubApp,
		})
	} else if t.Token != nil {
		return vcs.GetPersonalTokenClient(t.Kind, t.Hostname, *t.Token)
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
