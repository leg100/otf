package github

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventHandler(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		body      string
		want      *vcs.Event
	}{
		{
			"push",
			"push",
			"./testdata/github_push.json",
			&vcs.Event{
				Cloud:           cloud.GithubKind,
				Type:            vcs.EventTypePush,
				Branch:          "master",
				DefaultBranch:   "master",
				CommitSHA:       "42d6fc7dac35cc7945231195e248af2f6256b522",
				CommitURL:       "https://github.com/leg100/tfc-workspaces/commit/42d6fc7dac35cc7945231195e248af2f6256b522",
				Action:          vcs.ActionCreated,
				Paths:           []string{"main.tf"},
				SenderUsername:  "leg100",
				SenderAvatarURL: "https://avatars.githubusercontent.com/u/75728?v=4",
				SenderHTMLURL:   "https://github.com/leg100",
			},
		},
		{
			"pull request opened",
			"pull_request",
			"./testdata/github_pull_opened.json",
			&vcs.Event{
				Cloud:             cloud.GithubKind,
				Type:              vcs.EventTypePull,
				Branch:            "pr-2",
				DefaultBranch:     "master",
				CommitSHA:         "c560613b228f5e189520fbab4078284ea8312bcb",
				CommitURL:         "https://github.com/leg100/otf-workspaces/commit/c560613b228f5e189520fbab4078284ea8312bcb",
				PullRequestNumber: 2,
				PullRequestURL:    "https://github.com/leg100/otf-workspaces/pull/2",
				PullRequestTitle:  "pr-2",
				Action:            vcs.ActionCreated,
				SenderUsername:    "leg100",
				SenderAvatarURL:   "https://avatars.githubusercontent.com/u/75728?v=4",
				SenderHTMLURL:     "https://github.com/leg100",
			},
		},
		{
			"pull request updated",
			"pull_request",
			"./testdata/github_pull_update.json",
			&vcs.Event{
				Cloud:             cloud.GithubKind,
				Type:              vcs.EventTypePull,
				Branch:            "pr-1",
				DefaultBranch:     "master",
				CommitSHA:         "067e2b4c6394b3dad3c0ec89ffc428ab60ae7e5d",
				CommitURL:         "https://github.com/leg100/otf-workspaces/commit/067e2b4c6394b3dad3c0ec89ffc428ab60ae7e5d",
				PullRequestNumber: 1,
				PullRequestURL:    "https://github.com/leg100/otf-workspaces/pull/1",
				PullRequestTitle:  "pr-1",
				Action:            vcs.ActionUpdated,
				SenderUsername:    "leg100",
				SenderAvatarURL:   "https://avatars.githubusercontent.com/u/75728?v=4",
				SenderHTMLURL:     "https://github.com/leg100",
			},
		},
		{
			"tag pushed",
			"push",
			"./testdata/github_push_tag.json",
			&vcs.Event{
				Cloud:           cloud.GithubKind,
				Type:            vcs.EventTypeTag,
				Tag:             "v1.0.0",
				DefaultBranch:   "master",
				CommitSHA:       "07101e82c4f525d5f697111f0690bdd0ff40a865",
				CommitURL:       "https://github.com/leg100/terraform-otf-test/commit/07101e82c4f525d5f697111f0690bdd0ff40a865",
				Action:          vcs.ActionCreated,
				SenderUsername:  "leg100",
				SenderAvatarURL: "https://avatars.githubusercontent.com/u/75728?v=4",
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
			r.Header.Add(github.EventTypeHeader, tt.eventType)
			w := httptest.NewRecorder()
			got := HandleEvent(w, r, "")
			assert.Equal(t, 202, w.Code)
			assert.Equal(t, tt.want, got)
		})
	}
}
