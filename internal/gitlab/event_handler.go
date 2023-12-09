package gitlab

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/leg100/otf/internal/vcs"
	"github.com/xanzy/go-gitlab"
)

func HandleEvent(w http.ResponseWriter, r *http.Request, secret string) *vcs.EventPayload {
	event, err := handleWithError(r, secret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return nil
	}
	w.WriteHeader(http.StatusNoContent)
	return event
}

func handleWithError(r *http.Request, secret string) (*vcs.EventPayload, error) {
	if token := r.Header.Get("X-Gitlab-Token"); token != secret {
		return nil, errors.New("token validation failed")
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
	to := vcs.EventPayload{
		VCSKind: vcs.GitlabKind,
	}

	switch event := rawEvent.(type) {
	case *gitlab.PushEvent:
		refParts := strings.Split(event.Ref, "/")
		if len(refParts) != 3 {
			return nil, fmt.Errorf("malformed ref: %s", event.Ref)
		}
		to.Type = vcs.EventTypePush
		to.Action = vcs.ActionCreated
		to.Branch = refParts[2]
		to.CommitSHA = event.After
		to.CommitURL = event.Project.WebURL + "/commit/" + to.CommitSHA
		to.DefaultBranch = event.Project.DefaultBranch
		to.RepoPath = event.Project.PathWithNamespace
		to.SenderUsername = event.UserUsername
		to.SenderAvatarURL = event.UserAvatar
		to.SenderHTMLURL = event.UserAvatar
		// populate event with list of changed file paths
		for _, c := range event.Commits {
			to.Paths = append(to.Paths, c.Added...)
			to.Paths = append(to.Paths, c.Modified...)
			to.Paths = append(to.Paths, c.Removed...)
		}
		// remove duplicate file paths
		slices.Sort(to.Paths)
		to.Paths = slices.Compact(to.Paths)
	case *gitlab.TagEvent:
		refParts := strings.Split(event.Ref, "/")
		if len(refParts) != 3 {
			return nil, fmt.Errorf("malformed ref: %s", event.Ref)
		}
		to.Tag = refParts[2]
		// Action:     action,
		to.CommitSHA = event.After
		to.DefaultBranch = event.Project.DefaultBranch
	default:
		return nil, nil
	}

	if err := to.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}
	return &to, nil
}
