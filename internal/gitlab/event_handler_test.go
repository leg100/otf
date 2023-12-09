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
				VCSKind:       vcs.GitlabKind,
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
				SenderHTMLURL:   "https://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=8://s.gravatar.com/avatar/d4c74594d841139328695756648b6bd6?s=80",
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
			w := httptest.NewRecorder()
			got := HandleEvent(w, r, "")
			assert.Equal(t, 204, w.Code, w.Body.String())
			assert.Equal(t, tt.want, got)
		})
	}
}
