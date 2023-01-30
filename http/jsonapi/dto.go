package jsonapi

import (
	"time"
)

// Assembler is capable of assembling itself into a JSON-API DTO object.
type Assembler interface {
	// ToJSONAPI assembles a JSON-API DTO using the current request.
	ToJSONAPI() any
}

// PhaseStatusTimestamps holds the timestamps for individual statuses for a
// phase.
type PhaseStatusTimestamps struct {
	CanceledAt    *time.Time `json:"canceled-at,omitempty"`
	ErroredAt     *time.Time `json:"errored-at,omitempty"`
	FinishedAt    *time.Time `json:"finished-at,omitempty"`
	PendingAt     *time.Time `json:"pending-at,omitempty"`
	QueuedAt      *time.Time `json:"queued-at,omitempty"`
	StartedAt     *time.Time `json:"started-at,omitempty"`
	UnreachableAt *time.Time `json:"unreachable-at,omitempty"`
}
