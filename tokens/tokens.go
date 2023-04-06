// Package tokens manages token authentication
package tokens

import (
	"time"
)

const (
	// default user session expiry
	defaultExpiry = 24 * time.Hour
)
