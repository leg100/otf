package http

import (
	"bufio"
	"io"

	"github.com/acarl005/stripansi"
)

var _ io.Reader = (*AnsiStripper)(nil)

// AnsiStripper strips ANSI codes and makes the stripped output available via
// Read()
type AnsiStripper struct {
	io.Reader
}

// NewAnsiStripper constructs an ANSI stripper, stripping ANSI codes from the
// input stream.
func NewAnsiStripper(input io.Reader) *AnsiStripper {
	output, pipe := io.Pipe()

	// Strip ansi color codes, line-by-line
	go func() {
		defer pipe.Close()

		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			stripped := stripansi.Strip(scanner.Text())
			pipe.Write([]byte(stripped + "\n"))
		}
		if err := scanner.Err(); err != nil {
			pipe.Write([]byte("error encountered whilst stripping ansi codes: " + err.Error()))
		}
	}()

	return &AnsiStripper{Reader: output}
}
