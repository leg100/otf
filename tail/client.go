package tail

import (
	"github.com/leg100/otf"
)

var _ otf.TailClient = (*Client)(nil)

// Client is the buffer of logs for a client. The buffer is written to by
// server and read from by the client.
type Client struct {
	server *Server
	phase  otf.PhaseSpec
	buffer chan []byte
}

// Read a section of tailed logs; err is io.EOF when there are no more logs to
// tail.
func (c *Client) Read() <-chan []byte {
	return c.buffer
}

// Close must be called when a client has finished to ensure resources are freed
// up.
func (c *Client) Close() {
	c.server.removeClient(c)
}
