package forgejo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

// HACK: cached client object for handling pull_request webhook events
var forgejoClient *Client

type Client struct {
	client *forgejo.Client
}

func NewTokenClient(opts vcs.NewTokenClientOptions) (vcs.Client, error) {
	log.Printf("creating forgejo client object")
	url := fmt.Sprintf("https://%s", opts.Hostname)
	rv, err := forgejo.NewClient(url, forgejo.SetToken(opts.Token))
	if err != nil {
		return nil, err
	}
	forgejoClient = &Client{client: rv}
	return forgejoClient, nil
}

func (c *Client) ListRepositories(ctx context.Context, opts vcs.ListRepositoriesOptions) ([]string, error) {
	log.Printf("forgejo.ListRepositories called")

	// get the current user's id
	user, _, err := c.client.GetMyUserInfo()
	if err != nil {
		return nil, err
	}

	opt := forgejo.SearchRepoOptions{
		// search only for repos that the user with the given id owns or contributes to
		OwnerID: user.ID,
		ListOptions: forgejo.ListOptions{
			Page:     0,
			PageSize: 50,
		},
	}
	resp := &forgejo.Response{NextPage: 1, LastPage: 1}
	var rv []string
	for resp.LastPage > opt.Page {
		opt.Page = resp.NextPage
		var repolist []*forgejo.Repository
		var err error
		repolist, resp, err = c.client.SearchRepos(opt)
		if err != nil {
			return nil, err
		}
		for _, repo := range repolist {
			if repo.Permissions.Pull {
				rv = append(rv, repo.FullName)
			}
		}
	}
	return rv, nil
}

func (c *Client) GetRepository(ctx context.Context, identifier string) (vcs.Repository, error) {
	log.Printf("forgejo.GetRepository(%s) called", identifier)
	parts := strings.Split(identifier, "/")
	if len(parts) != 2 {
		return vcs.Repository{}, fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", identifier)
	}
	owner, reponame := parts[0], parts[1]
	repo, _, err := c.client.GetRepo(owner, reponame)
	if err != nil {
		return vcs.Repository{}, err
	}

	return vcs.Repository{
		Path:          repo.FullName,
		DefaultBranch: repo.DefaultBranch,
	}, nil
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
	log.Printf("forgejo.CreateWebhook() called")
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", opts.Repo)
	}
	owner, reponame := parts[0], parts[1]
	events, err := vcsOptToStringSlice(opts)
	if err != nil {
		return "", err
	}
	opt := forgejo.CreateHookOption{
		Type: forgejo.HookTypeForgejo,
		Config: map[string]string{
			"content_type": "json",
			"url":          opts.Endpoint,
			"secret":       opts.Secret, // used for for signatures
		},
		Events: events,
		Active: true,
	}
	wh, _, err := c.client.CreateRepoHook(owner, reponame, opt)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(wh.ID, 16), nil
}

