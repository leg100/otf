package http

import (
	"context"
	"net/http"
)

type devModeKey struct{}

func DevModeFromContext(ctx context.Context) bool {
	if mode, ok := ctx.Value(devModeKey{}).(bool); ok {
		return mode
	}
	return false
}

type requestKey struct{}

func RequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(requestKey{}).(*http.Request); ok {
		return r
	}
	return nil
}
