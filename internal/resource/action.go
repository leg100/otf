package resource

// Action is a verb that describes the thing being done to a resource.
type Action string

func (k Action) String() string {
	return string(k)
}

// Standard set of actions.
const (
	Get             Action = "get"
	List            Action = "list"
	Create          Action = "create"
	New             Action = "new"
	Edit            Action = "edit"
	Update          Action = "update"
	Delete          Action = "delete"
	Watch           Action = "watch"
	Upload          Action = "upload"
	Download        Action = "download"
	Apply           Action = "apply"
	Cancel          Action = "cancel"
	ForceCancel     Action = "force-cancel"
	Discard         Action = "discard"
	Retry           Action = "retry"
	Lock            Action = "lock"
	ForceUnlock     Action = "force-unlock"
	Unlock          Action = "unlock"
	Add             Action = "add"
	Remove          Action = "remove"
	SetPermission   Action = "set-permission"
	UnsetPermission Action = "unset-permission"
	EnqueuePlan     Action = "enqueue-plan"
	Rollback        Action = "rollback"
)
