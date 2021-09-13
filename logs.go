package ots

import "fmt"

const (
	MaxLogsLimit    = 65536
	StartLogsMarker = byte(2)
	EndLogsMarker   = byte(3)
)

// Logs is the output from a terraform task, with options for getting and
// appending a 'chunk' of logs
type Logs []byte

type GetLogOptions struct {
	// The maximum number of bytes of logs to return to the client
	Limit int `schema:"limit"`

	// The start position in the logs from which to send to the client
	Offset int `schema:"offset"`
}

type AppendLogOptions struct {
	// Start indicates this is the first chunk
	Start bool `schema:"start"`

	// End indicates this is the last and final chunk
	End bool `schema:"end"`
}

// Get retrieves a log chunk.
func (l Logs) Get(opts GetLogOptions) ([]byte, error) {
	if len(l) == 0 {
		return nil, nil
	}

	if opts.Offset > len(l) {
		return nil, fmt.Errorf("offset cannot be bigger than total logs")
	}

	if opts.Limit > MaxLogsLimit {
		opts.Limit = MaxLogsLimit
	}

	// Ensure specified chunk does not exceed slice length
	if (opts.Offset + opts.Limit) > len(l) {
		opts.Limit = len(l) - opts.Offset
	}

	return l[opts.Offset:(opts.Offset + opts.Limit)], nil
}

// Append appends a log chunk.
func (l *Logs) Append(logs []byte, opts AppendLogOptions) {
	if opts.Start {
		*l = []byte{StartLogsMarker}
	}

	*l = append(*l, logs...)

	if opts.End {
		*l = append(*l, EndLogsMarker)
	}
}
