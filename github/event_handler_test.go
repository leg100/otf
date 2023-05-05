package github

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventHandler(t *testing.T) {
	t.Run("push event", func(t *testing.T) {
		r := newTestPushEvent(t, "refs/heads/master")
		w := httptest.NewRecorder()
		got := HandleEvent(w, r, cloud.HandleEventOptions{})

		assert.Equal(t, 202, w.Code)

		want := cloud.VCSPushEvent{
			Branch: "master",
		}
		assert.Equal(t, want, got)
	})

	t.Run("pr open event", func(t *testing.T) {
		r := newTestPullRequestEvent(t, "pr-1", "opened")
		w := httptest.NewRecorder()
		got := HandleEvent(w, r, cloud.HandleEventOptions{})

		assert.Equal(t, 202, w.Code)

		want := cloud.VCSPullEvent{
			Action: cloud.VCSPullEventOpened,
			Branch: "pr-1",
		}
		assert.Equal(t, want, got)
	})

	t.Run("pr update event", func(t *testing.T) {
		r := newTestPullRequestEvent(t, "pr-1", "synchronize")
		w := httptest.NewRecorder()
		got := HandleEvent(w, r, cloud.HandleEventOptions{})

		assert.Equal(t, 202, w.Code)

		want := cloud.VCSPullEvent{
			Action: cloud.VCSPullEventUpdated,
			Branch: "pr-1",
		}
		assert.Equal(t, want, got)
	})
}

func newTestPushEvent(t *testing.T, ref string) *http.Request {
	push, err := json.Marshal(&github.PushEvent{
		Ref: otf.String(ref),
	})
	require.NoError(t, err)

	r := httptest.NewRequest("POST", "/", bytes.NewReader(push))
	r.Header.Add("Content-type", "application/json")
	r.Header.Add(github.EventTypeHeader, "push")
	return r
}

func newTestPullRequestEvent(t *testing.T, ref, action string) *http.Request {
	pr, err := json.Marshal(&github.PullRequestEvent{
		Action: &action,
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
