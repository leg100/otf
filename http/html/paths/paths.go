// Package paths provides rails-style path helpers for use with the web app.
package paths

//go:generate go run gen.go

const (
	// site-wide prefix added to all web UI routes
	UIPrefix = "/app"
)
