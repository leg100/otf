package otf

import (
	"encoding/base64"
	"fmt"
	"io"
)

// WriteSSEEvent writes an server-side-event to w. The data is base64 encoded
// before being written.
func WriteSSEEvent(w io.Writer, data []byte, event EventType) {
	output := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(output, data)

	fmt.Fprintf(w, "data: %s\n", output)
	fmt.Fprintf(w, "event: %s\n\n", event)
}
