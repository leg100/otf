package otf

type Team interface {
	Name() string
	Organization() string
	IsOwners() bool

	Subject
}
