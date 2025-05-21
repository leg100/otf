package components

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/a-h/templ"
	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/resource"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Websocket is an inbound websocket connection, on which snippets of HTML are
// sent from the server to the client.
type Websocket[Resource any] struct {
	*websocket.Conn
	r         *http.Request
	client    websocketClient[Resource]
	component func(Resource) templ.Component
	logger    logr.Logger
}

type websocketClient[Resource any] interface {
	Get(ctx context.Context, id resource.TfeID) (Resource, error)
}

func NewWebsocket[Resource any](
	logger logr.Logger,
	w http.ResponseWriter,
	r *http.Request,
	client websocketClient[Resource],
	component func(Resource) templ.Component,
) (*Websocket[Resource], error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &Websocket[Resource]{
		Conn:      conn,
		client:    client,
		component: component,
		r:         r,
		logger:    logger,
	}, nil
}

// Send an HTML fragment to the client rendered from a template populated with a
// resource corresponding to the given id. True is returned if it succeeded,
// otherwise it returns false, in which case the connection should be closed.
func (s *Websocket[Resource]) Send(id resource.TfeID) bool {
	if err := s.doSend(id); err != nil {
		// Ignore errors to do with the client closing the connection.
		var opError *net.OpError
		if !errors.As(err, &opError) {
			s.logger.Error(err, "sending websocket message", "resource_id", id)
		}
		return false
	}
	return true
}

func (s *Websocket[Resource]) doSend(id resource.TfeID) error {
	resource, err := s.client.Get(s.r.Context(), id)
	if err != nil {
		return fmt.Errorf("fetching resource: %w", err)
	}

	w, err := s.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	if err := html.RenderSnippet(s.component(resource), w, s.r); err != nil {
		return fmt.Errorf("rendering html: %w", err)
	}
	return nil
}

// SetAllowedOrigins sets the allowed origins for inbound websocket connections.
// If origins is empty, all origins are allowed. Otherwise origins is expected
// to be a comma separated list of origins and if the origin on an inbound
// websocket connection does not match one of those origins then the connection
// is rejected.
func SetAllowedOrigins(origins string) {
	if len(origins) == 0 {
		return
	}
	sl := strings.Split(strings.ToLower(origins), ",")
	sm := map[string]bool{}
	for _, o := range sl {
		o = strings.TrimPrefix(o, "https://")
		o = strings.TrimPrefix(o, "http://")
		sm[o] = true
	}
	if len(sm) > 0 {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			origins := r.Header["Origin"]
			if len(origins) == 0 {
				return true
			}
			u, err := url.Parse(origins[0])
			if err != nil {
				return false
			}
			origin := strings.ToLower(u.Host)
			_, ok := sm[origin]
			return ok
		}
	}
}
