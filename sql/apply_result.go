package sql

import (
	"time"

	"github.com/leg100/otf"
)

type applyResult interface {
	GetApplyID() *string
	Timestamps
	GetStatus() *string
}

type applyStatusTimestamp interface {
	GetApplyID() *string
	GetStatus() *string
	GetTimestamp() time.Time
}

func addResultToApply(apply *otf.Apply, result applyResult) {
	apply.ID = *result.GetApplyID()
	apply.Timestamps = convertTimestamps(result)
	apply.Status = otf.ApplyStatus(*result.GetStatus())
}

func convertApply(result applyResult) *otf.Apply {
	return &otf.Apply{
		ID:         *result.GetApplyID(),
		Timestamps: convertTimestamps(result),
		Status:     otf.ApplyStatus(*result.GetStatus()),
	}
}

func convertApplyStatusTimestamp(r applyStatusTimestamp) otf.ApplyStatusTimestamp {
	return otf.ApplyStatusTimestamp{
		Status:    otf.ApplyStatus(*r.GetStatus()),
		Timestamp: r.GetTimestamp(),
	}
}
