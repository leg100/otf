package resource

// Action is a verb that describes the thing being done to a resource.
type Action string

func (k Action) String() string {
	return string(k)
}

const (
	Get    Action = "get"
	List   Action = "list"
	Create Action = "create"
	New    Action = "new"
	Edit   Action = "edit"
	Update Action = "update"
	Delete Action = "delete"
)
