package watch

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/r3labs/sse/v2"
)

type api struct {
	Application
	eventsServer *sse.Server
}

func (a *api) AddHandlers(r *mux.Router) {
	r.HandleFunc(otf.DefaultWatchPath, a.watch).Methods("GET")
}

// watch subscribes to a stream of otf events using the server-side-events
// protocol
func (a *api) watch(w http.ResponseWriter, r *http.Request) {
	// r3lab's sse server expects a query parameter with the stream ID
	// but we don't want to bother the client with having to do that so we
	// handle it here
	streamID := otf.GenerateRandomString(6)
	q := r.URL.Query()
	q.Add("stream", streamID)
	r.URL.RawQuery = q.Encode()

	a.eventsServer.CreateStream(streamID)

	events, err := a.Watch(r.Context(), otf.WatchOptions{})
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

				// Watch currently only streams run events
				run, ok := event.Payload.(otf.Run)
				if !ok {
					continue
				}

				buf := bytes.Buffer{}
				if err = jsonapi.MarshalPayloadWithoutIncluded(&buf, (&Run{run, r, a}).ToJSONAPI()); err != nil {
					a.Error(err, "marshalling event", "event", event.Type)
					continue
				}

				a.eventsServer.Publish(streamID, &sse.Event{
					Data:  buf.Bytes(),
					Event: []byte(event.Type),
				})
			}
		}
	}()
	a.eventsServer.ServeHTTP(w, r)
}
