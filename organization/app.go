package organization

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

type application interface {
	create(ctx context.Context, opts otf.OrganizationCreateOptions) (*Organization, error)
	get(ctx context.Context, name string) (*Organization, error)
	list(ctx context.Context, opts ListOptions) (*OrganizationList, error)
	update(ctx context.Context, name string, opts UpdateOptions) (*Organization, error)
	delete(ctx context.Context, name string) error
	getEntitlements(ctx context.Context, organization string) (Entitlements, error)
}

// app is the implementation of application
type app struct {
	otf.Authorizer
	logr.Logger
	otf.PubSubService

	db *pgdb
}

func NewApplication(ctx context.Context, opts ApplicationOptions) (*app, error) {
	app := &app{
		Authorizer: opts.Authorizer,
		Logger:     opts.Logger,
		db:         newDB(opts.DB),
	}

	return app, nil
}

type ApplicationOptions struct {
	otf.Authorizer
	otf.DB
	otf.Renderer
	logr.Logger
}

func (a *app) create(ctx context.Context, opts otf.OrganizationCreateOptions) (*Organization, error) {
	subject, err := a.CanAccessSite(ctx, rbac.CreateOrganizationAction)
	if err != nil {
		return nil, err
	}

	org, err := NewOrganization(opts)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}

	if err := a.db.create(ctx, org); err != nil {
		a.Error(err, "creating organization", "id", org.ID(), "subject", subject)
		return nil, err
	}

	a.Publish(otf.Event{Type: otf.EventOrganizationCreated, Payload: org})

	a.V(0).Info("created organization", "id", org.ID(), "name", org.Name(), "subject", subject)

	return org, nil
}

func (a *app) get(ctx context.Context, name string) (*Organization, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.GetOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := a.db.get(ctx, name)
	if err != nil {
		a.Error(err, "retrieving organization", "name", name, "subject", subject)
		return nil, err
	}

	a.V(2).Info("retrieved organization", "name", name, "id", org.ID(), "subject", subject)

	return org, nil
}

func (a *app) list(ctx context.Context, opts ListOptions) (*OrganizationList, error) {
	subj, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user, ok := subj.(otf.User); ok && !user.IsSiteAdmin() {
		return a.db.listByUser(ctx, user.ID(), opts)
	}
	return a.db.list(ctx, opts)
}

func (a *app) update(ctx context.Context, name string, opts UpdateOptions) (*Organization, error) {
	subject, err := a.CanAccessOrganization(ctx, rbac.UpdateOrganizationAction, name)
	if err != nil {
		return nil, err
	}

	org, err := a.db.update(ctx, name, func(org *Organization) error {
		return org.Update(opts)
	})
	if err != nil {
		a.Error(err, "updating organization", "name", name, "subject", subject)
		return nil, err
	}

	a.V(2).Info("updated organization", "name", name, "id", org.ID(), "subject", subject)

	return org, nil
}

func (a *app) delete(ctx context.Context, name string) error {
	subject, err := a.CanAccessOrganization(ctx, rbac.DeleteOrganizationAction, name)
	if err != nil {
		return err
	}

	err = a.db.delete(ctx, name)
	if err != nil {
		a.Error(err, "deleting organization", "name", name, "subject", subject)
		return err
	}
	a.V(0).Info("deleted organization", "name", name, "subject", subject)

	return nil
}

func (a *app) getEntitlements(ctx context.Context, organization string) (Entitlements, error) {
	_, err := a.CanAccessOrganization(ctx, rbac.GetEntitlementsAction, organization)
	if err != nil {
		return Entitlements{}, err
	}

	org, err := a.get(ctx, organization)
	if err != nil {
		return Entitlements{}, err
	}
	return defaultEntitlements(org.ID()), nil
}

// newOrganizationList constructs a paginated OrganizationList given the list
// options and a complete list of organizations.
func newOrganizationList(opts ListOptions, orgs []*Organization) *OrganizationList {
	low := opts.GetOffset()
	if low > len(orgs) {
		low = len(orgs)
	}
	high := opts.GetOffset() + opts.GetLimit()
	if high > len(orgs) {
		high = len(orgs)
	}
	return &OrganizationList{
		Items:      orgs[low:high],
		Pagination: otf.NewPagination(opts.ListOptions, len(orgs)),
	}
}
