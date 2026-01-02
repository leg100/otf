package ui

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/logr"
)

type sseConnection struct {
	http.ResponseWriter
	*http.ResponseController
	base64 bool
	logger logr.Logger
}

type sseEvent string

func newSSEConnection(w http.ResponseWriter, base64 bool) *sseConnection {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	rc := http.NewResponseController(w)
	if err := rc.Flush(); err != nil {
		panic("flush not supported: " + err.Error())
	}
	return &sseConnection{
		ResponseWriter:     w,
		ResponseController: rc,
		base64:             base64,
	}
}

// Send writes an server-side-event to w. The data is optionally base64 encoded
// before being written.
func (conn *sseConnection) Send(data []byte, event sseEvent) {
	if conn.base64 {
		output := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(output, data)
		data = output
	}
	fmt.Fprintf(conn, "data: %s\n", strings.ReplaceAll(string(data), "\n", "&#13;"))
	fmt.Fprintf(conn, "event: %s\n\n", event)

	conn.Flush()
}

func (conn *sseConnection) Render(ctx context.Context, comp templ.Component, event sseEvent) error {
	var buf bytes.Buffer
	if err := comp.Render(ctx, &buf); err != nil {
		conn.logger.Error(err, "rendering html fragment on sse connection")
		return err
	}
	conn.Send(buf.Bytes(), event)
	return nil
}
