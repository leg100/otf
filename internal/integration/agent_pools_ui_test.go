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
	poolsSub, unsub := daemon.WatchAgentPools(ctx)
	defer unsub()

	// subscribe to agent events
	agentsSub, agentsUnsub := daemon.WatchAgents(ctx)
	defer agentsUnsub()

	// capture agent token
	var agentToken string
	browser.Run(t, ctx, chromedp.Tasks{
		// go to org main menu
		chromedp.Navigate(organizationURL(daemon.Hostname(), org.Name)),
		// go to list of agent pools
		chromedp.Click("#agent_pools > a", chromedp.ByQuery),
		screenshot(t),
		// expose new agent pool form
		chromedp.Click("#new-pool-details", chromedp.ByQuery),
		screenshot(t),
		// enter name for new agent pool
		chromedp.Focus("input#new-pool-name", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("pool-1"),
		// submit form
		chromedp.Click(`//button[text()='Create agent pool']`),
		screenshot(t),
		// expect flash message confirming pool creation
		matchText(t, `//div[@role='alert']`, `created agent pool: pool-1`),
		// grant access to specific workspaces
		chromedp.Click(`input#workspaces-specific`, chromedp.ByQuery),
		chromedp.Focus(`input#workspace-input`, chromedp.ByQuery, chromedp.NodeVisible),
		input.InsertText(ws1.Name),
		chromedp.Click(fmt.Sprintf(`//button[@id='%s']`, ws1.ID), chromedp.NodeVisible),
		// ws1 should appear in list of granted workspaces
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name)),
		// submit
		chromedp.Submit(`//button[text()='Save changes']`),
		chromedp.Click("#new-token-details", chromedp.ByQuery),
		// enter description for new agent token
		chromedp.Focus("input#new-token-description", chromedp.NodeVisible, chromedp.ByQuery),
		input.InsertText("token-1"),
		// submit form
		chromedp.Click(`//button[text()='Create token']`),
		screenshot(t),
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

	// confirm pool was created
	created := <-poolsSub
	assert.Equal(t, pubsub.CreatedEvent, created.Type)
	assert.Equal(t, "pool-1", created.Payload.Name)

	// start agent up, configured to use token.
	registered, shutdownAgent := daemon.startAgent(t, ctx, org.Name, "", agentToken, agent.Config{})

	browser.Run(t, ctx, chromedp.Tasks{
		// go to org main menu
		chromedp.Navigate(organizationURL(daemon.Hostname(), org.Name)),
		// go to list of agent pools
		chromedp.Click("#agent_pools > a", chromedp.ByQuery),
		screenshot(t),
		// go to agent pool
		chromedp.Click(fmt.Sprintf("//div[@id='%s']", created.Payload.ID)),
		screenshot(t),
		// confirm agent is listed
		chromedp.WaitVisible(fmt.Sprintf(`//div[@id='item-%s']`, registered.ID)),
	})

	// shut agent down and wait for it to exit
	shutdownAgent()
	testutils.Wait(t, agentsSub, func(event pubsub.Event[*agent.Agent]) bool {
		return event.Payload.Status == agent.AgentExited
	})

	browser.Run(t, ctx, chromedp.Tasks{
		// go to agent pool
		chromedp.Navigate("https://" + daemon.Hostname() + "/app/agent-pools/" + created.Payload.ID),
		// delete the token
		chromedp.Click(`//button[@id="delete-agent-token-button"]`),
		screenshot(t),
		matchText(t, `//div[@role='alert']`, `Deleted token: token-1`),
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
