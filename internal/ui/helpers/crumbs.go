package helpers

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

type crumbs struct {
	resolver crumbsResolver
}

type crumb struct {
	Name string
	Link string
}

type crumbsResolver interface {
	Lineage(ctx context.Context, id resource.ID) (parents []resource.Resource, err error)
}

func (c *crumbs) Crumbs(ctx context.Context, res resource.Resource) ([]crumb, error) {
	crumbs := []crumb{
		{
			Name: res.String(),
		},
	}
	parents, err := c.resolver.Lineage(ctx, res.GetID())
	if err != nil {
		return nil, err
	}
	for _, parent := range parents {
	}
	return crumbs, nil
}
