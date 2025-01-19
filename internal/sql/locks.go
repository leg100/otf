package sql

// Postgres advisory lock IDs for each subsystem to ensure only one of each
// subsystem is running on an OTF cluster. It's important that they don't share the same
// value, hence placing them all in one place makes sense.
const (
	ReporterLockID int64 = iota + 179366396344335597
	TimeoutLockID
	SchedulerLockID
	NotifierLockID
	AllocatorLockID
	RunnerManagerLockID
)
