package api

import (
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DataDog/jsonapi"
	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
)

func TestAPI_Watch(t *testing.T) {
	// input event channel
	in := make(chan pubsub.Event, 1)

	srv := &api{
		Logger:     logr.Discard(),
		marshaler:  &fakeMarshaler{run: &types.Run{ID: "run-123"}},
		RunService: &fakeRunService{ch: in},
	}

	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	// send one event and then close
	in <- pubsub.Event{
		Payload: &run.Run{ID: "run-123"},
		Type:    pubsub.CreatedEvent,
	}
	close(in)

	done := make(chan struct{})
	go func() {
		srv.watchRun(w, r)
		// should receive sse event that looks like "<whitespace>data:
		// <data><newline>event: <event><newline><newline>
		got := w.Body.String()
		got = strings.TrimSpace(got)
		parts := strings.Split(got, "\n")
		if assert.Equal(t, 2, len(parts)) {
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
