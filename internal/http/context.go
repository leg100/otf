package http

import "context"

type devModeKey struct{}

func DevModeFromContext(ctx context.Context) bool {
	if mode, ok := ctx.Value(devModeKey{}).(bool); ok {
		return mode
	}
	return false
}
