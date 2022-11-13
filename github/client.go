package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
)

type client struct {
	client *github.Client
}

func NewEnterpriseClient(hostname string, httpClient *http.Client) (*github.Client, error) {
	return github.NewEnterpriseClient(
		"https://"+hostname,
		"https://"+hostname,
		httpClient)
}

func (g *client) GetUser(ctx context.Context) (*otf.User, error) {
	guser, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	var orgs []*otf.Organization
	var teams []*otf.Team

	gorgs, _, err := g.client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, err
	}
	for _, gorg := range gorgs {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(gorg.GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)

		// Determine if they are an admin; if so, add them to the owners team.
		membership, _, err := g.client.Organizations.GetOrgMembership(ctx, "", org.Name())
		if err != nil {
			return nil, err
		}
		if membership.GetRole() == "admin" {
			teams = append(teams, otf.NewTeam("owners", org))
		}
	}

	gteams, _, err := g.client.Teams.ListUserTeams(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, gteam := range gteams {
		org, err := otf.NewOrganization(otf.OrganizationCreateOptions{
			Name: otf.String(gteam.GetOrganization().GetLogin()),
		})
		if err != nil {
			return nil, err
		}
		teams = append(teams, otf.NewTeam(gteam.GetName(), org))
	}

	user := otf.NewUser(guser.GetLogin(), otf.WithOrganizationMemberships(orgs...), otf.WithTeamMemberships(teams...))
	return user, nil
}

func (g *client) GetRepository(ctx context.Context, identifier string) (*otf.Repo, error) {
	owner, name, found := strings.Cut(identifier, "/")
	if !found {
		return nil, fmt.Errorf("malformed identifier: %s", identifier)
	}
	repo, _, err := g.client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	return &otf.Repo{
		Identifier: repo.GetFullName(),
		HTTPURL:    repo.GetURL(),
		Branch:     repo.GetDefaultBranch(),
	}, nil
}

func (g *client) ListRepositories(ctx context.Context, opts otf.ListOptions) (*otf.RepoList, error) {
	repos, resp, err := g.client.Repositories.List(ctx, "", &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			Page:    opts.SanitizedPageNumber(),
			PerPage: opts.SanitizedPageSize(),
		},
	})
	if err != nil {
		return nil, err
	}

	// convert to common repo type before returning
	var items []*otf.Repo
	for _, repo := range repos {
		items = append(items, &otf.Repo{
			Identifier: repo.GetFullName(),
			HTTPURL:    repo.GetURL(),
			Branch:     repo.GetDefaultBranch(),
		})
	}

	return &otf.RepoList{
		Items:      items,
		Pagination: otf.NewPagination(opts, resp.LastPage*opts.SanitizedPageSize()),
	}, nil
}

func (g *client) GetRepoTarball(ctx context.Context, repo *otf.VCSRepo) ([]byte, error) {
	opts := github.RepositoryContentGetOptions{
		Ref: repo.Branch,
	}
	link, _, err := g.client.Repositories.GetArchiveLink(ctx, repo.Owner(), repo.Repo(), github.Tarball, &opts, true)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.Client().Get(link.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// github tarball contains a parent directory of the format
	// <owner>-<repo>-<commit>. We need a tarball without this parent directory,
	// so we untar it to a temp dir, then tar it up the contents of the parent
	// directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("github-%s-%s-*", repo.Owner(), repo.Repo()))
	if err != nil {
		return nil, err
	}
	if err := otf.Unpack(resp.Body, untarpath); err != nil {
		return nil, err
	}
	contents, err := os.ReadDir(untarpath)
	if err != nil {
		return nil, err
	}
	if len(contents) != 1 {
		return nil, fmt.Errorf("malformed tarball archive")
	}
	parentDir := path.Join(untarpath, contents[0].Name())
	return otf.Pack(parentDir)
}
