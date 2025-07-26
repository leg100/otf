package forgejo

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

type Client struct {
	client *forgejo.Client
}

func NewTokenClient(opts vcs.NewTokenClientOptions) (vcs.Client, error) {
	options := make([]forgejo.ClientOption, 0, 2)
	options = append(options, forgejo.SetToken(opts.Token))
	if opts.SkipTLSVerification {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		client := &http.Client{
			Transport: transport,
		}
		options = append(options, forgejo.SetHTTPClient(client))
	}
	rv, err := forgejo.NewClient(opts.BaseURL.String(), options...)
	if err != nil {
		return nil, err
	}
	return &Client{client: rv}, nil
}

func (c *Client) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]vcs.Repo, error) {
	found := map[vcs.Repo]time.Time{}

	// search for repos the user owns
	err := c.findReposOwned(found)
	if err != nil {
		return nil, err
	}

	// search for repos in orgs which the user owns
	err = c.findOrgReposOwned(found)
	if err != nil {
		return nil, err
	}

	// sort by updated time (desc)
	rv := make([]vcs.Repo, 0, len(found))
	for fullname := range found {
		rv = append(rv, fullname)
	}
	slices.SortFunc(rv, func(A, B vcs.Repo) int {
		tA := found[A]
		tB := found[B]
		return tB.Compare(tA)
	})

	return rv, nil
}

func (c *Client) findReposOwned(found map[vcs.Repo]time.Time) error {
	opt := forgejo.ListReposOptions{
		ListOptions: forgejo.ListOptions{
			Page:     0,
			PageSize: 50,
		},
	}
	resp := &forgejo.Response{NextPage: 1, LastPage: 1}
	for resp.LastPage > opt.Page {
		opt.Page = resp.NextPage
		var repolist []*forgejo.Repository
		var err error
		repolist, resp, err = c.client.ListMyRepos(opt)
		if err != nil {
			return err
		}
		for _, forgejoRepo := range repolist {
			if forgejoRepo.Permissions.Admin {
				repo, err := vcs.NewRepo(forgejoRepo.Owner.UserName, forgejoRepo.Name)
				if err != nil {
					return err
				}
				found[repo] = forgejoRepo.Updated
			}
		}
	}
	return nil
}

func (c *Client) findOrgReposOwned(found map[vcs.Repo]time.Time) error {
	// find all teams the user is a member of (paginated)
	var teamids []int64
	userteamsopt := forgejo.ListTeamsOptions{
		ListOptions: forgejo.ListOptions{
			Page:     0,
			PageSize: 50,
		},
	}
	resp := &forgejo.Response{NextPage: 1, LastPage: 1}
	for resp.LastPage > userteamsopt.Page {
		userteamsopt.Page = resp.NextPage
		var rv []*forgejo.Team
		var err error
		rv, resp, err = c.client.ListMyTeams(&userteamsopt)
		if err != nil {
			return err
		}
		for _, team := range rv {
			// save the teams with "admin" permissions
			if team.Permission == forgejo.AccessModeAdmin {
				teamids = append(teamids, team.ID)
			}
		}
	}

	// find repos belonging to those teams (paginated)
	for _, teamid := range teamids {
		listteamrepoopt := forgejo.ListTeamRepositoriesOptions{
			ListOptions: forgejo.ListOptions{
				Page:     0,
				PageSize: 50,
			},
		}
		resp := &forgejo.Response{NextPage: 1, LastPage: 1}
		for resp.LastPage > listteamrepoopt.Page {
			listteamrepoopt.Page = resp.NextPage
			var rv []*forgejo.Repository
			var err error
			rv, resp, err = c.client.ListTeamRepositories(teamid, listteamrepoopt)
			if err != nil {
				return err
			}
			for _, forgejoRepo := range rv {
				repo, err := vcs.NewRepo(forgejoRepo.Owner.UserName, forgejoRepo.Name)
				if err != nil {
					return err
				}
				found[repo] = forgejoRepo.Updated
			}
		}
	}

	return nil
}

func (c *Client) GetDefaultBranch(ctx context.Context, identifier string) (string, error) {
	parts := strings.Split(identifier, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", identifier)
	}
	owner, reponame := parts[0], parts[1]
	repo, _, err := c.client.GetRepo(owner, reponame)
	if err != nil {
		return "", err
	}
	return repo.DefaultBranch, nil
}

// map from otf/vcs event names to forgejo event names
func vcsOptToStringSlice(opts vcs.CreateWebhookOptions) ([]string, error) {
	// https://codeberg.org/forgejo/forgejo/src/branch/forgejo/modules/webhook/type.go
	rv := make([]string, 0, len(opts.Events))
	for _, event := range opts.Events {
		lookup, ok := map[string]string{
			"pull": "pull_request",
			"push": "push",
			"tag":  "push",
		}[string(event)]
		if !ok {
			return nil, fmt.Errorf("forgejo does not have an event type corresponding to '%s'", event)
		}
		rv = append(rv, lookup)
	}
	return rv, nil
}

