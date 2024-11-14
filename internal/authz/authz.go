// Package authz handles all things authorization
package authz

import (
	"context"
)

// unexported key types prevents collisions
type (
	skipAuthzCtxKeyType string
)

const (
	skipAuthzCtxKey skipAuthzCtxKeyType = "skip_authz"
)

// AddSkipAuthz adds to the context an instruction to skip authorization.
// Authorizers should obey this instruction using SkipAuthz
func AddSkipAuthz(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipAuthzCtxKey, "")
}

// SkipAuthz determines whether the context contains an instruction to skip
// authorization.
func SkipAuthz(ctx context.Context) bool {
	if v := ctx.Value(skipAuthzCtxKey); v != nil {
		return true
	}
	return false
}
