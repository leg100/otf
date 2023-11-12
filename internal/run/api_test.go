package run

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/tfeapi/types"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestAPI_Watch(t *testing.T) {
	// input event channel
	in := make(chan pubsub.Event, 1)

	srv := &api{
		Logger:  logr.Discard(),
		Service: &fakeRunService{ch: in},
	}

	r := httptest.NewRequest("", "/", nil)
	r = r.WithContext(internal.AddSubjectToContext(r.Context(), &user.User{ID: "janitor"}))
	w := httptest.NewRecorder()

	// send one event and then close
	in <- pubsub.Event{
		// we need to provide all IDs otherwise the json-api marshaler complains
		// the primary field is empty.
		Payload: &Run{
			ID:                     "run-123",
			WorkspaceID:            "ws-123",
			ConfigurationVersionID: "cv-123",
			Plan:                   Phase{RunID: "run-123"},
			Apply:                  Phase{RunID: "run-123"},
		},
		Type: pubsub.CreatedEvent,
	}
	close(in)

	done := make(chan struct{})
	go func() {
		srv.watch(w, r)
		// should receive sse event that looks like "<whitespace>data:
		// <data><newline>event: <event><newline><newline>
		got := w.Body.String()
		got = strings.TrimSpace(got)
		parts := strings.Split(got, "\n")
		if assert.Equal(t, 2, len(parts), got, got) {
			assert.Equal(t, "event: created", parts[1])
			if assert.Regexp(t, `data: .*`, parts[0]) {
				data := strings.TrimPrefix(parts[0], "data: ")
				// base64 decode
				decoded, err := base64.StdEncoding.DecodeString(data)
				if assert.NoError(t, err) {
					// unmarshal into json:api struct
					var run types.Run
					err := jsonapi.Unmarshal(decoded, &run)
					assert.NoError(t, err)
				}
			}
		}

		done <- struct{}{}
	}()
	<-done
}

type (
	fakeRunService struct {
		ch chan pubsub.Event

		RunService
	}
)

func (f *fakeRunService) Watch(context.Context, WatchOptions) (<-chan pubsub.Event, error) {
	return f.ch, nil
}
