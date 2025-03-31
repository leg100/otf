package resource

import "context"

type Ancestry struct {
	parentResolvers map[Kind]ParentResolver
}

type ParentResolver func(ctx context.Context, id ID) (ID, error)

func (a *Ancestry) RegisterParentResolver(kind Kind, resolver ParentResolver) {
	a.parentResolvers[kind] = resolver
}
