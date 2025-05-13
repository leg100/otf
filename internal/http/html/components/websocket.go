package components

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func SetAllowedOrigins(origins string) {
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
