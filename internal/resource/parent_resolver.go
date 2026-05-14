package resource

import (
	"context"
	"fmt"
)

// parentResolver manages the resolution of resource parentage.
type parentResolver struct {
	resolvers map[Kind]resolver
}

type resolver func(ctx context.Context, id ID) (Resource, error)

// Register a resolver with the parent resolver.
func (p *parentResolver) Register(kind Kind, resolver resolver) {
	p.resolvers[kind] = resolver
}

// Parent returns a resource's parent resource.
func (p *parentResolver) Parent(ctx context.Context, id ID) (Resource, error) {
	resolver, ok := p.resolvers[id.Kind()]
	if !ok {
		return nil, fmt.Errorf("kind %v has no parent kind", id.Kind())
	}
	return resolver(ctx, id)
}

// Lineage returns a resource's parent lineage, with direct parent first.
func (p *parentResolver) Lineage(ctx context.Context, id ID) (parents []Resource, err error) {
	for {
		resolver, ok := p.resolvers[id.Kind()]
		if !ok {
			break
		}
		parent, err := resolver(ctx, id)
		if err != nil {
			return nil, err
		}
		parents = append(parents, parent)
	}
	return
}
