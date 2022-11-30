package otf

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithubEventHandler(t *testing.T) {
	handler := &GithubEventHandler{}

	t.Run("push event", func(t *testing.T) {
		r := newTestGithubPushEvent(t, "refs/heads/master")
		w := httptest.NewRecorder()
		got := handler.HandleEvent(w, r, &Webhook{})

		assert.Equal(t, 202, w.Code)

		want := VCSEvent{
			Branch: "master",
		}
		assert.Equal(t, want, got)
	})

	t.Run("pr event", func(t *testing.T) {
		r := newTestGithubPullRequestEvent(t, "pr-1")
		w := httptest.NewRecorder()
		got := handler.HandleEvent(w, r, &Webhook{})

		assert.Equal(t, 202, w.Code)

		want := VCSEvent{
			Branch:        "pr-1",
			IsPullRequest: true,
		}
		assert.Equal(t, want, got)
	})
}

func newTestGithubPushEvent(t *testing.T, ref string) *http.Request {
	push, err := json.Marshal(&github.PushEvent{
		Ref: String(ref),
	})
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/", bytes.NewReader(push))
	r.Header.Add("Content-type", "application/json")
	r.Header.Add(github.EventTypeHeader, "push")
	return r
}

func newTestGithubPullRequestEvent(t *testing.T, ref string) *http.Request {
	pr, err := json.Marshal(&github.PullRequestEvent{
		PullRequest: &github.PullRequest{
			Head: &github.PullRequestBranch{
				Ref: String(ref),
			},
		},
	})
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/", bytes.NewReader(pr))
	r.Header.Add("Content-type", "application/json")
	r.Header.Add(github.EventTypeHeader, "pull_request")
	return r
}
