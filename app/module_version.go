package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateModuleVersion(ctx context.Context, opts otf.CreateModuleVersionOptions) (*otf.ModuleVersion, error) {
	// retrieve module first in order to get organization for authorization
	module, err := a.db.GetModuleByID(ctx, opts.ModuleID)
	if err != nil {
		return nil, err
	}
	organization := module.Organization().Name()

	subject, err := a.CanAccessOrganization(ctx, otf.CreateModuleAction, organization)
	if err != nil {
		return nil, err
	}

	version := otf.NewModuleVersion(opts)

	if err := a.db.CreateModuleVersion(ctx, version); err != nil {
		a.Error(err, "creating module version", "organization", organization, "subject", subject, "module_version", version)
		return nil, err
	}
	a.V(0).Info("created module version", "organization", organization, "subject", subject, "module_version", version)
	return version, nil
}

func (a *Application) UploadModuleVersion(ctx context.Context, opts otf.UploadModuleVersionOptions) (*otf.ModuleVersion, error) {
	_, modver, err := a.ModuleVersionUploader.Upload(ctx, opts)
	if err != nil {
		a.Error(err, "uploading module version", "module_version_id", opts.ModuleVersionID)
		return nil, err
	}
	if modver.Status() != otf.ModuleVersionStatusOk {
		a.Error(err, "uploading module version", "module_version", modver)
		return modver, err
	}

	a.V(0).Info("uploaded module version", "module_version", modver)
	return modver, nil
}

// DownloadModuleVersion should be accessed via signed URL
func (a *Application) DownloadModuleVersion(ctx context.Context, opts otf.DownloadModuleOptions) ([]byte, error) {
	tarball, err := a.db.DownloadModuleVersion(ctx, opts)
	if err != nil {
		a.Error(err, "downloading module", "module_version_id", opts.ModuleVersionID)
		return nil, err
	}
	a.V(2).Info("downloaded module", "module_version_id", opts.ModuleVersionID)
	return tarball, nil
}
