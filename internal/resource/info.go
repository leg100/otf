package resource

// Info provides brief information about a resource.
type Info struct {
	ID   TfeID  `json:"id"`
	Name string `json:"name"`
}
