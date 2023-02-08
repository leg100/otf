package watch

import (
	"bytes"
	"net/http"

	"github.com/gin-contrib/sse"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// watch subscribes to a stream of otf events using the server-side-events
// protocol
func (s *Server) watch(w http.ResponseWriter, r *http.Request) {
	// r3lab's sse server expects a query parameter with the stream ID
	// but we don't want to bother the client with having to do that so we
	// handle it here
	streamID := otf.GenerateRandomString(6)
	q := r.URL.Query()
	q.Add("stream", streamID)
	r.URL.RawQuery = q.Encode()

	s.eventsServer.CreateStream(streamID)

	events, err := s.Watch(r.Context(), otf.WatchOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}
	go func() {
		for {
			select {
			case <-r.Context().Done():
				// client closed connection
				s.eventsServer.RemoveStream(streamID)
				return
			case event, ok := <-events:
				if !ok {
					// server closes connection
					s.eventsServer.RemoveStream(streamID)
					return
				}

				// Watch currently only streams run events
				run, ok := event.Payload.(*otf.Run)
				if !ok {
					continue
				}

				buf := bytes.Buffer{}
				if err = jsonapi.MarshalPayloadWithoutIncluded(&buf, (&Run{run, r, s}).ToJSONAPI()); err != nil {
					s.Error(err, "marshalling event", "event", event.Type)
					continue
				}

				s.eventsServer.Publish(streamID, &sse.Event{
					Data:  buf.Bytes(),
					Event: []byte(event.Type),
				})
			}
		}
	}()
	s.eventsServer.ServeHTTP(w, r)
}
