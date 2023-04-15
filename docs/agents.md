# Agents

OTF agents are dedicated processes for executing runs. They are functionally equivalent to [Terraform Cloud Agents](https://developer.hashicorp.com/terraform/cloud-docs/agents).


The `otf-agent` process maintains an outbound connection to the otf server; no inbound connectivity is required. This makes it suited to deployment in parts of your network that are segregated. For example, you may have a kubernetes cluster for which connectivity is only possible within a local subnet. By deploying an agent to the subnet, terraform can connect to the cluster and provision kubernetes resources.

!!! Note
    An agent only handles runs for a single organization.

### Setup agent

* Log into the web app.
* Select an organization. This will be the organization that the agent handles runs on behalf of.
* Ensure you are on the main menu for the organization.
* Select `agent tokens`.
* Click `New Agent Token`.
* Provide a description for the token.
* Click the `Create token`.
* Copy the token to your clipboard (clicking on the token should do this).
* Start the agent in your terminal:

```bash
otf-agent --token <the-token-string> --address <otf-server-hostname>
```

* The agent will confirm it has successfully authenticated:

```bash
2022-10-30T09:15:30Z INF successfully authenticated organization=automatize
```

### Configure workspace

* Login into the web app
* Select the organization in which you created an agent
* Ensure you are on the main menu for the organization.
* Select `workspaces`.
* Select a workspace.
* Click `settings` in the top right menu.
* Set `execution mode` to `agent`
* Click `save changes`.

Now runs for that workspace will be handled by an agent.
