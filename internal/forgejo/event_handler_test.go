package forgejo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func genSigForFile(fn, secret string) string {
	payload, err := os.ReadFile(fn)
	if err != nil {
		return err.Error()
	}
	hash := hmac.New(sha256.New, []byte(secret))
	if _, err := hash.Write(payload); err != nil {
		return err.Error()
	}
	raw := hash.Sum(nil)
	rv := hex.EncodeToString(raw)
	return rv
}

func TestEventHandler(t *testing.T) {
	secret := "abc123"
	tests := []struct {
		name      string
		eventType string
		body      string
		sig       string
		want      *vcs.EventPayload
		ignore    bool
	}{
		{
			name:      "admin pushed the 'test delivery' button on the webhook page",
			eventType: "push",
			body:      "./testdata/test_delivery.json",
			sig:       genSigForFile("./testdata/test_delivery.json", secret),
			want: &vcs.EventPayload{
				Type:            vcs.EventTypePush,
				Repo:            vcs.Repo{Owner: "tf", Name: "thing"},
				Branch:          "main",
				DefaultBranch:   "main",
				CommitSHA:       "671b42120e88955bc8d3f76a3fc18f63a8ecd90e",
				CommitURL:       "https://forgejo.example.com/tf/thing/commit/671b42120e88955bc8d3f76a3fc18f63a8ecd90e",
				Action:          vcs.ActionCreated,
				Paths:           []string(nil),
				SenderUsername:  "mark",
				SenderAvatarURL: "https://forgejo.example.com/avatars/10d9c0d0a1711ade4157e15e3ab93b0c5e2d64f4544733f4da5c4eacf240d82d",
				SenderHTMLURL:   "https://forgejo.example.com/mark",
			},
			ignore: false,
		},
		{
			name:      "tag added",
			eventType: "push",
			body:      "./testdata/add_tag.json",
			sig:       genSigForFile("./testdata/add_tag.json", secret),
			want: &vcs.EventPayload{
				Type:            vcs.EventTypeTag,
				Repo:            vcs.Repo{Owner: "tf", Name: "thing"},
				Tag:             "test",
				DefaultBranch:   "main",
				CommitSHA:       "671b42120e88955bc8d3f76a3fc18f63a8ecd90e",
				CommitURL:       "",
				Action:          vcs.ActionCreated,
				Paths:           []string(nil),
				SenderUsername:  "mark",
				SenderAvatarURL: "https://forgejo.example.com/avatars/10d9c0d0a1711ade4157e15e3ab93b0c5e2d64f4544733f4da5c4eacf240d82d",
				SenderHTMLURL:   "https://forgejo.example.com/mark",
			},
			ignore: false,
		},
		{
			name:      "tag deleted",
			eventType: "push",
			body:      "./testdata/delete_tag.json",
			sig:       genSigForFile("./testdata/delete_tag.json", secret),
			want: &vcs.EventPayload{
				Type:            vcs.EventTypeTag,
				Repo:            vcs.Repo{Owner: "tf", Name: "thing"},
				Tag:             "test",
				DefaultBranch:   "main",
				CommitSHA:       "671b42120e88955bc8d3f76a3fc18f63a8ecd90e",
				CommitURL:       "",
				Action:          vcs.ActionDeleted,
				Paths:           []string(nil),
				SenderUsername:  "mark",
				SenderAvatarURL: "https://forgejo.example.com/avatars/10d9c0d0a1711ade4157e15e3ab93b0c5e2d64f4544733f4da5c4eacf240d82d",
				SenderHTMLURL:   "https://forgejo.example.com/mark",
			},
			ignore: false,
		},
		{
			name:      "PR opened",
			eventType: "pull_request",
			body:      "./testdata/pr_opened.json",
			sig:       genSigForFile("./testdata/pr_opened.json", secret),
			want: &vcs.EventPayload{
				Type:              vcs.EventTypePull,
				Repo:              vcs.Repo{Owner: "tf", Name: "thing"},
				PullRequestTitle:  "test",
				PullRequestNumber: 64,
				PullRequestURL:    "https://forgejo.example.com/tf/thing/pulls/64",
				Branch:            "test",
				DefaultBranch:     "main",
				CommitSHA:         "69f477ffc7923880e2e945d29f5d4d804cffc584",
				CommitURL:         "",
				Action:            vcs.ActionCreated,
				Paths:             []string(nil),
				SenderUsername:    "mark",
				SenderAvatarURL:   "https://forgejo.example.com/avatars/10d9c0d0a1711ade4157e15e3ab93b0c5e2d64f4544733f4da5c4eacf240d82d",
				SenderHTMLURL:     "https://forgejo.example.com/mark",
			},
			ignore: false,
		},
		{
			name:      "PR comment",
			eventType: "issue_comment",
			body:      "./testdata/pr_comment.json",
			sig:       genSigForFile("./testdata/pr_comment.json", secret),
			ignore:    true,
		},
		{
			name:      "PR closed",
			eventType: "pull_request",
			body:      "./testdata/pr_closed.json",
			sig:       genSigForFile("./testdata/pr_closed.json", secret),
			want: &vcs.EventPayload{
				Type:              vcs.EventTypePull,
				Repo:              vcs.Repo{Owner: "tf", Name: "thing"},
				PullRequestTitle:  "test",
				PullRequestNumber: 64,
				PullRequestURL:    "https://forgejo.example.com/tf/thing/pulls/64",
				Branch:            "test",
				DefaultBranch:     "main",
				CommitSHA:         "69f477ffc7923880e2e945d29f5d4d804cffc584",
				CommitURL:         "",
				Action:            vcs.ActionDeleted,
				Paths:             []string(nil),
				SenderUsername:    "mark",
				SenderAvatarURL:   "https://forgejo.example.com/avatars/10d9c0d0a1711ade4157e15e3ab93b0c5e2d64f4544733f4da5c4eacf240d82d",
				SenderHTMLURL:     "https://forgejo.example.com/mark",
			},
			ignore: false,
		},
		{
			name:      "PR merged",
			eventType: "pull_request",
			body:      "./testdata/pr_merged.json",
			sig:       genSigForFile("./testdata/pr_merged.json", secret),
			want: &vcs.EventPayload{
				Type:              vcs.EventTypePull,
				Repo:              vcs.Repo{Owner: "tf", Name: "thing"},
				PullRequestTitle:  "test",
				PullRequestNumber: 64,
				PullRequestURL:    "https://forgejo.example.com/tf/thing/pulls/64",
				Branch:            "test",
				DefaultBranch:     "main",
				CommitSHA:         "69f477ffc7923880e2e945d29f5d4d804cffc584",
				CommitURL:         "",
				Action:            vcs.ActionMerged,
				Paths:             []string(nil),
				SenderUsername:    "mark",
				SenderAvatarURL:   "https://forgejo.example.com/avatars/10d9c0d0a1711ade4157e15e3ab93b0c5e2d64f4544733f4da5c4eacf240d82d",
				SenderHTMLURL:     "https://forgejo.example.com/mark",
			},
			ignore: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.body)
			require.NoError(t, err)
			defer func() {
				_ = f.Close()
			}()

			r := httptest.NewRequest("POST", "/", f)
			r.Header.Add("Content-type", "application/json")
			r.Header.Add("X-Forgejo-Event", tt.eventType)
			r.Header.Add("X-Gitea-Event", tt.eventType)
			r.Header.Add("X-Forgejo-Signature", tt.sig)
			r.Header.Add("X-Gitea-Signature", tt.sig)
			w := httptest.NewRecorder()
			got, err := HandleEvent(r, secret)
			t.Logf("error is %v", err)
			if tt.ignore {
				var ignore vcs.ErrIgnoreEvent
				assert.True(t, errors.As(err, &ignore))
			} else {
				require.NoError(t, err)
				assert.Equal(t, 200, w.Code, w.Body.String())
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
