package configversion

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/surl"
)

type (
	Service interface {
		CreateConfigurationVersion(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
		// CloneConfigurationVersion creates a new configuration version using the
		// config tarball of an existing configuration version.
		CloneConfigurationVersion(ctx context.Context, cvID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
		GetConfigurationVersion(ctx context.Context, id string) (*ConfigurationVersion, error)
		GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (*ConfigurationVersion, error)

		// Upload handles verification and upload of the config tarball, updating
		// the config version upon success or failure.
		UploadConfig(ctx context.Context, id string, config []byte) error

		// Download retrieves the config tarball for the given config version ID.
		DownloadConfig(ctx context.Context, id string) ([]byte, error)

		create(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
		// CloneConfigurationVersion creates a new configuration version using the
		// config tarball of an existing configuration version.
		clone(ctx context.Context, cvID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)

		get(ctx context.Context, id string) (*ConfigurationVersion, error)
		getLatest(ctx context.Context, workspaceID string) (*ConfigurationVersion, error)
		list(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)
		delete(ctx context.Context, cvID string) error

		// Upload handles verification and upload of the config tarball, updating
		// the config version upon success or failure.
		upload(ctx context.Context, id string, config []byte) error

		// Download retrieves the config tarball for the given config version ID.
		download(ctx context.Context, id string) ([]byte, error)
	}
	service struct {
		logr.Logger

		workspace otf.Authorizer

		db    *db
		cache otf.Cache

		*api
	}
	Options struct {
		logr.Logger

		WorkspaceAuthorizer otf.Authorizer
		MaxUploadSize       int64

		otf.Cache
		otf.DB
		*surl.Signer
	}
)

func NewService(opts Options) *service {
	svc := service{
		Logger: opts.Logger,
	}

	svc.workspace = opts.WorkspaceAuthorizer

	svc.db = newPGDB(opts.DB)
	svc.cache = opts.Cache

	svc.api = newAPI(apiOptions{&svc, opts.MaxUploadSize, opts.Signer})

	return &svc
}

func (s *service) CreateConfigurationVersion(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	return s.svc.create(ctx, workspaceID, opts)
}

func (s *service) CloneConfigurationVersion(ctx context.Context, cvID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	return s.svc.clone(ctx, cvID, opts)
}

func (s *service) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error) {
	// convert from list of concrete CVs to list of interface CVs
	from, err := s.svc.list(ctx, workspaceID, opts)
	if err != nil {
		return nil, err
	}
	to := ConfigurationVersionList{
		Pagination: from.Pagination,
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, i)
	}
	return &to, nil
}

func (s *service) GetConfigurationVersion(ctx context.Context, cvID string) (*ConfigurationVersion, error) {
	return s.svc.get(ctx, cvID)
}

func (s *service) GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (*ConfigurationVersion, error) {
	return s.svc.getLatest(ctx, workspaceID)
}

func (s *service) DeleteConfigurationVersion(ctx context.Context, cvID string) error {
	return s.svc.delete(ctx, cvID)
}

func (s *service) UploadConfig(ctx context.Context, workspaceID string, config []byte) error {
	return s.svc.upload(ctx, workspaceID, config)
}

func (s *service) DownloadConfig(ctx context.Context, workspaceID string) ([]byte, error) {
	return s.svc.download(ctx, workspaceID)
}

func (a *service) create(ctx context.Context, workspaceID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.CreateConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		a.Error(err, "constructing configuration version", "id", cv.ID, "subject", subject)
		return nil, err
	}
	if err := a.db.CreateConfigurationVersion(ctx, cv); err != nil {
		a.Error(err, "creating configuration version", "id", cv.ID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("created configuration version", "id", cv.ID, "subject", subject)
	return cv, nil
}

func (a *service) clone(ctx context.Context, cvID string, opts ConfigurationVersionCreateOptions) (*ConfigurationVersion, error) {
	cv, err := a.get(ctx, cvID)
	if err != nil {
		return nil, err
	}

	cv, err = a.create(ctx, cv.WorkspaceID, opts)
	if err != nil {
		return nil, err
	}

	config, err := a.download(ctx, cvID)
	if err != nil {
		return nil, err
	}

	if err := a.upload(ctx, cv.ID, config); err != nil {
		return nil, err
	}

	return cv, nil
}

func (a *service) list(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.ListConfigurationVersionsAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cvl, err := a.db.ListConfigurationVersions(ctx, workspaceID, ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
	if err != nil {
		a.Error(err, "listing configuration versions", "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed configuration versions", "subject", subject)
	return cvl, nil
}

func (a *service) get(ctx context.Context, cvID string) (*ConfigurationVersion, error) {
	subject, err := a.canAccess(ctx, rbac.GetConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	cv, err := a.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		a.Error(err, "retrieving configuration version", "id", cvID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved configuration version", "id", cvID, "subject", subject)
	return cv, nil
}

func (a *service) getLatest(ctx context.Context, workspaceID string) (*ConfigurationVersion, error) {
	subject, err := a.workspace.CanAccess(ctx, rbac.GetConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := a.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		a.Error(err, "retrieving latest configuration version", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved latest configuration version", "workspace_id", workspaceID, "subject", subject)
	return cv, nil
}

func (a *service) delete(ctx context.Context, cvID string) error {
	subject, err := a.canAccess(ctx, rbac.DeleteConfigurationVersionAction, cvID)
	if err != nil {
		return err
	}

	err = a.db.DeleteConfigurationVersion(ctx, cvID)
	if err != nil {
		a.Error(err, "deleting configuration version", "id", cvID, "subject", subject)
		return err
	}
	a.V(2).Info("deleted configuration version", "id", cvID, "subject", subject)
	return nil
}

func (a *service) canAccess(ctx context.Context, action rbac.Action, cvID string) (otf.Subject, error) {
	cv, err := a.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		return nil, err
	}
	return a.workspace.CanAccess(ctx, action, cv.WorkspaceID)
}
