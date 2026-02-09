// Package helpers provides shared components for templ templates.
package helpers

import (
	"bytes"
	"context"

	"github.com/a-h/templ"
)

func ToString(comp templ.Component) string {
	var buf bytes.Buffer
	comp.Render(context.Background(), &buf)
	return buf.String()
}

func ToBytes(comp templ.Component) []byte {
	var buf bytes.Buffer
	comp.Render(context.Background(), &buf)
	return buf.Bytes()
}
