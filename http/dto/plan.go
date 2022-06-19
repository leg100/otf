package dto

import "time"

// Plan represents a Terraform Enterprise plan.
type Plan struct {
	ID                   string                `jsonapi:"primary,plans"`
	HasChanges           bool                  `jsonapi:"attr,has-changes"`
	LogReadURL           string                `jsonapi:"attr,log-read-url"`
	ResourceAdditions    int                   `jsonapi:"attr,resource-additions"`
	ResourceChanges      int                   `jsonapi:"attr,resource-changes"`
	ResourceDestructions int                   `jsonapi:"attr,resource-destructions"`
	Status               string                `jsonapi:"attr,status"`
	StatusTimestamps     *PlanStatusTimestamps `jsonapi:"attr,status-timestamps"`
}

// PlanStatusTimestamps holds the timestamps for individual plan statuses.
type PlanStatusTimestamps struct {
	CanceledAt    *time.Time `json:"canceled-at,omitempty"`
	ErroredAt     *time.Time `json:"errored-at,omitempty"`
	FinishedAt    *time.Time `json:"finished-at,omitempty"`
	UnreachableAt *time.Time `json:"unreachable-at,omitempty"`
	QueuedAt      *time.Time `json:"queued-at,omitempty"`
	StartedAt     *time.Time `json:"started-at,omitempty"`
	PendingAt     *time.Time `json:"pending-at,omitempty"`
}
