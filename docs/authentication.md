# Authentication

Users sign into OTF primarily via an SSO provider. Support currently exists for:

* Github
* Gitlab

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
