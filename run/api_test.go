package run

import (
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestAPI_Watch(t *testing.T) {
	// input event channel
	in := make(chan otf.Event, 1)

	want := "{}"

	srv := &api{
		Logger:           logr.Discard(),
		jsonapiMarshaler: &fakeJSONAPIMarshaler{marshaled: []byte(want)},
		svc:              &fakeService{ch: in},
	}

	r := httptest.NewRequest("", "/", nil)
	w := httptest.NewRecorder()

	// send one event and then close
	in <- otf.Event{
		Payload: &Run{},
		Type:    otf.EventRunCreated,
	}
	close(in)

	done := make(chan struct{})
	go func() {
		srv.watch(w, r)
		assert.Equal(t, "data: e30=\nevent: run_created\n\n", w.Body.String())
		done <- struct{}{}
	}()
	<-done
}
