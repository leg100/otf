package run

import "github.com/leg100/otf/internal/resource"

const (
	Tail        resource.Action = "tail"
	Apply       resource.Action = "apply"
	Cancel      resource.Action = "cancel"
	ForceCancel resource.Action = "force-cancel"
	Discard     resource.Action = "discard"
	Retry       resource.Action = "retry"
)
