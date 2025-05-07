package components

import (
	"context"
	"fmt"
	"net/http"

	"time"

	"github.com/a-h/templ"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
)

// WebsocketGetHandler handles dynamically retrieving and updating a resource
// via a websocket.
type WebsocketGetHandler[Resource any, ResourceEvent Identifiable] struct {
	Logger    logr.Logger
	Client    websocketGetHandlerClient[Resource, ResourceEvent]
	Component func(Resource) templ.Component
}

type Identifiable interface {
	GetID() resource.TfeID
}

type websocketGetHandlerClient[Resource any, ResourceEvent Identifiable] interface {
	Watch(ctx context.Context) (<-chan pubsub.Event[ResourceEvent], func())
	Get(ctx context.Context, id resource.TfeID) (Resource, error)
	LookupEvent(ctx context.Context, event ResourceEvent) (Resource, error)
}

func (h *WebsocketGetHandler[Resource, ResourceEvent]) Handler(w http.ResponseWriter, r *http.Request, initialID *resource.TfeID, condition func(ResourceEvent) bool) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Logger.Error(err, "upgrading websocket connection")
		return
	}
	defer conn.Close()

	sub, unsub := h.Client.Watch(r.Context())
	defer unsub()

	send := func(res Resource) error {
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()

		if err := html.RenderSnippet(h.Component(res), w, r); err != nil {
			return fmt.Errorf("rendering html: %w", err)
		}
		return nil
	}

	// Send an initial response to client on startup if an initial ID was
	// specified.
	if initialID != nil {
		res, err := h.Client.Get(r.Context(), *initialID)
		if err != nil {
			h.Logger.Error(err, "fetching resource")
			return
		}
		if err := send(res); err != nil {
			h.Logger.Error(err, "handling websocket connection")
			return
		}
	}

	// 1) Watch server events, check whether event matches the condition. If so,
	// send resource to the client.
	// To avoid overwhelming the client, do not send resource more than once a
	// second.
	for {
		var resource Resource
		// Block on receiving an event.
		select {
		case res, ok := <-sub:
			if !ok {
				return
			}
			if res.Type == pubsub.DeletedEvent {
				// TODO: resource has been deleted: user should be alerted and
				// client should not reconnect.
				return
			}
			if !condition(res.Payload) {
				continue
			}
			resource = res.Payload
		case <-r.Context().Done():
			return
		}
		// Then consume any remaining events. If any of the remaining events
		// match the condition then update the resource to be sent to the
		// client, to ensure the lastest version is sent.
		for {
			select {
			case res, ok := <-sub:
				if !ok {
					return
				}
				if res.Type == pubsub.DeletedEvent {
					// TODO: resource has been deleted: user should be alerted and
					// client should not reconnect.
					return
				}
				if condition(res.Payload) {
					resource = res.Payload
				}
			default:
				goto done
			}
		}
	done:
		// Send resource to client
		if err := send(resource); err != nil {
			h.Logger.Error(err, "sending websocket message")
			return
		}
		// Wait a second before sending anything more to client.
		select {
		case <-time.After(time.Second):
		case <-r.Context().Done():
			return
		}
	}
}
