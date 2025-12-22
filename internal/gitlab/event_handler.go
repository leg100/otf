package gitlab

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func HandleEvent(r *http.Request, secret string) (*vcs.EventPayload, error) {
	if token := r.Header.Get("X-Gitlab-Token"); token != secret {
		return nil, errors.New("token validation failed")
	}
	var origin *url.URL
	if instance := r.Header.Get("X-Gitlab-Instance"); instance == "" {
		return nil, errors.New("missing X-Gitlab-Instance header")
	} else {
		u, err := url.Parse(instance)
		if err != nil {
			return nil, fmt.Errorf("parsing X-Gitlab-Instance URL: %w", err)
		}
		origin = u
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		return nil, errors.New("error reading request body")
	}
	rawEvent, err := gitlab.ParseWebhook(gitlab.HookEventType(r), payload)
	if err != nil {
		return nil, fmt.Errorf("parsing webhook: %w", err)
	}

	// convert gitlab event to an OTF event
	var to vcs.EventPayload
	switch event := rawEvent.(type) {
	case *gitlab.PushEvent:
		to.Type = vcs.EventTypePush

		repo, err := vcs.NewRepoFromString(event.Project.PathWithNamespace)
		if err != nil {
			return nil, err
		}
		to.Repo = repo

		branch, found := strings.CutPrefix(event.Ref, "refs/heads/")
		if !found {
			return nil, fmt.Errorf("malformed ref: %s", event.Ref)
		}
		to.Action = vcs.ActionCreated
		to.Branch = branch
		to.CommitSHA = event.After
		to.CommitURL = event.Project.WebURL + "/commit/" + to.CommitSHA
		to.DefaultBranch = event.Project.DefaultBranch
		to.SenderUsername = event.UserUsername
		to.SenderAvatarURL = event.UserAvatar
		to.SenderHTMLURL = userURL(origin, event.UserUsername)
		// populate event with list of changed file paths
		for _, c := range event.Commits {
			to.Paths = append(to.Paths, c.Added...)
			to.Paths = append(to.Paths, c.Modified...)
			to.Paths = append(to.Paths, c.Removed...)
		}
		// remove duplicate file paths
		slices.Sort(to.Paths)
		to.Paths = slices.Compact(to.Paths)
	case *gitlab.MergeEvent:
		to.Type = vcs.EventTypePull

		repo, err := vcs.NewRepoFromString(event.Project.PathWithNamespace)
		if err != nil {
			return nil, err
		}
		to.Repo = repo

		to.Branch = event.ObjectAttributes.SourceBranch
		switch event.ObjectAttributes.Action {
		case "open":
			to.Action = vcs.ActionCreated
		case "update":
			to.Action = vcs.ActionUpdated
		default:
			return nil, vcs.NewErrIgnoreEvent("unsupported action: %s", event.ObjectAttributes.Action)
		}
		to.CommitSHA = event.ObjectAttributes.LastCommit.ID
		to.CommitURL = event.ObjectAttributes.LastCommit.URL
		to.PullRequestNumber = int(event.ObjectAttributes.IID)
		to.PullRequestURL = event.ObjectAttributes.URL
		to.PullRequestTitle = event.ObjectAttributes.Title
		to.DefaultBranch = event.Project.DefaultBranch
		to.SenderUsername = event.User.Username
		to.SenderAvatarURL = event.User.AvatarURL
		to.SenderHTMLURL = userURL(origin, event.User.Username)
	case *gitlab.TagEvent:
		to.Type = vcs.EventTypeTag

		repo, err := vcs.NewRepoFromString(event.Project.PathWithNamespace)
		if err != nil {
			return nil, err
		}
		to.Repo = repo

		tag, err := internal.ParseTagRef(event.Ref)
		if err != nil {
			return nil, err
		}
		to.Tag = tag
		// use presence of commits to determine whether the tag has been created
		// or deleted
		if len(event.Commits) > 0 {
			to.Action = vcs.ActionCreated
			to.CommitURL = event.Commits[0].URL
			to.CommitSHA = event.Commits[0].ID
		} else {
			to.Action = vcs.ActionDeleted
		}
		to.DefaultBranch = event.Project.DefaultBranch
		to.SenderUsername = event.UserUsername
		to.SenderAvatarURL = event.UserAvatar
		to.SenderHTMLURL = userURL(origin, event.UserUsername)
	default:
		return nil, vcs.NewErrIgnoreEvent("unsupported event type: %T", rawEvent)
	}
	if err := to.Validate(); err != nil {
		return nil, fmt.Errorf("failed building OTF event: %w", err)
	}
	return &to, nil
}

func userURL(origin *url.URL, username string) string {
	u := &url.URL{Scheme: origin.Scheme, Host: origin.Host, Path: username}
	return u.String()
}
