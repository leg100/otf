package inmem

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapper(t *testing.T) {
	ctx := context.Background()

	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String("test-org")})
	require.NoError(t, err)

	t.Run("add workspace and run from db", func(t *testing.T) {
		// close ch immediately so that Start() exits after populating mapper
		ch := make(chan otf.Event)
		close(ch)

		// Populate db
		ws1 := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
		cv1 := otf.NewTestConfigurationVersion(t, ws1, otf.ConfigurationVersionCreateOptions{})
		run1 := otf.NewRun(cv1, ws1, otf.RunCreateOptions{})
		app := &fakeMapperApp{
			workspaces: []*otf.Workspace{ws1},
			runs:       []*otf.Run{run1},
			events:     ch,
		}

		m := NewMapper(app)
		err = m.Start(ctx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(m.workspaces.idOrgMap))
		assert.Equal(t, org.Name(), m.workspaces.idOrgMap[ws1.ID()])

		assert.Equal(t, 1, len(m.workspaces.nameIDMap))
		assert.Equal(t, ws1.ID(), m.workspaces.nameIDMap[ws1.QualifiedName()])

		assert.Equal(t, 1, len(m.runs.idWorkspaceMap))
		assert.Equal(t, ws1.ID(), m.runs.idWorkspaceMap[run1.ID()])

		t.Run("lookup workspace ID", func(t *testing.T) {
			assert.Equal(t, ws1.ID(), m.LookupWorkspaceID(ws1.SpecName()))
		})
	})

	t.Run("add workspace and run from events", func(t *testing.T) {
		ws1 := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
		cv1 := otf.NewTestConfigurationVersion(t, ws1, otf.ConfigurationVersionCreateOptions{})
		run1 := otf.NewRun(cv1, ws1, otf.RunCreateOptions{})

		app := &fakeMapperApp{
			events: make(chan otf.Event, 2),
		}
		app.events <- otf.Event{Type: otf.EventWorkspaceCreated, Payload: ws1}
		app.events <- otf.Event{Type: otf.EventRunCreated, Payload: run1}
		// terminates Start()
		close(app.events)

		m := NewMapper(app)
		err = m.Start(ctx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(m.workspaces.idOrgMap))
		assert.Equal(t, org.Name(), m.workspaces.idOrgMap[ws1.ID()])

		assert.Equal(t, 1, len(m.workspaces.nameIDMap))
		assert.Equal(t, ws1.ID(), m.workspaces.nameIDMap[ws1.QualifiedName()])

		assert.Equal(t, 1, len(m.runs.idWorkspaceMap))
		assert.Equal(t, ws1.ID(), m.runs.idWorkspaceMap[run1.ID()])
	})

	t.Run("update workspace mapping", func(t *testing.T) {
		ws1 := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})

		app := &fakeMapperApp{
			workspaces: []*otf.Workspace{ws1},
			events:     make(chan otf.Event, 1),
		}

		// rename workspace and send as event
		err := ws1.UpdateWithOptions(ctx, otf.WorkspaceUpdateOptions{
			Name: otf.String("ws-updated"),
		})
		require.NoError(t, err)
		app.events <- otf.Event{Type: otf.EventWorkspaceRenamed, Payload: ws1}
		// terminates Start()
		close(app.events)

		m := NewMapper(app)
		err = m.Start(ctx)
		require.NoError(t, err)

		assert.Equal(t, 1, len(m.workspaces.nameIDMap))
		assert.Contains(t, m.workspaces.nameIDMap, otf.WorkspaceQualifiedName{
			Organization: "test-org",
			Name:         "ws-updated",
		})
	})

	t.Run("remove workspace mapping", func(t *testing.T) {
		ws1 := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})

		app := &fakeMapperApp{
			workspaces: []*otf.Workspace{ws1},
			events:     make(chan otf.Event),
		}
		done := make(chan error)
		m := NewMapper(app)
		go func() {
			done <- m.Start(ctx)
		}()

		// block on sending dummy event so we know Start() has finished adding
		// mapping from db
		app.events <- otf.Event{}

		// should have one ws from db
		assert.Equal(t, 1, len(m.workspaces.idOrgMap))
		assert.Equal(t, 1, len(m.workspaces.nameIDMap))

		// send remove event, instruct Start() to exit and wait for it to exit
		app.events <- otf.Event{Type: otf.EventWorkspaceDeleted, Payload: ws1}
		close(app.events)
		require.NoError(t, <-done)

		assert.Equal(t, 0, len(m.workspaces.nameIDMap))
		assert.Equal(t, 0, len(m.workspaces.idOrgMap))
	})
}
