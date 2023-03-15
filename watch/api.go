package watch

import (
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/run"
	"github.com/r3labs/sse/v2"
)

// eventsServer is a server capable of streaming SSE events
type eventsServer interface {
	CreateStream(string) *sse.Stream
	RemoveStream(string)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Publish(string, *sse.Event)
}

type api struct {
	logr.Logger

	svc          Service
	eventsServer eventsServer
}

func (a *api) addHandlers(r *mux.Router) {
	r.HandleFunc(otf.DefaultWatchPath, a.watch).Methods("GET")
}

// Watch handler responds with a stream of events, using the json:api encoding.
//
// NOTE: Only run events are currently supported.
func (a *api) watch(w http.ResponseWriter, r *http.Request) {
	// r3lab's sse server expects a query parameter with the stream ID
	// but we don't want to bother the client with having to do that so we
	// handle it here
	streamID := otf.GenerateRandomString(6)
	q := r.URL.Query()
	q.Add("stream", streamID)
	r.URL.RawQuery = q.Encode()

	a.eventsServer.CreateStream(streamID)

	// TODO: populate watch options
	events, err := a.svc.Watch(r.Context(), otf.WatchOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	go func() {
		for {
			select {
			case <-r.Context().Done():
				// client closed connection
				a.eventsServer.RemoveStream(streamID)
				return
			case event, ok := <-events:
				if !ok {
					// server closes connection
					a.eventsServer.RemoveStream(streamID)
					return
				}

				// Only run events are supported
				run, ok := event.Payload.(*run.Run)
				if !ok {
					continue
				}

				data, err := json.Marshal(run)
				if err != nil {
					a.Error(err, "marshalling event", "event", event.Type)
					continue
				}

				a.eventsServer.Publish(streamID, &sse.Event{
					Data:  data,
					Event: []byte(event.Type),
				})
			}
		}
	}()
	a.eventsServer.ServeHTTP(w, r)
}
