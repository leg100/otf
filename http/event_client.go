package http

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/leg100/otf"
)

// events implements otf.EventService.
type events struct {
	client *client
	otf.EventService
}

type subscription struct {
	conn *websocket.Conn
	ch   chan otf.Event
}

func (e *events) Subscribe(id string) (otf.Subscription, error) {
	u := url.URL{Scheme: "wss", Host: e.client.baseURL.Host, Path: "/events"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	ch := make(chan otf.Event)
	go func() {
		defer c.Close()
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				ch <- otf.Event{Type: otf.EventError, Payload: fmt.Sprintf("websocket read error: %s\n", err.Error())}
				return
			}
			var ev otf.Event
			if err := json.Unmarshal(msg, &ev); err != nil {
				ch <- otf.Event{Type: otf.EventError, Payload: fmt.Sprintf("websocket decode error: %s\n", err.Error())}
				return
			}
			ch <- ev
		}
	}()
	return &subscription{conn: c, ch: ch}, nil
}

func (s *subscription) C() <-chan otf.Event {
	return s.ch
}

func (s *subscription) Close() error {
	// Cleanly close the connection by sending a close message and then waiting
	// (with timeout) for the server to close the connection.
	err := s.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	return nil
}