func (c *Client) UpdateWebhook(ctx context.Context, id string, opts vcs.UpdateWebhookOptions) error {
	log.Printf("forgejo.UpdateWebhook() called")
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", opts.Repo)
	}
	owner, reponame := parts[0], parts[1]
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
	_, err = c.client.EditRepoHook(owner, reponame, idint, opt)
	return err
}
func (c *Client) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	log.Printf("forgejo.GetWebhook() called")
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return vcs.Webhook{}, fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", opts.Repo)
	}
	owner, reponame := parts[0], parts[1]
	idint, err := strconv.ParseInt(opts.ID, 16, 64)
	if err != nil {
		return vcs.Webhook{}, err
	}
	wh, _, err := c.client.GetRepoHook(owner, reponame, idint)
	if err != nil {
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
	log.Printf("forgejo.DeleteWebhook() called")
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", opts.Repo)
	}
	owner, reponame := parts[0], parts[1]
	idint, err := strconv.ParseInt(opts.ID, 16, 64)
	if err != nil {
		return err
	}
	_, err = c.client.DeleteRepoHook(owner, reponame, idint)
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
	log.Printf("forgejo.GetRepoTarball(%s) called", opts.Repo)
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", opts.Repo)
	}
	owner, reponame := parts[0], parts[1]
	ref := ""
	if opts.Ref != nil {
		ref = *opts.Ref
	}
	if ref == "" {
		// nil means default branch
		repo, _, err := c.client.GetRepo(owner, reponame)
		if err != nil {
			return nil, "", err
		}

		ref = repo.DefaultBranch
	}
	fqref, err := c.fullyQualifyRef(owner, reponame, ref)
	if err != nil {
		return nil, "", err
	}
	log.Printf("ref is %s, fqref is %s", ref, fqref)

	tarball, _, err := c.client.GetArchive(owner, reponame, ref, forgejo.TarGZArchive)
	if err != nil {
		return nil, "", fmt.Errorf("GetArchive(\"%s\", \"%s\", \"%s\", \"%s\") failed: %v", owner, reponame, ref, forgejo.TarGZArchive, err)
	}

	// Forgejo tarball contents are contained within a top-level directory
	// named after the repo. We want the tarball without this directory,
	// so we re-tar the contents without the top-level directory.
	untarpath, err := os.MkdirTemp("", fmt.Sprintf("forgejo-%s-%s-*", owner, reponame))
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
	log.Printf("forgejo.SetStatus(%s, %s, %s) called", opts.Repo, opts.Ref, opts.Status)
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", opts.Repo)
	}
	owner, reponame := parts[0], parts[1]
	opt := forgejo.CreateStatusOption{
		State:       vcsStateToForgejo(opts.Status),
		TargetURL:   opts.TargetURL,
		Description: opts.Description,
		Context:     "otf",
	}
	_, _, err := c.client.CreateStatus(owner, reponame, opts.Ref, opt)
	return err
}

func (c *Client) ListTags(ctx context.Context, opts vcs.ListTagsOptions) ([]string, error) {
	log.Printf("forgejo.ListTags(%s, %s) called", opts.Repo, opts.Prefix)
	parts := strings.Split(opts.Repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", opts.Repo)
	}
	owner, reponame := parts[0], parts[1]

	opt := forgejo.ListRepoTagsOptions{
		ListOptions: forgejo.ListOptions{
			Page:     0,
			PageSize: 50,
		},
	}
	resp := &forgejo.Response{NextPage: 1, LastPage: 1}
	var rv []string
	for resp.LastPage > opt.Page {
		opt.Page = resp.NextPage
		var tags []*forgejo.Tag
		var err error
		tags, resp, err = c.client.ListRepoTags(owner, reponame, opt)
		if err != nil {
			return nil, err
		}
		for _, tag := range tags {
			rv = append(rv, fmt.Sprintf("tags/%s", tag.Name))
		}
	}
	return rv, nil
}

// ListPullRequestFiles returns the paths of files that are modified in the pull request
func (c *Client) ListPullRequestFiles(ctx context.Context, repo string, pull int) ([]string, error) {
	log.Printf("forgejo.ListPullRequestFiles(%s, %d) called", repo, pull)
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", repo)
	}
	owner, reponame := parts[0], parts[1]
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
		files, _, err = c.client.ListPullRequestFiles(owner, reponame, int64(pull), opt)
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
func (c *Client) GetCommit(ctx context.Context, repo, refname string) (vcs.Commit, error) {
	log.Printf("forgejo.GetCommit(%s, %s) called", repo, refname)
	rv := vcs.Commit{}
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return rv, fmt.Errorf("identifier '%s' must be in the form 'owner/repo'", repo)
	}
	owner, reponame := parts[0], parts[1]
	refs, _, err := c.client.GetRepoRefs(owner, reponame, refname)
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
	commit, _, err := c.client.GetSingleCommit(owner, reponame, ref.Object.SHA)
	if err != nil {
		return rv, fmt.Errorf("forgejo.GetSingleCommit failed: %v", err)
	}
	if commit.RepoCommit == nil {
		return rv, errors.New("commit has no RepoCommit")
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
