provider "tfe" {
  hostname = "otf.local:8080"
  token    = "abc"
}

resource "tfe_organization" "loadtest" {
  name  = "loadtest"
  email = "admin@company.com"
}

resource "tfe_agent_pool" "loadtest" {
  name         = "loadtest"
  organization = tfe_organization.loadtest.id
}

resource "tfe_agent_token" "loadtest" {
  agent_pool_id = tfe_agent_pool.loadtest.id
  description   = "loadtest"
}

resource "tfe_workspace" "remote" {
  name         = "remote"
  organization = tfe_organization.loadtest.name
}

resource "tfe_workspace" "agent" {
  name         = "agent"
  organization = tfe_organization.loadtest.name
}

resource "tfe_workspace_settings" "agent" {
  workspace_id   = tfe_workspace.agent.id
  execution_mode = "agent"
  agent_pool_id  = tfe_agent_pool.loadtest.id
}

resource "local_file" "agent_token" {
  content  = tfe_agent_token.loadtest.token
  filename = "pool.token"
}
