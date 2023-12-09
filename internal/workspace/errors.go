package workspace

import "errors"

var (
	ErrWorkspaceAlreadyLocked         = errors.New("workspace already locked")
	ErrWorkspaceLockedByDifferentUser = errors.New("workspace locked by different user")
	ErrWorkspaceLockedByRun           = errors.New("workspace is locked by Run")
	ErrWorkspaceAlreadyUnlocked       = errors.New("workspace already unlocked")
	ErrWorkspaceUnlockDenied          = errors.New("unauthorized to unlock workspace")
	ErrWorkspaceInvalidLock           = errors.New("invalid workspace lock")
	ErrUnsupportedTerraformVersion    = errors.New("unsupported terraform version")

	ErrTagsRegexAndTriggerPatterns     = errors.New("cannot specify both tags-regex and trigger-patterns")
	ErrTagsRegexAndAlwaysTrigger       = errors.New("cannot specify both tags-regex and always-trigger")
	ErrTriggerPatternsAndAlwaysTrigger = errors.New("cannot specify both trigger-patterns and always-trigger")
	ErrInvalidTriggerPattern           = errors.New("invalid trigger glob pattern")
	ErrInvalidTagsRegex                = errors.New("invalid vcs tags regular expression")
	ErrAgentExecutionModeWithoutPool   = errors.New("agent execution mode requires agent pool ID")
	ErrNonAgentExecutionModeWithPool   = errors.New("agent pool ID can only be specified with agent execution mode")
)
