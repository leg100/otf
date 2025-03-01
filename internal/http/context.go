package http

import (
	"context"
	"net/http"
)

type requestKey struct{}

func RequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(requestKey{}).(*http.Request); ok {
		return r
	}
	return nil
}
