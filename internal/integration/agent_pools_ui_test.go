package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/runner"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentPoolsUI demonstrates managing agent pools and tokens via the UI.
func TestAgentPoolsUI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)

	// create some workspaces to assign to pool
	ws1 := daemon.createWorkspace(t, ctx, org)

	// subscribe to agent pool events
	poolsSub, unsub := daemon.Runners.WatchAgentPools(ctx)
	defer unsub()

	// subscribe to agent events
	runnersSub, runnersUnsub := daemon.Runners.WatchRunners(ctx)
	defer runnersUnsub()

	// create agent pool via UI
	browser.New(t, ctx, func(page playwright.Page) {
		// go to org main menu
		_, err := page.Goto(organizationURL(daemon.System.Hostname(), org.Name))
		require.NoError(t, err)

		// go to list of agent pools
		err = page.Locator("#menu-item-agent-pools > a").Click()
		require.NoError(t, err)

		// expose new agent pool form
		err = page.Locator("#new-pool-details").Click()
		require.NoError(t, err)
		screenshot(t, page, "new_agent_pool")
		//

		// enter name for new agent pool
		err = page.Locator("input#new-pool-name").Fill("pool-1")
		require.NoError(t, err)

		// submit form
		err = page.Locator(`//button[text()='Create agent pool']`).Click()
		require.NoError(t, err)
		screenshot(t, page, "created_agent_pool")

		// expect flash message confirming pool creation
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`created agent pool: pool-1`)
		require.NoError(t, err)

		// confirm pool was created
		created := <-poolsSub
		assert.Equal(t, pubsub.CreatedEvent, created.Type)
		assert.Equal(t, "pool-1", created.Payload.Name)

		// grant and assign workspace to agent pool, and create agent token.
		//
		// go back to agent pool
		_, err = page.Goto(fmt.Sprintf("https://%s/app/agent-pools/%s", daemon.System.Hostname(), created.Payload.ID))
		require.NoError(t, err)
		// grant access to specific workspace
		err = page.Locator(`input#workspaces-specific`).Click()
		require.NoError(t, err)

		err = page.Locator(`input#workspace-input`).Fill(ws1.Name)
		require.NoError(t, err)
		screenshot(t, page, "agent_pool_grant_workspace_form")

		err = page.Locator(fmt.Sprintf(`//button[@id='%s']`, ws1.ID)).Click()
		require.NoError(t, err)

		// ws1 should appear in list of granted workspaces
		err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name))).ToBeVisible()
		require.NoError(t, err)

		// submit
		err = page.Locator(`//button[text()='Save changes']`).Click()
		require.NoError(t, err)

		screenshot(t, page, "agent_pool_granted_workspace")

		// ws1 should still appear in list of granted workspaces
		err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='granted-workspaces']//a[text()='%s']`, ws1.Name))).ToBeVisible()
		require.NoError(t, err)

		// go to workspace
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
		require.NoError(t, err)

		// go to workspace settings
		err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
		require.NoError(t, err)

		// select agent execution mode radio button
		err = page.Locator(`//input[@id='agent']`).Click()
		require.NoError(t, err)

		err = page.Locator(`//a[@id='agent-pools-link']`).ScrollIntoViewIfNeeded()
		require.NoError(t, err)
		screenshot(t, page, "workspace_select_agent_execution_mode")

		// save changes to workspace
		err = page.Locator(`//button[text()='Save changes']`).Click()
		require.NoError(t, err)

		// confirm execution mode change has persisted
		err = expect.Locator(page.Locator(`input#agent:checked`)).ToBeVisible()
		require.NoError(t, err)

		// go back to agent pool
		_, err = page.Goto(fmt.Sprintf("https://%s/app/agent-pools/%s", daemon.System.Hostname(), created.Payload.ID))
		require.NoError(t, err)

		screenshot(t, page, "agent_pool_workspace_granted_and_assigned")
		// confirm workspace is now listed under 'Granted & Assigned'
		err = expect.Locator(page.Locator(fmt.Sprintf(`//div[@id='granted-and-assigned-workspaces']//a[text()='%s']`, ws1.Name))).ToBeVisible()
		require.NoError(t, err)

		// create agent token
		err = page.Locator("#new-token-details").Click()
		require.NoError(t, err)
		screenshot(t, page, "agent_pool_open_new_token_form")

		// enter description for new agent token
		err = page.Locator("input#new-token-description").Fill("token-1")
		require.NoError(t, err)

		// submit form
		err = page.Locator(`//button[text()='Create token']`).Click()
		require.NoError(t, err)
		screenshot(t, page, "agent_pool_token_created")

		// expect flash message confirming token creation
		err = expect.Locator(page.GetByRole("alert")).ToHaveText(regexp.MustCompile(`Created token:\s+[\w-]+\.[\w-]+\.[\w-]+`))
		require.NoError(t, err)

		// click clipboard icon to copy token into clipboard
		err = page.Locator(`//div[@role='alert']//img[@id='clipboard-icon']`).Click()
		require.NoError(t, err)

		// read token from clipboard
		clipboardContents, err := page.EvaluateHandle(`window.navigator.clipboard.readText()`)
		require.NoError(t, err)
		token := clipboardContents.String()

		// token should be a base64 encoded JWT with no trailing/leading white
		// space.
		require.NotEmpty(t, token)
		require.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, token)

		// start agent up, configured to use token.
		registered, shutdownAgent := daemon.startAgent(t, ctx, org.Name, nil, token)

		// go back to agent pool
		_, err = page.Goto(fmt.Sprintf("https://%s/app/agent-pools/%s", daemon.System.Hostname(), created.Payload.ID))
		require.NoError(t, err)

		// confirm agent is listed
		err = page.Locator(fmt.Sprintf(`//*[@id='item-%s']`, registered.ID)).ScrollIntoViewIfNeeded()
		require.NoError(t, err)

		err = expect.Locator(page.Locator(fmt.Sprintf(`//*[@id='item-%s']`, registered.ID))).ToBeVisible()
		require.NoError(t, err)
		screenshot(t, page, "agent_pool_with_idle_agent")

		// shut agent down and wait for it to exit
		shutdownAgent()
		wait(t, runnersSub, func(event pubsub.Event[*runner.RunnerEvent]) bool {
			return event.Payload.Status == runner.RunnerExited
		})

		// go to agent pool
		_, err = page.Goto(fmt.Sprintf("https://%s/app/agent-pools/%s", daemon.System.Hostname(), created.Payload.ID))
		require.NoError(t, err)

		// delete the token
		err = page.Locator(`//form[@id="delete-agent-token"]/button`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`Deleted token: token-1`)
		require.NoError(t, err)

		// confirm user is notified that they must unassign pool from workspace
		// before pool can be deleted
		err = expect.Locator(page.Locator(fmt.Sprintf(`//ul[@id='unassign-workspaces-before-deletion']/a[text()='%s']`, ws1.Name))).ToBeVisible()
		require.NoError(t, err)

		// go to workspace
		_, err = page.Goto(workspaceURL(daemon.System.Hostname(), org.Name, ws1.Name))
		require.NoError(t, err)

		// go to workspace settings
		err = page.Locator(`//li[@id='menu-item-settings']/a`).Click()
		require.NoError(t, err)

		// switch execution mode from agent to remote
		err = page.Locator(`//input[@id='remote']`).Click()
		require.NoError(t, err)

		// save changes to workspace
		err = page.Locator(`//button[text()='Save changes']`).Click()
		require.NoError(t, err)

		// confirm execution mode change has persisted
		err = expect.Locator(page.Locator(`input#remote:checked`)).ToBeVisible()
		require.NoError(t, err)

		// go to agent pool
		_, err = page.Goto(fmt.Sprintf("https://%s/app/agent-pools/%s", daemon.System.Hostname(), created.Payload.ID))
		require.NoError(t, err)

		// delete the pool
		err = page.Locator(`//button[@id="delete-agent-pool-button"]`).Click()
		require.NoError(t, err)

		err = expect.Locator(page.GetByRole("alert")).ToHaveText(`Deleted agent pool: pool-1`)
		require.NoError(t, err)

		// confirm pool was deleted
		wait(t, poolsSub, func(event pubsub.Event[*runner.Pool]) bool {
			return event.Type == pubsub.DeletedEvent
		})
	})
}
