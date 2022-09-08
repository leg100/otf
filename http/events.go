package http

import (
	"bytes"
	"net/http"

	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
	"github.com/r3labs/sse/v2"
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
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	go func() {
		for {
			select {
			case <-r.Context().Done():
				// client closes connection
				s.eventsServer.RemoveStream(streamID)
				return
			case event, ok := <-events:
				if !ok {
					// server closes connection
					s.eventsServer.RemoveStream(streamID)
					return
				}

				marshalable, ok := event.Payload.(dto.Assembler)
				if !ok {
					// skip events that cannot be marshaled into a JSON-API
					// object
					continue
				}

				buf := bytes.Buffer{}
				if err = jsonapi.MarshalPayloadWithoutIncluded(&buf, marshalable.ToJSONAPI(r)); err != nil {
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
