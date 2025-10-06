package components

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"golang.org/x/sync/errgroup"
)

// WebsocketListHandler handles dynamically updating lists of resources via
// a websocket.
type WebsocketListHandler[Resource, ResourceEvent, Options any] struct {
	logr.Logger
	Client    websocketListHandlerClient[Resource, ResourceEvent, Options]
	Populator TablePopulator[Resource]
	ID        string
}

type websocketListHandlerClient[Resource, ResourceEvent, Options any] interface {
	Watch(ctx context.Context) (<-chan pubsub.Event[ResourceEvent], func())
	List(ctx context.Context, opts Options) (*resource.Page[Resource], error)
}

func (h *WebsocketListHandler[Resource, ResourceEvent, Options]) Handler(w http.ResponseWriter, r *http.Request) {
	var opts Options
	if err := decode.All(&opts, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Error(err, "upgrading websocket connection")
		return
	}
	defer conn.Close()

	sub, unsub := h.Client.Watch(r.Context())
	defer unsub()

	// Mutex serializes go routine access to the list options and to the
	// websocket writer.
	var mu sync.Mutex

	sendList := func() error {
		mu.Lock()
		defer mu.Unlock()

		page, err := h.Client.List(r.Context(), opts)
		if err != nil {
			return fmt.Errorf("fetching list of resources: %w", err)
		}

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()

		comp := Table(h.Populator, page, h.ID)
		if err := html.RenderSnippet(comp, w, r); err != nil {
			return fmt.Errorf("rendering html: %w", err)
		}
		return nil
	}

	// Send an initial list to client on startup.
	if err := sendList(); err != nil {
		h.Error(err, "handling websocket connection")
		return
	}

	// Two go-routines:
	// 1) Watch server events and upon receiving an event send a new list to the
	// client. This is necesary because the event can be a notification that a
	// resource has been created, updated or deleted, any of which alters the
	// list of resources on the client.
	// 2) Receive messages from the client altering the list of resources to
	// retrieve (e.g. next page, filtering by status, etc), and send new list to
	// the client.
	g, ctx := errgroup.WithContext(r.Context())
	g.Go(func() error {
		// To avoid overwhelming the client, do not send a list more than once a
		// second.
		for {
			// Block on receiving an event.
			select {
			case _, ok := <-sub:
				if !ok {
					return nil
				}
			case <-ctx.Done():
				return nil
			}
			// Then consume any remaining events.
			for {
				select {
				case _, ok := <-sub:
					if !ok {
						goto done
					}
				default:
					goto done
				}
			}
		done:
			// Send list to client
			if err := sendList(); err != nil {
				return err
			}
			// Wait a second before sending anything more to client.
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return nil
			}
		}
	})
	g.Go(func() error {
		for {
			_, p, err := conn.ReadMessage()
			if err != nil {
				return fmt.Errorf("reading websocket message: %w", err)
			}

			values, err := url.ParseQuery(string(p))
			if err != nil {
				return fmt.Errorf("parsing query: %w", err)
			}

			// Serialize access to opts, which is read by the other go
			// routine.
			mu.Lock()
			if err := decode.Decode(&opts, values); err != nil {
				return fmt.Errorf("decoding query: %w", err)
			}
			mu.Unlock()

			if err := sendList(); err != nil {
				return err
			}
		}
	})
	if err := g.Wait(); err != nil {
		// Don't log errors resulting from the client closing the connection.
		closeError := &websocket.CloseError{}
		if !errors.As(err, &closeError) {
			h.Error(err, "terminated websocket connection")
		}
	}
}