// map from forgejo event names to otf/vcs event names
func stringSliceToVcs(es []string) ([]vcs.EventType, error) {
	// https://codeberg.org/forgejo/forgejo/src/branch/forgejo/modules/webhook/type.go
	rv := make([]vcs.EventType, 0, len(es))
	for _, event := range es {
		lookup, ok := map[string]vcs.EventType{
			"pull_request": "pull",
			"push":         "push",
		}[string(event)]
		if !ok {
			return nil, fmt.Errorf("otf does not have an event type corresponding to '%s'", event)
		}
		rv = append(rv, lookup)
	}
	return rv, nil
}

func (c *Client) CreateWebhook(ctx context.Context, opts vcs.CreateWebhookOptions) (string, error) {
	events, err := vcsOptToStringSlice(opts)
	if err != nil {
		return "", err
	}
	opt := forgejo.CreateHookOption{
		Type: forgejo.HookTypeGitea,
		Config: map[string]string{
			"content_type": "json",
			"url":          opts.Endpoint,
			"secret":       opts.Secret, // used for for signatures
		},
		Events: events,
		Active: true,
	}
	wh, _, err := c.client.CreateRepoHook(opts.Repo.Owner(), opts.Repo.Name(), opt)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(wh.ID, 16), nil
}

func (c *Client) UpdateWebhook(ctx context.Context, id string, opts vcs.UpdateWebhookOptions) error {
	idint, err := strconv.ParseInt(id, 16, 64)
	if err != nil {
		return err
	}
	events, err := vcsOptToStringSlice(vcs.CreateWebhookOptions(opts))
	if err != nil {
		return err
	}
	opt := forgejo.EditHookOption{
		Config: map[string]string{
			"content_type": "text/json",
			"url":          opts.Endpoint,
			"secret":       opts.Secret, // used for for signatures
		},
		Events: events,
	}
	_, err = c.client.EditRepoHook(opts.Repo.Owner(), opts.Repo.Name(), idint, opt)
	return err
}
func (c *Client) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	idint, err := strconv.ParseInt(opts.ID, 16, 64)
	if err != nil {
		return vcs.Webhook{}, err
	}
	wh, resp, err := c.client.GetRepoHook(opts.Repo.Owner(), opts.Repo.Name(), idint)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return vcs.Webhook{}, internal.ErrResourceNotFound
		}
		return vcs.Webhook{}, err
	}
	events, err := stringSliceToVcs(wh.Events)
	if err != nil {
		return vcs.Webhook{}, err
	}
	return vcs.Webhook{
		ID:       opts.ID,
		Repo:     opts.Repo,
		Events:   events,
		Endpoint: wh.URL,
	}, nil
}
func (c *Client) DeleteWebhook(ctx context.Context, opts vcs.DeleteWebhookOptions) error {
	idint, err := strconv.ParseInt(opts.ID, 16, 64)
	if err != nil {
		return err
	}
	_, err = c.client.DeleteRepoHook(opts.Repo.Owner(), opts.Repo.Name(), idint)
	return err
}

func (c *Client) fullyQualifyRef(owner, reponame, ref string) (string, error) {
	branches := map[string]string{}
	tags := map[string]string{}
	refs, _, err := c.client.GetRepoRefs(owner, reponame, "")
	if err != nil {
		return "", err
	}
	for _, ref := range refs {
		fullname := ref.Ref
		parts := strings.SplitN(fullname, "/", 3)
		if strings.HasPrefix(fullname, "refs/heads/") {
			branches[parts[2]] = fullname
		} else if strings.HasPrefix(fullname, "refs/tags/") {
			tags[parts[2]] = fullname
		}
	}

	// prefer tags, presumably they are more noteworthy
	if fullname, ok := tags[ref]; ok {
		return fullname, nil
	}
	// try looking up by branch
	if fullname, ok := branches[ref]; ok {
		return fullname, nil
	}
	// either ref is an SHA, or a branch/tag could not be found
	return ref, nil
}

