# OTF

OTF is an open-source alternative to Terraform Enterprise, sharing many of its features:

* Full Terraform CLI integration
* Remote execution mode: plans and applies run on server
* Agent execution mode: plans and applies run on agents
* Remote state backend: state stored in PostgreSQL
* SSO: sign in using Github and gitlab
* Organization and team synchronisation from Github and gitlab
* Module registry (provider registry coming soon)
* Authorization: control team access to workspaces
* VCS integration: trigger runs and publish modules from git commits
* Compatible with much of the Terraform Enterprise/Cloud API
* Minimal dependencies: requires only PostgreSQL
* Stateless: horizontally scale servers in pods on Kubernetes, etc

Feel free to trial it using the demo deployment: [https://demo.otf.ninja](https://demo.otf.ninja)

## Requirements

* Linux - the server and agent components are tested on Linux only; the client CLI is not tested on other platforms but should work.
* PostgreSQL - at least version 12.
* Terraform >= 1.2.0
* An SSL certificate.

## Installation

### Download

There are three components that can be downloaded:

* `otfd` - the server daemon
* `otf` - the client CLI
* `otf-agent` - the agent daemon

Download them from [Github releases](https://github.com/leg100/otf/releases).

The server and agent components are also available as docker images:

* `leg100/otfd`
* `leg100/otf-agent`

### Quick install

These steps will get you started with running the server on your local system.

Download a [release](https://github.com/leg100/otf/releases) of the server component, `otfd`. The release is a zip file. Extract the `otfd` binary to your current directory.

Ensure you have access to a postgres server. OTF by default assumes postgres is running locally, accessible via a domain socket in `/var/run/postgresql`, and defaults to using a database named `otf`. You need to create the database first:

```bash
createdb otf
```

otfd requires a secret for creating cryptographic signatures. It should be up to 64 characters long and you should use a cryptographically secure random number generator, e.g.:

```bash
> openssl rand -hex 32
56789f6076a66323643f57a1016cdde7e7e39914785d36d61fdd8b9a30081f14
```

To get up and running quickly, we'll use the **site admin** account. This account has complete privileges and should only be used for administrative tasks rather than day-to-day usage. To use the account you need to set a token, which can any combination of characters. Make a note of this.

Now start the otf daemon with both the secret and the token:

```bash
> ./otfd --secret=my-secret --site-token=my-token
2022-10-30T20:06:10Z INF started cache max_size=0 ttl=10m0s
2022-10-30T20:06:10Z INF successfully connected component=database path=postgres:///otf?host=/var/run/postgresql
2022-10-30T20:06:10Z INF goose: no migrations to run. current version: 20221017170815 compone
nt=database
2022-10-30T20:06:10Z INF started server address=[::]:8080 ssl=false
```

You have now successfully installed `otfd` and confirmed you can start `otfd` with minimal configuration. Proceed to create your first organization.

#### Create organization

You can navigate to the web app in your browser:

[http://localhost:8080](http://localhost:8080)

Note it announces you have 'no authenticators configured'. The normal method of login is to use SSO signin, via Github etc, but in this quickstart we're using the site admin account. Click on the 'site admin' link in the bottom right, and use your token to login.

Go to 'organizations' and click `New Organization`. Give the organization a name and create.

#### Run Terraform

!!! Note
    The terraform CLI will be connecting to the server and it expects to make a verified SSL connection. Therefore we need to configure SSL first.

Generate a self-signed SSL certificate and key:

```bash
openssl req -x509 -newkey rsa:4096 -sha256 -keyout key.pem -out cert.crt -days 365 -nodes -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost'
```

Ensure your system trusts the generated cert. For example, on Ubuntu based systems:

```bash
sudo cp cert.crt /usr/local/share/ca-certificates
sudo update-ca-certificates
```

Now return to the terminal in which `otfd` is running. You'll need to kill it and start it again, this time with SSL enabled:

```bash
> ./otfd --secret=my-secret --site-token=my-token --ssl --cert-file=cert.crt --key-file=key.pem
```

Terraform needs to use your token to authenticate with `otfd`:

```bash
terraform login localhost:8080
```

Enter `yes` to proceed.

!!! bug
    You'll notice `terraform login` opens a browser window. However it ignores the port, thereby failing to open the correct page on the server. Once you properly deploy the server on a non-custom port this won't be a problem.

Ignore the browser window it has opened and enter your token at the terminal prompt. You should receive confirmation of success:

```
Success! Logged in to Terraform Enterprise (localhost:8080)
```

Now we'll write some terraform configuration. Configure the terraform backend and define a resource:

```bash
cat > main.tf <<EOF
terraform {
  backend "remote" {
    hostname = "localhost:8080"
    organization = "default"

    workspaces {
      name = "dev"
    }
  }
}

resource "null_resource" "quickstart" {}
EOF
```

Initialize terraform:

```bash
terraform init
```

Run a plan:

```bash
terraform plan
```

That starts a run on the server. You can click on the link to the run to view status and logs.

And apply:

```bash
terraform apply
```

This starts another run on the server. Again you can click on the link to see logs.

You have reached the end of this quickstart guide. Have a look at the remainder of the documentation to further complete the installation of OTF, to setup SSO, run agents, etc.

### Install from source

You'll need [Go](https://golang.org/doc/install).

Clone the repo, then build and install using the make task:

```bash
git clone https://github.com/leg100/otf
cd otf
make install
```

That'll install the binaries inside your go bin directory (defaults to `$HOME/go/bin`).

### Install helm chart

You can install the OTF server on Kubernetes using the helm chart.

```bash
helm repo add otf https://leg100.github.io/otf-charts
helm upgrade --install otf otf/otf
```

To see all configurable options with detailed comments:

```
helm show values otf/otf
```

!!! note
    The helm chart is maintained in a separate [github repo](https://github.com/leg100/otf-charts).

### Configuration from env vars

OTF can be configured from environment variables. Arguments can be converted to the equivalent env var by prefixing
it with `OTF_`, replacing all `-` with `_`, and upper-casing it. For example:

- `--secret` becomes `OTF_SECRET`
- `--site-token` becomes `OTF_SITE_TOKEN`

Env variables can be suffixed with `_FILE` to tell otf to read the values from a file. This is useful for container
environments where secrets are often mounted as files.

## Authentication

Users sign into OTF primarily via an SSO provider. Support currently exists for:

* Github
* Gitlab
* OIDC

Alternatively, an administrator can sign into OTF using a Site Admin token. This should only be used ad-hoc, e.g. to investigate issues.

### Github SSO

You can configure OTF to sign users in using their Github account. Upon sign in, their organizations and teams are automatically synchronised across to OTF.

Create an OAuth application in Github by following their [step-by-step instructions](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app).

* Set application name to something appropriate, e.g. `otf`
* Set the homepage URL to the URL of your otfd installation (although this is purely informational).
* Set an optional description.
* Set the authorization callback URL to:

    `https://<otfd_install_hostname>/oauth/github/callback`

Once you've registered the application, note the client ID and secret.

Set the following flags when running otfd:

    `--github-client-id=<client_id>`
    `--github-client-secret=<client_secret>`

If you're using Github Enterprise you'll also need to inform otfd of its hostname:

    `--github-hostname=<hostname>`

Now when you start `otfd` navigate to its URL in your browser and you'll be prompted to login with Github.

#### Organization and team synchronization

Upon sign in, a user's organization and team memberships are synchronised from Github to OTF. If the organization or team does not exist in OTF then it is created.

If the user is an admin of a Github organization then they are made a member of the **owners** team in OTF. The same applies to members of any team named **owners** in Github. Because the owners team has admin privileges across an organization in OTF care should be taken with membership of this team in Github.

### Gitlab SSO

You can configure OTF to sign users in using their Gitlab account. Upon sign in, their Gitlab groups and access levels are synchronised to OTF organizations and teams respectively, e.g. a user who has access level `developers` on the `acme` group in Gitlab will be made a member of the `developers` team in the `acme` organization in OTF.

!!! note
    Only top-level Gitlab groups are synchronised. Sub-groups are ignored.

Create an OAuth application for your Gitlab group by following their [step-by-step instructions](https://docs.gitlab.com/ee/integration/oauth_provider.html#group-owned-applications).

* Set name to something appropriate, e.g. `otf`
* Set the redirect URI to:

    `https://<otfd_install_hostname>/oauth/gitlab/callback`

* Select `Confidential`.
* Select the `read_api` and `read_user` scopes.

Once you've created the application, note the Application ID and Secret.

Set the following flags when running otfd:

    `--gitlab-client-id=<application_id>`
    `--gitlab-client-secret=<secret>`

If you're hosting your own Gitlab you'll also need to inform otfd of its hostname:

    `--gitlab-hostname=<hostname>`

Now when you start `otfd` navigate to its URL in your browser and you'll be prompted to login with Gitlab.

### OIDC

!!! note
    OIDC functions differently from the other authentication providers. It does noes create organizations, or teams as the gitlab, and github authentication providers do.

You can configure OTF to sign users in using [OpenID-Connect](https://openid.net/connect/) (OIDC). The OIDC authentication provider uses an upstream identity provider(idp) such as [Azure AD](https://learn.microsoft.com/en-us/azure/active-directory/develop/v2-protocols-oidc), [Google](https://developers.google.com/identity/openid-connect/openid-connect), or [Dex](https://dexidp.io/).

Create an Application for OTF in your preferred idp.

* Set the name to something appropriate, e.g. `otf`
* Add the following `redirect uri` to the application:

    `https://<otfd_install_hostname>/oauth/<oidc_name>/callback`

Once you've created the application note the client id and secret.

Set the following flags when running otfd: 

* `--oidc-name=<oidc_name>` which is the user-friendly name of the idp. This can be something like `azure-sso`, or `google`.
* `--oidc-issuer-url=<issuer-url>` which is the URL of the idp's oidc configuration. This varies depending on the identity provider used.
* `--oidc-redirect-url=<redirect-uri>` which is the otf url that is invoked by the idp after the authentication process. This should match the redirect uri that was added to the idp's application earlier.
* `--oidc-client-id=<client-id>` which is the `client-id` generated by the idp when we created the application. 
* `--oidc-client-secret=<client-secret>` which is the `client-secret` generated by the idp when we created the application. 

### Site Admin

The site admin user has supreme privileges. To enable the user, set its token when starting `otfd`:

```bash
./otfd --site-token=643f57a1016cdde7e7e39914785d36d61fd
```

Ensure the token is no more than 64 characters. You should use a cryptographically secure random number generator, for example using `openssl`:

```bash
openssl rand -hex 32
```

!!! note
    Keep the token secure. Anyone with access to the token has complete access to otf.

!!! note
    You can also set the token using the environment variable `OTF_SITE_TOKEN`.

You can sign into the web app via a link in the bottom right corner of the login page.

You can also use configure the `otf` client CLI and the `terraform` CLI to use this token:

```bash
terraform login <otf hostname>
```

And enter the token when prompted. It'll be persisted to a local credentials file.

!!! note
    This is recommended only for testing purposes. You should use your SSO account in most cases.

## Authorization

The authorization model largely follows that of TFC/E. An organization comprises a number of teams. A user is a member of one or more teams.

### Owners team

Every organization has an `owners` team. The user that creates an organization becomes its owner. The owners team must have at least one member and it cannot be deleted.

Members of the owners team enjoy broad privileges across an organization. "Owners" are the only users permitted to alter organization-level permissions. They are also automatically assigned all the organization-level permissions; these permissions cannot be unassigned.

### Synchronisation

Upon signing in, a user's organizations and teams are synchronised or "mapped" to those of their SSO provider. If an organization or team does not exist it is created.

The mapping varies according to the SSO provider. If the provider doesn't have the concept of an organization or team then equivalent units of authorization are used. Special rules apply to the mapping of the Owners team too. The exact mappings for each provider are listed here:

|provider|organization|team|owners|
|-|-|-|-|
|Github|organization|team|admin role or "owners" team|
|Gitlab|top-level group|access level|owners access level|

### Personal organization

A user is assigned a personal organization matching their username. They are automatically an owner of this organization. The organization is created the first time the user logs in.

### Permissions

Permissions are assigned to teams on two levels: organizations and workspaces. Organization permissions confer privileges across the organization:

* Manage Workspaces: Allows members to create and administrate all workspaces within the organization.
* Manage VCS Settings: Allows members to manage the set of VCS providers available within the organization.
* Manage Registry: Allows members to publish and delete modules within the organization.

Workspace permissions confer privileges on the workspace alone, and are based on the [fixed permission sets of TFC/TFE](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#fixed-permission-sets):

* Read
* Plan
* Write
* Admin

See the [TFC/TFE documentation](https://developer.hashicorp.com/terraform/cloud-docs/users-teams-organizations/permissions#fixed-permission-sets) for more information on the privileges each permission set confers.


## VCS Providers

To connect workspaces and modules to git repositories containing Terraform configurations, you need to provide OTF with access to your VCS provider.

Firstly, create a provider for your organization. On your organization's main menu, select 'VCS providers'.

You'll be presented with a choice of providers to create. The choice is restricted to those for which you have enabled [SSO](#authentication). For instance, if you have enabled Github SSO then you can create a Github VCS provider.

Select the provider you would like to create. You will then be prompted to enter a personal access token. Instructions for generating the token are included on the page. The token permits OTF to access your git repository and retrieve terraform configuration. Once you've generated and inserted the token into the field you also need to give the provider a name that describes it.

!!! note
    Be sure to restrict the permissions on the token according to the instructions.

Create the provider and it'll appear on the list of providers. You can now proceed to connecting workspaces and publishing modules.

### Connecting a workspace

Once you have a provider you can connect a workspace to a git repository for that provider.

Select a workspace. Go to its 'settings' (in the top right of the workspace page).

Click 'Connect to VCS'.

Select the provider.

You'll then be presented with a list of repositories. Select the repository containing the terraform configuration you want to use in your workspace. If you cannot see your repository you can enter its name.

Once connected you can start a run via the web UI. On the workspace page select the 'start run' drop-down box and select an option to either start a plan or both a plan and an apply.

That will start a run, retrieving the configuration from the repository, and you will see the progress of its plan and apply.

## Module Registry

OTF includes a registry of terraform modules. You can publish modules to the registry from a git repository and source the modules in your terraform configuration.

### Publish module

To publish a module, go to the organization main menu, select 'modules' and click 'publish'.

You then need to select a VCS provider. If none are visible you need to first create a [provider](#vcs-providers).

Connect to the provider and you are presented with a list of git repositories. Select the repository that contains the module you want to publish. If the repository is not visible you can enter its path instead.

!!! note
    Only the first 100 repositories found on your provider are shown.

Once you select a repository, you are asked to confirm your selection. OTF then retrieves the repository's git tags. For each tag that looks like a semantic version, e.g. `v1.0.0` or `0.10.3`, it'll download the contents of the repository for each tag and publish a module with that version. You should then be redirected to the module's page, containing information regarding its resources, inputs and outputs, along with usage instructions.

!!! note
    Ensure your repository has at least one tag that looks like a semantic version. Otherwise OTF will fail to publish the module.

A webhook is also added to the repository. Any tags pushed to the repository will trigger the webhook and new module versions will be published.

## Agents

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

## Plugin Cache

* System: `otfd`, `otf-agent`
* Flag: `--plugin-cache`
* Default: disabled

Each plan and apply starts afresh without any provider plugins. They first invoke `terraform init`, which downloads plugins from registries. Given that plugins can be quite large this can use a lot of bandwidth. Terraform's [plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache) avoids this by caching plugins into a shared directory.

However, enabling the cache causes a [known issue](https://github.com/hashicorp/terraform/issues/28041). If the user is on a different platform to that running OTF, e.g. the user is on a Mac but `otfd` is running on Linux, then you might see an error similar to the following:

```
Error: Failed to install provider from shared cache

Error while importing hashicorp/null v3.2.1 from the shared cache
directory: the provider cache at .terraform/providers has a copy of
registry.terraform.io/hashicorp/null 3.2.1 that doesn't match any of the
checksums recorded in the dependency lock file.
```

The workaround is for users to include checksums for OTF's platform in the lock file too, e.g. if `otfd` is running on Linux on amd64 then they would run the following:

```
terraform providers lock -platform=linux_amd64
```

That'll update `.terraform.lock.hcl` accordingly. This command should be invoked whenever a change is made to the providers and their versions in the configuration.

!!! note
    Another alternative is to configure OTF to use an [HTTPS caching proxy](https://github.com/leg100/squid).

## Client CLI

`otf` is a CLI program for interacting with the server.

Download a [release](https://github.com/leg100/otf/releases). Ensure you select the client component, `otf`. The release is a zip file. Extract the `otf` binary to a directory in your system PATH.

Run `otf` with no arguments to receive usage instructions:

```bash
Usage:
  otf [command]

Available Commands:
  agents        Agent management
  help          Help about any command
  organizations Organization management
  runs          Runs management
  workspaces    Workspace management

Flags:
      --address string   Address of OTF server (default "localhost:8080")
  -h, --help             help for otf

Use "otf [command] --help" for more information about a command.
```

Credentials are sourced from the same file the terraform CLI uses (`~/.terraform.d/credentials.tfrc.json`). To populate credentials, run:

```bash
terraform login <otfd_hostname>
```

!!! note
    `terraform login` has a bug wherein it ignores the port when opening a browser. If you have included a port, e.g. `localhost:8080`, then you'll need to fix the URL in the browser address bar accordingly.

!!! note
    Functionality is presently limited, restricted to a subset of what is available via the web app.

## Development

Many tests, including unit tests, require access to a postgres database. A dedicated database should assigned for this purpose. Tests expect the environment variable `OTF_TEST_DATABASE_URL` to contain a valid postgres connection string.
