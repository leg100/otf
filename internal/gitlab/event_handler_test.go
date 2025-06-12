package gitlab

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventHandler(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		body      string
		want      *vcs.EventPayload
	}{
		{
			"push",
			"Push Hook",
			"./testdata/push.json",
			&vcs.EventPayload{
				Type:          vcs.EventTypePush,
				RepoPath:      "mike/diaspora",
				Branch:        "master",
				DefaultBranch: "master",
				CommitSHA:     "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
				CommitURL:     "http://example.com/mike/diaspora/commit/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
				Action:        vcs.ActionCreated,
				Paths: []string{
					"CHANGELOG",
					"app/controller/application.rb",
				},
				SenderUsername:  "jsmith",
				SenderAvatarURL: "https://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=8://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=80",
				SenderHTMLURL:   "https://github.com/jsmith",
			},
		},
		{
			"open merge request",
			"Merge Request Hook",
			"./testdata/merge_opened.json",
			&vcs.EventPayload{
				Type:              vcs.EventTypePull,
				Action:            vcs.ActionCreated,
				RepoPath:          "leg100/otf-workspaces",
				Branch:            "pr-1",
				DefaultBranch:     "master",
				CommitSHA:         "eea3783a079cd610b748e406610e78c7ce2f34e6",
				CommitURL:         "https://gitlab.com/leg100/otf-workspaces/-/commit/eea3783a079cd610b748e406610e78c7ce2f34e6",
				PullRequestNumber: 1,
				PullRequestURL:    "https://gitlab.com/leg100/otf-workspaces/-/merge_requests/1",
				PullRequestTitle:  "Pr 1",
				SenderUsername:    "leg100",
				SenderAvatarURL:   "https://secure.gravatar.com/avatar/de3ca65d31c67b63a795b88c677bba5d?s=80&d=identicon",
				SenderHTMLURL:     "https://github.com/leg100",
			},
		},
		{
			"update merge request",
			"Merge Request Hook",
			"./testdata/merge_updated.json",
			&vcs.EventPayload{
				Type:              vcs.EventTypePull,
				Action:            vcs.ActionUpdated,
				RepoPath:          "leg100/otf-workspaces",
				Branch:            "pr-1",
				DefaultBranch:     "master",
				CommitSHA:         "30c78003043f3a5d8f34eda6332ad11376b1d41b",
				CommitURL:         "https://gitlab.com/leg100/otf-workspaces/-/commit/30c78003043f3a5d8f34eda6332ad11376b1d41b",
				PullRequestNumber: 1,
				PullRequestURL:    "https://gitlab.com/leg100/otf-workspaces/-/merge_requests/1",
				PullRequestTitle:  "Pr 1",
				SenderUsername:    "leg100",
				SenderAvatarURL:   "https://secure.gravatar.com/avatar/de3ca65d31c67b63a795b88c677bba5d?s=80&d=identicon",
				SenderHTMLURL:     "https://github.com/leg100",
			},
		},
		{
			"push tag",
			"Tag Push Hook",
			"./testdata/tag_created.json",
			&vcs.EventPayload{
				Type:            vcs.EventTypeTag,
				Action:          vcs.ActionCreated,
				RepoPath:        "leg100/otf-workspaces",
				Tag:             "v3",
				DefaultBranch:   "master",
				CommitSHA:       "eea3783a079cd610b748e406610e78c7ce2f34e6",
				CommitURL:       "https://gitlab.com/leg100/otf-workspaces/-/commit/eea3783a079cd610b748e406610e78c7ce2f34e6",
				SenderUsername:  "leg100",
				SenderAvatarURL: "https://secure.gravatar.com/avatar/de3ca65d31c67b63a795b88c677bba5d?s=80&d=identicon",
				SenderHTMLURL:   "https://github.com/leg100",
			},
		},
		{
			"push deleted tag",
			"Tag Push Hook",
			"./testdata/tag_deleted.json",
			&vcs.EventPayload{
				Type:            vcs.EventTypeTag,
				Action:          vcs.ActionDeleted,
				RepoPath:        "leg100/otf-workspaces",
				Tag:             "v3",
				DefaultBranch:   "master",
				SenderUsername:  "leg100",
				SenderAvatarURL: "https://secure.gravatar.com/avatar/de3ca65d31c67b63a795b88c677bba5d?s=80&d=identicon",
				SenderHTMLURL:   "https://github.com/leg100",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.body)
			require.NoError(t, err)
			defer f.Close()

			r := httptest.NewRequest("POST", "/", f)
			r.Header.Add("Content-type", "application/json")
			r.Header.Add("X-Gitlab-Event", tt.eventType)
			r.Header.Add("X-Gitlab-Instance", "https://github.com")
			w := httptest.NewRecorder()
			got, err := HandleEvent(r, "")
			require.NoError(t, err)
			assert.Equal(t, 200, w.Code, w.Body.String())
			assert.Equal(t, tt.want, got)
		})
	}
}
