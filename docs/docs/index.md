# Introduction

OTF is an open-source alternative to Terraform Enterprise, sharing many of its features:

* Full Terraform CLI integration
* Remote execution mode: plans and applies run on servers
* Agent execution mode: plans and applies run on [agents](runners.md)
* Remote state backend: state stored in PostgreSQL
* SSO: sign in using an identity provider via OIDC, OAuth, etc.
* RBAC: control team access to workspaces
* VCS integration: trigger runs and publish modules from git commits
* Create and install a [Github app](github_app.md) to integrate OTF with Github
* Compatible with much of the [Terraform Enterprise/Cloud API](tfe_api.md)
* Minimal dependencies: requires only PostgreSQL
* Stateless: horizontally scale servers in pods on Kubernetes, etc
* Module registry

...and full support for [OpenTofu](engines.md).

<figure markdown>
![run page planned and finished state](images/run_page_planned_and_finished_state.png){.screenshot}
<figcaption>Real-time streaming of a terraform plan</figcaption>
</figure>

<figure markdown>
![run page planned and finished state](images/github_pull_request_status_check_planned.png){.screenshot}
<figcaption>A status check for a pull request on github.com</figcaption>
</figure>

<figure markdown>
![workspace main page](images/workspace_page.png){.screenshot}
<figcaption>The main page for a workspace</figcaption>
</figure>

<figure markdown>
![workspace main page](images/workspace_edit_trigger_patterns.png){.screenshot}
<figcaption>Setting VCS triggers</figcaption>
</figure>

<figure markdown>
![team permissions](images/team_permissions_added_workspace_manager.png){.screenshot}
<figcaption>Setting organization-level permissions for a team</figcaption>
</figure>

<figure markdown>
![new variable form](images/variables_entering_top_secret.png){.screenshot}
<figcaption>Editing a workspace variable</figcaption>
</figure>

<figure markdown>
![agent pools](./images/agent_pool_workspace_granted_and_assigned.png){.screenshot}
<figcaption>Managing workspace access to specific agent pool</figcaption>
</figure>
