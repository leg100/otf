package app

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql"
)

var _ otf.WorkspaceService = (*WorkspaceService)(nil)

type WorkspaceService struct {
	*inmem.Mapper
	db *sql.DB
	f  otf.WorkspaceFactory
	es otf.EventService
	otf.WorkspaceQueue
	*inmem.WorkspaceQueueManager

	logr.Logger
}

func NewWorkspaceService(db *sql.DB, logger logr.Logger, os otf.OrganizationService, es otf.EventService, mapper *inmem.Mapper) (*WorkspaceService, error) {
	svc := &WorkspaceService{
		db:                    db,
		Mapper:                mapper,
		es:                    es,
		f:                     otf.WorkspaceFactory{OrganizationService: os},
		Logger:                logger,
		WorkspaceQueueManager: inmem.NewWorkspaceQueueManager(),
	}
	return svc, nil
}

func (s WorkspaceService) Create(ctx context.Context, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	if !otf.CanAccess(ctx, &opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	ws, err := s.f.NewWorkspace(ctx, opts)
	if err != nil {
		s.Error(err, "constructing workspace", "name", opts.Name)
		return nil, err
	}

	if err := s.db.CreateWorkspace(ctx, ws); err != nil {
		s.Error(err, "creating workspace", "id", ws.ID(), "name", ws.Name(), "organization", ws.OrganizationID())
		return nil, err
	}

	// Create mappings
	s.AddWorkspace(ws)

	// create workspace queue
	s.WorkspaceQueueManager.Create(ws.ID())

	s.V(0).Info("created workspace", "id", ws.ID(), "name", ws.Name(), "organization", ws.OrganizationID())

	s.es.Publish(otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws})

	return ws, nil
}

func (s WorkspaceService) Update(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	if !s.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	if err := opts.Valid(); err != nil {
		s.Error(err, "updating workspace")
		return nil, err
	}

	var oldName string
	ws, err := s.db.UpdateWorkspace(ctx, spec, func(ws *otf.Workspace) error {
		oldName = ws.Name()
		return ws.UpdateWithOptions(ctx, opts)
	})
	if err != nil {
		s.Error(err, "updating workspace", spec.LogFields()...)
		return nil, err
	}

	// update mapper if name changed
	if ws.Name() != oldName {
		s.Mapper.UpdateWorkspace(oldName, ws)
	}

	s.V(0).Info("updated workspace", spec.LogFields()...)

	return ws, nil
}

func (s WorkspaceService) UpdateQueue(run *otf.Run) error {
	return s.WorkspaceQueueManager.Update(run.WorkspaceID(), run)
}

func (s WorkspaceService) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	if !otf.CanAccess(ctx, opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	return s.db.ListWorkspaces(ctx, opts)
}

// ListWatch lists workspaces and then watches for changes to workspaces. Note:
// The options filter the list but not the watch.
func (s WorkspaceService) ListWatch(ctx context.Context, opts otf.WorkspaceListOptions) (<-chan *otf.Workspace, error) {
	if !otf.CanAccess(ctx, opts.OrganizationName) {
		return nil, otf.ErrAccessNotPermitted
	}

	// retrieve workspaces from db
	existing, err := s.db.ListWorkspaces(ctx, opts)
	if err != nil {
		return nil, err
	}
	spool := make(chan *otf.Workspace, len(existing.Items))
	for _, r := range existing.Items {
		spool <- r
	}
	sub, err := s.es.Subscribe("workspace-listwatch")
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

func (s WorkspaceService) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	if !s.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	if err := spec.Valid(); err != nil {
		s.Error(err, "retrieving workspace")
		return nil, err
	}

	ws, err := s.db.GetWorkspace(ctx, spec)
	if err != nil {
		s.Error(err, "retrieving workspace", spec.LogFields()...)
		return nil, err
	}

	s.V(2).Info("retrieved workspace", spec.LogFields()...)

	return ws, nil
}

func (s WorkspaceService) GetQueue(workspaceID string) ([]*otf.Run, error) {
	return s.WorkspaceQueueManager.Get(workspaceID)
}

func (s WorkspaceService) Delete(ctx context.Context, spec otf.WorkspaceSpec) error {
	if !s.CanAccessWorkspace(ctx, spec) {
		return otf.ErrAccessNotPermitted
	}

	// Get workspace so we can publish it in an event after we delete it
	ws, err := s.db.GetWorkspace(ctx, spec)
	if err != nil {
		return err
	}

	if err := s.db.DeleteWorkspace(ctx, spec); err != nil {
		s.Error(err, "deleting workspace", "id", ws.ID(), "name", ws.Name())
		return err
	}

	// Remove mappings
	s.RemoveWorkspace(ws)

	// delete workspace queue
	s.WorkspaceQueueManager.Delete(ws.ID())

	s.es.Publish(otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws})

	s.V(0).Info("deleted workspace", "id", ws.ID(), "name", ws.Name())

	return nil
}

func (s WorkspaceService) Lock(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	if !s.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	ws, err := s.db.LockWorkspace(ctx, spec, opts)
	if err != nil {
		s.Error(err, "locking workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)
		return nil, err
	}

	s.V(1).Info("locked workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)

	return ws, nil
}

func (s WorkspaceService) Unlock(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	if !s.CanAccessWorkspace(ctx, spec) {
		return nil, otf.ErrAccessNotPermitted
	}

	ws, err := s.db.UnlockWorkspace(ctx, spec, opts)
	if err != nil {
		s.Error(err, "unlocking workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)
		return nil, err
	}

	s.V(1).Info("unlocked workspace", append(spec.LogFields(), "requestor", opts.Requestor.String())...)

	return ws, nil
}
