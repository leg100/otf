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
	browser.Run(t, ctx, chromedp.Tasks{
		// go to org main menu
		chromedp.Navigate(organizationURL(daemon.System.Hostname(), org.Name)),
		// go to list of agent pools
		chromedp.Click("#agent_pools > a", chromedp.ByQuery),
		screenshot(t),
		// expose new agent pool form
		chromedp.Click("#new-pool-details", chromedp.ByQuery),
		screenshot(t, "new_agent_pool"),
		// enter name for new agent pool
		chromedp.Focus("input#new-pool-name", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("pool-1"),
		// submit form
		chromedp.Click(`//button[text()='Create agent pool']`),
		screenshot(t, "created_agent_pool"),
		// expect flash message confirming pool creation
		matchText(t, `//div[@role='alert']`, `created agent pool: pool-1`),
	})

	// confirm pool was created
	created := <-poolsSub
	assert.Equal(t, pubsub.CreatedEvent, created.Type)
	assert.Equal(t, "pool-1", created.Payload.Name)

	// capture agent token
	var agentToken string

	// grant and assign workspace to agent pool, and create agent token.
	browser.Run(t, ctx, chromedp.Tasks{
		// go back to agent pool
		chromedp.Navigate("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID),
		// grant access to specific workspace
		chromedp.Click(`input#workspaces-specific`, chromedp.ByQuery),
		chromedp.Focus(`input#workspace-input`, chromedp.ByQuery, chromedp.NodeVisible),
		screenshot(t, "agent_pool_grant_workspace_form"),
		input.InsertText(ws1.Name),
		chromedp.Click(fmt.Sprintf(`//button[@id='%s']`, ws1.ID), chromedp.NodeVisible),
		// ws1 should appear in list of granted workspaces
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name)),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		screenshot(t, "agent_pool_granted_workspace"),
		// ws1 should still appear in list of granted workspaces
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name)),
		// go to workspaces
		chromedp.Navigate(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name)),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		// select agent execution mode radio button
		chromedp.Click(`//input[@id='agent']`),
		chromedp.ScrollIntoView(`//a[@id='agent-pools-link']`),
		screenshot(t, "workspace_select_agent_execution_mode"),
		// save changes to workspace
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm execution mode change has persisted
		chromedp.WaitVisible(`input#agent:checked`, chromedp.ByQuery),
		// go back to agent pool
		chromedp.Navigate("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID),
		screenshot(t, "agent_pool_workspace_granted_and_assigned"),
		// confirm workspace is now listed under 'Granted & Assigned'
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='granted-and-assigned-workspaces']//a[text()='%s']`, ws1.Name)),
		// create agent token
		chromedp.Click("#new-token-details", chromedp.ByQuery),
		screenshot(t, "agent_pool_open_new_token_form"),
		// enter description for new agent token
		chromedp.Focus("input#new-token-description", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("token-1"),
		// submit form
		chromedp.Click(`//button[text()='Create token']`),
		screenshot(t, "agent_pool_token_created"),
		// expect flash message confirming token creation
		matchRegex(t, `//div[@role='alert']`, `Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`),
		// click clipboard icon to copy token into clipboard
		chromedp.Click(`//div[@role='alert']//img[@id='clipboard-icon']`),
		chromedp.Evaluate(`window.navigator.clipboard.readText()`, &agentToken, func(p *runtime.EvaluateParams) *runtime.EvaluateParams {
			return p.WithAwaitPromise(true)
		}),
	})
	// clipboard should contained agent token (base64 encoded JWT) and no white
	// space.
	assert.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, agentToken)

	// start agent up, configured to use token.
	registered, shutdownAgent := daemon.startAgent(t, ctx, org.Name, "", agentToken, agent.Config{})

	browser.Run(t, ctx, chromedp.Tasks{
		// go back to agent pool
		chromedp.Navigate("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID),
		// confirm agent is listed
		chromedp.ScrollIntoView(fmt.Sprintf(`//div[@id='item-%s']`, registered.ID)),
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='item-%s']`, registered.ID)),
		screenshot(t, "agent_pool_with_idle_agent"),
	})

	// shut agent down and wait for it to exit
	shutdownAgent()
	testutils.Wait(t, agentsSub, func(event pubsub.Event[*agent.Agent]) bool {
		return event.Payload.Status == agent.AgentExited
	})

	browser.Run(t, ctx, chromedp.Tasks{
		// go to agent pool
		chromedp.Navigate("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID),
		// delete the token
		chromedp.Click(`//button[@id="delete-agent-token-button"]`),
		screenshot(t),
		matchText(t, `//div[@role='alert']`, `Deleted token: token-1`),
		// confirm user is notified that they must unassign pool from workspace
		// before pool can be deleted
		chromedp.WaitVisible(fmt.Sprintf(`//ul[@id='unassign-workspaces-before-deletion']/a[text()='%s']`, ws1.Name)),
		// go to workspace
		chromedp.Navigate(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name)),
		// go to workspace settings
		chromedp.Click(`//a[text()='settings']`),
		// switch execution mode from agent to remote
		chromedp.Click(`//input[@id='remote']`),
		// save changes to workspace
		chromedp.Submit(`//button[text()='Save changes']`),
		// confirm execution mode change has persisted
		chromedp.WaitVisible(`input#remote:checked`, chromedp.ByQuery),
		// go to agent pool
		chromedp.Navigate("https://" + daemon.System.Hostname() + "/app/agent-pools/" + created.Payload.ID),
		// delete the pool
		chromedp.Click(`//button[@id="delete-agent-pool-button"]`),
		screenshot(t),
		matchText(t, `//div[@role='alert']`, `Deleted agent pool: pool-1`),
	})

	// confirm pool was deleted
	testutils.Wait(t, poolsSub, func(event pubsub.Event[*agent.Pool]) bool {
		return event.Type == pubsub.DeletedEvent
	})
}
