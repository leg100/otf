package integration

import (
	"fmt"
	"testing"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/testutils"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
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
	err := page.Locator("#new-pool-details").Click()
require.NoError(t, err)
	// //screenshot(t, "new_agent_pool"),
	//

	// enter name for new agent pool
	err = page.Locator("input#new-pool-name").Fill("pool-1")
	require.NoError(t, err)

	// submit form
	err := page.Locator(`//button[text()='Create agent pool']`).Click()
	require.NoError(t, err)
	////screenshot(t, "created_agent_pool"),

	// expect flash message confirming pool creation
	err := expect.Locator(page.Locator(`//div[@role='alert']`)).ToHaveText(`created agent pool: pool-1`)
	require.NoError(t, err)

	// confirm pool was created
	created := <-poolsSub
	assert.Equal(t, pubsub.CreatedEvent, created.Type)
	assert.Equal(t, "pool-1", created.Payload.Name)

	// capture agent token
	var agentToken string

	// grant and assign workspace to agent pool, and create agent token.
	page := browser.New(t, ctx)
		// go back to agent pool
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
require.NoError(t, err)
		// grant access to specific workspace
		err := page.Locator(`input#workspaces-specific`).Click()
require.NoError(t, err)
		chromedp.Focus(`input#workspace-input`, chromedp.ByQuery, chromedp.NodeVisible),
		//screenshot(t, "agent_pool_grant_workspace_form"),
		input.InsertText(ws1.Name),

		err := page.Locator(fmt.Sprintf(`//button[@id='%s']`, ws1.ID)).Click()
		require.NoError(t, err)

		// ws1 should appear in list of granted workspaces
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name)),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		//screenshot(t, "agent_pool_granted_workspace"),
		// ws1 should still appear in list of granted workspaces
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name)),
		// go to workspaces
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
require.NoError(t, err)
		// go to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
		// select agent execution mode radio button
		err := page.Locator(`//input[@id='agent']`).Click()
require.NoError(t, err)
		chromedp.ScrollIntoView(`//a[@id='agent-pools-link']`),
		//screenshot(t, "workspace_select_agent_execution_mode"),
		// save changes to workspace
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm execution mode change has persisted
		chromedp.WaitVisible(`input#agent:checked`, chromedp.ByQuery),
		// go back to agent pool
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
require.NoError(t, err)
		//screenshot(t, "agent_pool_workspace_granted_and_assigned"),
		// confirm workspace is now listed under 'Granted & Assigned'
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='granted-and-assigned-workspaces']//a[text()='%s']`, ws1.Name)),
		// create agent token
		err := page.Locator("#new-token-details").Click()
require.NoError(t, err)
		//screenshot(t, "agent_pool_open_new_token_form"),
		// enter description for new agent token
		chromedp.Focus("input#new-token-description", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("token-1"),
		// submit form
		err := page.Locator(`//button[text()='Create token']`).Click()
require.NoError(t, err)
		//screenshot(t, "agent_pool_token_created"),
		// expect flash message confirming token creation
		matchRegex(t, `//div[@role='alert']`, `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
		// click clipboard icon to copy token into clipboard
		err := page.Locator(`//div[@role='alert']//img[@id='clipboard-icon']`).Click()
require.NoError(t, err)
		chromedp.Evaluate(`window.navigator.clipboard.readText()`, &agentToken, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	})
	// clipboard should contained agent token (base64 encoded JWT) and no white
	// space.
	assert.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, agentToken)

	// start agent up, configured to use token.
	registered, shutdownAgent := daemon.startAgent(t, ctx, org.Name, "", agentToken, agent.Config{})

	page := browser.New(t, ctx)
		// go back to agent pool
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
require.NoError(t, err)
		// confirm agent is listed
		chromedp.ScrollIntoView(fmt.Sprintf(`//div[@id='item-%s']`, registered.ID)),
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='item-%s']`, registered.ID)),
		//screenshot(t, "agent_pool_with_idle_agent"),
	})

	// shut agent down and wait for it to exit
	shutdownAgent()
	testutils.Wait(t, agentsSub, func(event pubsub.Event[*agent.Agent]) bool {
		return event.Payload.Status == agent.AgentExited
	})

	page := browser.New(t, ctx)
		// go to agent pool
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
require.NoError(t, err)
		// delete the token
		err := page.Locator(`//button[@id="delete-agent-token-button"]`).Click()
require.NoError(t, err)
		//screenshot(t),
		matchText(t, `//div[@role='alert']`, `Deleted token: token-1`),
		// confirm user is notified that they must unassign pool from workspace
		// before pool can be deleted
		chromedp.WaitVisible(fmt.Sprintf(`//ul[@id='unassign-workspaces-before-deletion']/a[text()='%s']`, ws1.Name)),
		// go to workspace
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
require.NoError(t, err)
		// go to workspace settings
		err := page.Locator(`//a[text()='settings']`).Click()
require.NoError(t, err)
		// switch execution mode from agent to remote
		err := page.Locator(`//input[@id='remote']`).Click()
require.NoError(t, err)
		// save changes to workspace
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm execution mode change has persisted
		chromedp.WaitVisible(`input#remote:checked`, chromedp.ByQuery),
		// go to agent pool
		_, err = page.Goto("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID)
require.NoError(t, err)
		// delete the pool
		err := page.Locator(`//button[@id="delete-agent-pool-button"]`).Click()
require.NoError(t, err)
		//screenshot(t),
		matchText(t, `//div[@role='alert']`, `Deleted agent pool: pool-1`),
	})

	// confirm pool was deleted
	testutils.Wait(t, poolsSub, func(event pubsub.Event[*agent.Pool]) bool {
		return event.Type == pubsub.DeletedEvent
	})
}
