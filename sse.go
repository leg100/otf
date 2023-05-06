package internal

import (
	"encoding/base64"
	"fmt"
	"io"
)

// WriteSSEEvent writes an server-side-event to w. The data is optionally base64 encoded
// before being written.
func WriteSSEEvent(w io.Writer, data []byte, event EventType, base64encode bool) {
	if base64encode {
		output := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(output, data)
		data = output
	}

	fmt.Fprintf(w, "data: %s\n", data)
	fmt.Fprintf(w, "event: %s\n\n", event)
}
