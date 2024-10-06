package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentPoolsUI demonstrates managing agent pools and tokens via the UI.
func TestAgentPoolsUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	// create some workspaces to assign to pool
	ws1 := daemon.createWorkspace(t, ctx, org)
	//ws2 := daemon.createWorkspace(t, ctx, org)

	// subscribe to agent pool events
	poolsSub, unsub := daemon.Agents.WatchAgentPools(ctx)
	defer unsub()

	// subscribe to agent events
	agentsSub, agentsUnsub := daemon.Agents.WatchAgents(ctx)
	defer agentsUnsub()

	// create agent pool via UI
	page := browser.New(t, ctx)

	// go to org main menu
	_, err := page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
	require.NoError(t, err)

	// go to list of agent pools
	err = page.Locator("#agent_pools > a").Click()
	require.NoError(t, err)
	////screenshot(t),

	// expose new agent pool form
	err = page.Locator("#new-pool-details").Click()
	require.NoError(t, err)
	// //screenshot(t, "new_agent_pool"),
	//

	// enter name for new agent pool
	err = page.Locator("input#new-pool-name").Fill("pool-1")
	require.NoError(t, err)

	// submit form
	err = page.Locator(`//button[text()='Create agent pool']`).Click()
	require.NoError(t, err)
	////screenshot(t, "created_agent_pool"),

	// expect flash message confirming pool creation
	err = expect.Locator(page.Locator(`//div[@role='alert']`)).ToHaveText(`created agent pool: pool-1`)
	require.NoError(t, err)

	// confirm pool was created
	created := <-poolsSub
	assert.Equal(t, pubsub.CreatedEvent, created.Type)
	assert.Equal(t, "pool-1", created.Payload.Name)

	// capture agent token
	var agentToken string

	// grant and assign workspace to agent pool, and create agent token.
	//
	// go back to agent pool
	_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
	require.NoError(t, err)
	// grant access to specific workspace
	err = page.Locator(`input#workspaces-specific`).Click()
	require.NoError(t, err)

	err = page.Locator(`input#workspace-input`).Fill(ws1.Name)
	require.NoError(t, err)
	//screenshot(t, "agent_pool_grant_workspace_form"),

	err = page.Locator(fmt.Sprintf(`//button[@id='%s']`, ws1.ID)).Click()
	require.NoError(t, err)

	// ws1 should appear in list of granted workspaces
	err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name))).ToBeVisible()
	require.NoError(t, err)

	// submit
	err = page.Locator(`//button[text()='Save changes']`).Click()
	require.NoError(t, err)

	//screenshot(t, "agent_pool_granted_workspace"),
	// ws1 should still appear in list of granted workspaces
	err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name))).ToBeVisible()
	require.NoError(t, err)

	// go to workspaces
	_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
	require.NoError(t, err)

	// go to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)

	// select agent execution mode radio button
	err = page.Locator(`//input[@id='agent']`).Click()
	require.NoError(t, err)

	err = page.Locator(`//a[@id='agent-pools-link']`).ScrollIntoViewIfNeeded()
	require.NoError(t, err)
	//screenshot(t, "workspace_select_agent_execution_mode"),

	// save changes to workspace
	err = page.Locator(`//button[text()='Save changes']`).Click()
	require.NoError(t, err)

	// confirm execution mode change has persisted
	err = expect.Locator(page.Locator(`input#agent:checked`)).ToBeVisible()
	require.NoError(t, err)

	// go back to agent pool
	_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
	require.NoError(t, err)

	//screenshot(t, "agent_pool_workspace_granted_and_assigned"),
	// confirm workspace is now listed under 'Granted & Assigned'
	err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='granted-and-assigned-workspaces']//a[text()='%s']`, ws1.Name))).ToBeVisible()
	require.NoError(t, err)

	// create agent token
	err = page.Locator("#new-token-details").Click()
	require.NoError(t, err)
	//screenshot(t, "agent_pool_open_new_token_form"),

	// enter description for new agent token
	err = page.Locator("input#new-token-description").Fill("token-1")
	require.NoError(t, err)

	// submit form
	err = page.Locator(`//button[text()='Create token']`).Click()
	require.NoError(t, err)
	//screenshot(t, "agent_pool_token_created"),

	// expect flash message confirming token creation
	err = expect.Locator(page.Locator(`//div[@role='alert']`)).ToHaveText(regexp.MustCompile(`Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`))
	require.NoError(t, err)
	// click clipboard icon to copy token into clipboard
	err = page.Locator(`//div[@role='alert']//img[@id='clipboard-icon']`).Click()
	require.NoError(t, err)

	_, err = page.Evaluate(`window.navigator.clipboard.readText()`, &agentToken)
	//return p.WithAwaitPromise(true)
	// clipboard should contained agent token (base64 encoded JWT) and no white
	// space.
	assert.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, agentToken)

	// start agent up, configured to use token.
	registered, shutdownAgent := daemon.startAgent(t, ctx, org.Name, "", agentToken, agent.Config{})

	// go back to agent pool
	_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
	require.NoError(t, err)

	// confirm agent is listed
	err = page.Locator(fmt.Sprintf(`//div[@id='item-%s']`, registered.ID)).ScrollIntoViewIfNeeded()
	require.NoError(t, err)

	err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='item-%s']`, registered.ID))).ToBeVisible()
	require.NoError(t, err)
	//screenshot(t, "agent_pool_with_idle_agent"),

	// shut agent down and wait for it to exit
	shutdownAgent()
	testutils.Wait(t, agentsSub, func(event pubsub.Event[*agent.Agent]) bool {
		return event.Payload.Status == agent.AgentExited
	})

	// go to agent pool
	_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
	require.NoError(t, err)

	// delete the token
	err = page.Locator(`//button[@id="delete-agent-token-button"]`).Click()
	require.NoError(t, err)

	//screenshot(t),
	err = expect.Locator(page.Locator(`//div[@role='alert']`)).ToHaveText(`Deleted token: token-1`)
	require.NoError(t, err)

	// confirm user is notified that they must unassign pool from workspace
	// before pool can be deleted
	err = expect.Locator(page.Locator(fmt.Sprintf(`//ul[@id='unassign-workspaces-before-deletion']/a[text()='%s']`, ws1.Name))).ToBeVisible()
	require.NoError(t, err)

	// go to workspace
	_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
	require.NoError(t, err)

	// go to workspace settings
	err = page.Locator(`//a[text()='settings']`).Click()
	require.NoError(t, err)

	// switch execution mode from agent to remote
	err = page.Locator(`//input[@id='remote']`).Click()
	require.NoError(t, err)

	// save changes to workspace
	err = page.Locator(`//button[text()='Save changes']`).Click()
	require.NoError(t, err)

	// confirm execution mode change has persisted
	err = expect.Locator(page.Locator(`input#remote:checked`)).ToBeVisible()
	// go to agent pool
	_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
	require.NoError(t, err)

	// delete the pool
	err = page.Locator(`//button[@id="delete-agent-pool-button"]`).Click()
	require.NoError(t, err)
	//screenshot(t),

	err = expect.Locator(page.Locator(`//div[@role='alert']`)).ToHaveText(`Deleted agent pool: pool-1`)
	require.NoError(t, err)

	// confirm pool was deleted
	testutils.Wait(t, poolsSub, func(event pubsub.Event[*agent.Pool]) bool {
		return event.Type == pubsub.DeletedEvent
	})
}
