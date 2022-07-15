package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
)

var _ otf.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	// map org name to org id
	mapper map[string]string

	db *sql.DB
	es otf.EventService

	logr.Logger
}

func NewOrganizationService(db *sql.DB, logger logr.Logger, es otf.EventService) (*OrganizationService, error) {
	svc := &OrganizationService{
		db:     db,
		es:     es,
		Logger: logger,
	}

	// Populate mapper
	opts := otf.OrganizationListOptions{}
	for {
		listing, err := svc.List(context.Background(), opts)
		if err != nil {
			return nil, err
		}
		if svc.mapper == nil {
			// allocate map now we know how many runs there are
			svc.mapper = make(map[string]string, listing.TotalCount())
		}
		for _, org := range listing.Items {
			svc.mapper[org.Name()] = org.ID()
		}
		if listing.NextPage() == nil {
			break
		}
		opts.PageNumber = *listing.NextPage()
	}

	return svc, nil
}

func (s OrganizationService) Create(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	org, err := otf.NewOrganization(opts)
	if err != nil {
		return nil, err
	}

	if err := s.db.CreateOrganization(ctx, org); err != nil {
		s.Error(err, "creating organization", "id", org.ID())
		return nil, err
	}

	s.mapper[org.Name()] = org.ID()

	s.es.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	s.V(0).Info("created organization", "id", org.ID(), "name", org.Name())

	return org, nil
}

func (s OrganizationService) EnsureCreated(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	org, err := s.db.GetOrganization(ctx, *opts.Name)
	if err == nil {
		return org, nil
	}

	if err != otf.ErrResourceNotFound {
		s.Error(err, "retrieving organization", "name", *opts.Name)
		return nil, err
	}

	return s.Create(ctx, opts)
}

func (s OrganizationService) Get(ctx context.Context, name string) (*otf.Organization, error) {
	org, err := s.db.GetOrganization(ctx, name)
	if err != nil {
		s.Error(err, "retrieving organization", "name", name)
		return nil, err
	}

	s.V(2).Info("retrieved organization", "name", name, "id", org.ID())

	return org, nil
}

func (s OrganizationService) List(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return s.db.ListOrganizations(ctx, opts)
}

func (s OrganizationService) Update(ctx context.Context, name string, opts *otf.OrganizationUpdateOptions) (*otf.Organization, error) {
	org, err := s.db.UpdateOrganization(ctx, name, func(org *otf.Organization) error {
		return otf.UpdateOrganizationFromOpts(org, *opts)
	})
	if err != nil {
		s.Error(err, "updating organization", "name", name)
		return nil, err
	}
	// update mapping if name changed
	if org.Name() != name {
		s.mapper[org.Name()] = org.ID()
		delete(s.mapper, name)
	}

	s.V(2).Info("updated organization", "name", name, "id", org.ID())

	return org, nil
}

func (s OrganizationService) Delete(ctx context.Context, name string) error {
	err := s.db.DeleteOrganization(ctx, name)
	if err != nil {
		return err
	}
	delete(s.mapper, name)
	return nil
}

func (s OrganizationService) GetEntitlements(ctx context.Context, organizationName string) (*otf.Entitlements, error) {
	org, err := s.Get(ctx, organizationName)
	if err != nil {
		return nil, err
	}
	return otf.DefaultEntitlements(org.ID()), nil
}
