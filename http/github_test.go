package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithubEventHandler(t *testing.T) {
	events := make(chan otf.VCSEvent, 1)
	handler := &GithubEventHandler{
		events: events,
		Logger: logr.Discard(),
	}

	t.Run("push event", func(t *testing.T) {
		r := newTestGithubPushEvent(t, "refs/heads/master")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		assert.Equal(t, 202, w.Code)

		want := otf.VCSEvent{
			Branch: "master",
		}
		assert.Equal(t, want, <-events)
	})

	t.Run("pr event", func(t *testing.T) {
		r := newTestGithubPullRequestEvent(t, "pr-1")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		assert.Equal(t, 202, w.Code)

		want := otf.VCSEvent{
			Branch:        "pr-1",
			IsPullRequest: true,
		}
		assert.Equal(t, want, <-events)
	})
}

func newTestGithubPushEvent(t *testing.T, ref string) *http.Request {
	push, err := json.Marshal(&github.PushEvent{
		Ref: otf.String(ref),
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
				Ref: otf.String(ref),
			},
		},
	})
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/", bytes.NewReader(pr))
	r.Header.Add("Content-type", "application/json")
	r.Header.Add(github.EventTypeHeader, "pull_request")
	return r
}