func (c *Client) GetRepoTarball(ctx context.Context, opts vcs.GetRepoTarballOptions) ([]byte, string, error) {
	var (
		owner = opts.Repo.Owner()
		name  = opts.Repo.Name()
		ref   string
	)
	if opts.Ref != nil {
		ref = *opts.Ref
	}
	if ref == "" {
		// nil means default branch
		repo, _, err := c.client.GetRepo(owner, name)
		if err != nil {
			return nil, "", err
		}

		ref = repo.DefaultBranch
	}
	fqref, err := c.fullyQualifyRef(owner, name, ref)
	if err != nil {
		return nil, "", err
	}

	tarball, _, err := c.client.GetArchive(opts.Repo.Owner(), opts.Repo.Name(), ref, forgejo.TarGZArchive)
	if err != nil {
		return nil, "", fmt.Errorf("GetArchive(\"%s\", \"%s\", \"%s\", \"%s\") failed: %v", owner, name, ref, forgejo.TarGZArchive, err)
	}

	// Forgejo tarball contents are contained within a top-level directory
	// named after the repo. We want the tarball without this directory,
	// so we re-tar the contents without the top-level directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("forgejo-%s-%s-*", owner, name))
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = os.RemoveAll(untarpath)
	}()

	if err := internal.Unpack(bytes.NewReader(tarball), untarpath); err != nil {
		return nil, "", err
	}
	contents, err := os.ReadDir(untarpath)
	if err != nil {
		return nil, "", err
	}
	if len(contents) != 1 {
		return nil, "", fmt.Errorf("expected only one top-level directory; instead got %s", contents)
	}
	dir := contents[0].Name()
	tarball, err = internal.Pack(path.Join(untarpath, dir))
	if err != nil {
		return nil, "", err
	}
	return tarball, fqref, nil
}

func vcsStateToForgejo(s vcs.Status) forgejo.StatusState {
	return map[vcs.Status]forgejo.StatusState{
		vcs.PendingStatus: forgejo.StatusPending,
		vcs.SuccessStatus: forgejo.StatusSuccess,
		vcs.ErrorStatus:   forgejo.StatusError,
		vcs.FailureStatus: forgejo.StatusFailure,
	}[s]
}
func (c *Client) SetStatus(ctx context.Context, opts vcs.SetStatusOptions) error {
	opt := forgejo.CreateStatusOption{
		State:       vcsStateToForgejo(opts.Status),
		TargetURL:   opts.TargetURL,
		Description: opts.Description,
		Context:     "otf",
	}
	_, _, err := c.client.CreateStatus(opts.Repo.Owner(), opts.Repo.Name(), opts.Ref, opt)
	return err
}

func (c *Client) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error) {
	opt := forgejo.ListRepoTagsOptions{
		ListOptions: forgejo.ListOptions{
			Page:     0,
			PageSize: 50,
		},
	}
	resp := &forgejo.Response{NextPage: 1, LastPage: 1}
	rv := []string{}
	for resp.LastPage > opt.Page {
		opt.Page = resp.NextPage
		var tags []*forgejo.Tag
		var err error
		tags, resp, err = c.client.ListRepoTags(opts.Repo.Owner(), opts.Repo.Name(), opt)
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			if len(opts.Prefix) == 0 || strings.HasPrefix(tag.Name, opts.Prefix) {
				rv = append(rv, fmt.Sprintf("tags/%s", tag.Name))
			}
		}
	}
	return rv, nil
}

// ListPullRequestFiles returns the paths of files that are modified in the pull request
func (c *Client) ListPullRequestFiles(ctx context.Context, repo vcs.Repo, pull int) ([]string, error) {
	opt := forgejo.ListPullRequestFilesOptions{
		ListOptions: forgejo.ListOptions{
			Page:     0,
			PageSize: 50,
		},
	}
	resp := &forgejo.Response{NextPage: 1, LastPage: 1}
	var rv []string
	for resp.LastPage > opt.Page {
		opt.Page = resp.NextPage
		var files []*forgejo.ChangedFile
		var err error
		files, _, err = c.client.ListPullRequestFiles(repo.Owner(), repo.Name(), int64(pull), opt)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			rv = append(rv, file.Filename)
		}
	}
	return rv, nil
}

// GetCommit retrieves commit from the repo with the given git ref
func (c *Client) GetCommit(ctx context.Context, repo vcs.Repo, refname string) (vcs.Commit, error) {
	rv := vcs.Commit{}
	refs, _, err := c.client.GetRepoRefs(repo.Owner(), repo.Name(), refname)
	if err != nil {
		return rv, err
	}
	if len(refs) == 0 {
		return rv, errors.New("ref not found")
	}
	// the commit may be both a branch and a tag; just pick one (we only need the SHA, which should be the same for all refs)
	ref := refs[0]
	if ref.Object == nil {
		return rv, errors.New("ref has no commit")
	}
	commit, _, err := c.client.GetSingleCommit(repo.Owner(), repo.Name(), ref.Object.SHA)
	if err != nil {
		return rv, fmt.Errorf("forgejo.GetSingleCommit failed: %v", err)
	}
	if commit.Author != nil {
		rv.Author.Username = commit.Author.UserName
		rv.Author.AvatarURL = commit.Author.AvatarURL
		rv.Author.ProfileURL = commit.Author.HTMLURL
	}
	rv.SHA = commit.SHA
	rv.URL = commit.HTMLURL

	return rv, nil
}
