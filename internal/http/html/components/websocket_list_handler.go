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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebsocketListHandler handles dynamically updating lists of resources via
// a websocket.
type WebsocketListHandler[Resource any, Options any] struct {
	logr.Logger
	Client    websocketListHandlerClient[Resource, Options]
	Populator TablePopulator[Resource]
}

type websocketListHandlerClient[Resource any, Options any] interface {
	Watch(ctx context.Context) (<-chan pubsub.Event[Resource], func())
	List(ctx context.Context, opts Options) (*resource.Page[Resource], error)
}

func (h *WebsocketListHandler[Resource, Options]) Handler(w http.ResponseWriter, r *http.Request) {
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
	var (
		mu   sync.Mutex
		opts Options
	)

	sendList := func() error {
		mu.Lock()
		defer mu.Unlock()

		page, err := h.Client.List(r.Context(), opts)
		if err != nil {
			return fmt.Errorf("fetching workspaces: %w", err)
		}

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		defer w.Close()

		comp := Table(h.Populator, page)
		if err := html.RenderSnippet(comp, w, r); err != nil {
			return fmt.Errorf("rendering html: %w", err)
		}
		return nil
	}

	// Two go-routines:
	// 1) Watch server events and upon receiving an event send a new list to the
	// client. This is necesary because the event can be a notification that a
	// resource has been created, updated or deleted, any of which alters the
	// list of resources on the client.
	// 2) Receive messages from the client altering the list of resources to
	// retrieve (e.g. next page, filtering by status, etc), and send new list to
	// the client.
	g := errgroup.Group{}
	g.Go(func() error {
		// To avoid overwhelming the client, do not send a list more than once a
		// second.
		for {
			// Block on receiving an event.
			select {
			case <-sub:
			case <-r.Context().Done():
				return nil
			}
			// Then consume any remaining events.
			for {
				select {
				case <-sub:
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
			case <-r.Context().Done():
				return nil
			}
		}
	})
	g.Go(func() error {
		for {
			_, p, err := conn.ReadMessage()
			closeError := &websocket.CloseError{}
			if errors.As(err, &closeError) {
				return nil
			} else if err != nil {
				return fmt.Errorf("reading websocket message: %w", err)
			}

			values, err := url.ParseQuery(string(p))
			if err != nil {
				return fmt.Errorf("parsing query: %w", err)
			}
			var msg Options
			if err := decode.Decode(&msg, values); err != nil {
				return fmt.Errorf("decoding query: %w", err)
			}

			// Serialize access to opts, which is read by the other go
			// routine.
			mu.Lock()
			opts = msg
			mu.Unlock()

			if err := sendList(); err != nil {
				return err
			}
		}
	})
	if err := g.Wait(); err != nil {
		h.Error(err, "handling websocket connection")
	}
}
