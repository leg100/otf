// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// CostEstimateStatus represents a costEstimate state.
type CostEstimateStatus string

// List all available costEstimate statuses.
const (
	CostEstimateCanceled              CostEstimateStatus = "canceled"
	CostEstimateErrored               CostEstimateStatus = "errored"
	CostEstimateFinished              CostEstimateStatus = "finished"
	CostEstimatePending               CostEstimateStatus = "pending"
	CostEstimateQueued                CostEstimateStatus = "queued"
	CostEstimateSkippedDueToTargeting CostEstimateStatus = "skipped_due_to_targeting"
)

// CostEstimate represents a Terraform Enterprise costEstimate.
type CostEstimate struct {
	ID                      resource.ID                   `jsonapi:"primary,cost-estimates"`
	DeltaMonthlyCost        string                        `jsonapi:"attribute" json:"delta-monthly-cost"`
	ErrorMessage            string                        `jsonapi:"attribute" json:"error-message"`
	MatchedResourcesCount   int                           `jsonapi:"attribute" json:"matched-resources-count"`
	PriorMonthlyCost        string                        `jsonapi:"attribute" json:"prior-monthly-cost"`
	ProposedMonthlyCost     string                        `jsonapi:"attribute" json:"proposed-monthly-cost"`
	ResourcesCount          int                           `jsonapi:"attribute" json:"resources-count"`
	Status                  CostEstimateStatus            `jsonapi:"attribute" json:"status"`
	StatusTimestamps        *CostEstimateStatusTimestamps `jsonapi:"attribute" json:"status-timestamps"`
	UnmatchedResourcesCount int                           `jsonapi:"attribute" json:"unmatched-resources-count"`
}

// CostEstimateStatusTimestamps holds the timestamps for individual costEstimate statuses.
type CostEstimateStatusTimestamps struct {
	CanceledAt              time.Time `json:"canceled-at"`
	ErroredAt               time.Time `json:"errored-at"`
	FinishedAt              time.Time `json:"finished-at"`
	PendingAt               time.Time `json:"pending-at"`
	QueuedAt                time.Time `json:"queued-at"`
	SkippedDueToTargetingAt time.Time `json:"skipped-due-to-targeting-at"`
}
