package app

import (
	"context"

	"github.com/leg100/otf"
)

func (a *Application) CreateWorkspace(ctx context.Context, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	if !otf.CanAccess(ctx, &opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	ws, err := a.NewWorkspace(ctx, opts)
	if err != nil {
		a.Error(err, "constructing workspace", "name", opts.Name)
		return nil, err
	}

	if err := a.db.CreateWorkspace(ctx, ws); err != nil {
		a.Error(err, "creating workspace", "id", ws.ID(), "name", ws.Name(), "organization", ws.OrganizationID())
		return nil, err
	}

	// Create mappings
	a.AddWorkspace(ws)

	a.queues.Create(ws.ID())

	a.V(0).Info("created workspace", "id", ws.ID(), "name", ws.Name(), "organization", ws.OrganizationID())

	a.Publish(otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws})

	return ws, nil
}

func (a *Application) UpdateWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	if !a.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	if err := opts.Valid(); err != nil {
		a.Error(err, "updating workspace")
		return nil, err
	}

	var oldName string
	ws, err := a.db.UpdateWorkspace(ctx, spec, func(ws *otf.Workspace) error {
		oldName = ws.Name()
		return ws.UpdateWithOptions(ctx, opts)
	})
	if err != nil {
		a.Error(err, "updating workspace", spec.LogFields()...)
		return nil, err
	}

	// update mapper if name changed
	if ws.Name() != oldName {
		a.Mapper.UpdateWorkspace(oldName, ws)
	}

	a.V(0).Info("updated workspace", spec.LogFields()...)

	return ws, nil
}

func (a *Application) UpdateWorkspaceQueue(run *otf.Run) error {
	return a.queues.Update(run.WorkspaceID(), run)
}

func (a *Application) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	if !otf.CanAccess(ctx, opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	return a.db.ListWorkspaces(ctx, opts)
}

// ListWatchWorkspace lists workspaces and then watches for changes to
// workspaces. Note: The options filter the list but not the watch.
func (a *Application) ListWatchWorkspace(ctx context.Context, opts otf.WorkspaceListOptions) (<-chan *otf.Workspace, error) {
	if !otf.CanAccess(ctx, opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	// retrieve workspaces from db
	existing, err := a.db.ListWorkspaces(ctx, opts)
	if err != nil {
		return nil, err
	}
	spool := make(chan *otf.Workspace, len(existing.Items))
	for _, r := range existing.Items {
		spool <- r
	}
	sub, err := a.Subscribe("workspace-listwatch")
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				// context cancelled; shutdown spooler
				close(spool)
				return
			case event, ok := <-sub.C():
				if !ok {
					// sender closed channel; shutdown spooler
					close(spool)
					return
				}
				ws, ok := event.Payload.(*otf.Workspace)
				if !ok {
					// skip non-workspace events
					continue
				}
				spool <- ws
			}
		}
	}()
	return spool, nil
}

func (a *Application) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	if !a.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	if err := spec.Valid(); err != nil {
		a.Error(err, "retrieving workspace")
		return nil, err
	}

	ws, err := a.db.GetWorkspace(ctx, spec)
	if err != nil {
		a.Error(err, "retrieving workspace", spec.LogFields()...)
		return nil, err
	}

	a.V(2).Info("retrieved workspace", spec.LogFields()...)

	return ws, nil
}

func (a *Application) GetWorkspaceQueue(workspaceID string) ([]*otf.Run, error) {
	return a.queues.Get(workspaceID)
}

func (a *Application) DeleteWorkspace(ctx context.Context, spec otf.WorkspaceSpec) error {
	if !a.CanAccessWorkspace(ctx, spec) {
		return otf.ErrAccessNotPermitted
	}

	// Get workspace so we can publish it in an event after we delete it
	ws, err := a.db.GetWorkspace(ctx, spec)
	if err != nil {
		return err
	}

	if err := a.db.DeleteWorkspace(ctx, spec); err != nil {
		a.Error(err, "deleting workspace", "id", ws.ID(), "name", ws.Name())
		return err
	}

	// Remove mappings
	a.RemoveWorkspace(ws)

	a.queues.Delete(ws.ID())

	a.Publish(otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws})

	a.V(0).Info("deleted workspace", "id", ws.ID(), "name", ws.Name())

	return nil
}

func (a *Application) LockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	if !a.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	ws, err := a.db.LockWorkspace(ctx, spec, opts)
	if err != nil {
		a.Error(err, "locking workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)
		return nil, err
	}

	a.V(1).Info("locked workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)

	return ws, nil
}

func (a *Application) UnlockWorkspace(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	if !a.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	ws, err := a.db.UnlockWorkspace(ctx, spec, opts)
	if err != nil {
		a.Error(err, "unlocking workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)
		return nil, err
	}

	a.V(1).Info("unlocked workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)

	return ws, nil
}
