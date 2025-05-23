package runner

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWebHandlers_createAgentPool(t *testing.T) {
	organization := organization.NewTestName(t)
	id := testutils.ParseID(t, "pool-123")
	svc := &fakeService{
		pool: &Pool{ID: id},
	}
	h := &webHandlers{svc: svc}
	q := "/?organization_name=" + organization.String() + "&name=my-pool"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.createAgentPool(w, r)

	want := CreateAgentPoolOptions{
		Name:         "my-pool",
		Organization: organization,
	}
	assert.Equal(t, want, svc.createAgentPoolOptions)
	testutils.AssertRedirect(t, w, paths.AgentPool(id))
}

func TestWebHandlers_listAgentPools(t *testing.T) {
	h := &webHandlers{
		svc: &fakeService{
			pool: &Pool{ID: testutils.ParseID(t, "pool-123")},
		},
	}
	q := "/?organization_name=acme-org"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.listAgentPools(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWebHandlers_createAgentToken(t *testing.T) {
	id := testutils.ParseID(t, "pool-123")
	h := &webHandlers{
		svc: &fakeService{},
	}
	q := "/?pool_id=pool-123&description=lorem-ipsum-etc"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()

	h.createAgentToken(w, r)

	testutils.AssertRedirect(t, w, paths.AgentPool(id))
}

func TestAgentToken_DeleteHandler(t *testing.T) {
	agentPoolID := resource.NewTfeID(resource.AgentPoolKind)

	h := &webHandlers{
		svc: &fakeService{
			at: &agentToken{
				AgentPoolID: agentPoolID,
			},
		},
	}
	q := fmt.Sprintf("/?token_id=%s", agentPoolID)
	r := httptest.NewRequest("POST", q, nil)
	w := httptest.NewRecorder()

	h.deleteAgentToken(w, r)

	testutils.AssertRedirect(t, w, paths.AgentPool(agentPoolID))
}
