package otf

type Team interface {
	ID() string
	Name() string
	Organization() string
	IsOwners() bool
}
