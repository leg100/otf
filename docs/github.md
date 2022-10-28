# Configure Github sign-in

You can configure otf to prompt users to sign in using their github account. Upon sign in, their organizations and teams are automatically synchronised across to otf.

!!! Note
    Admins of a github organization are made members of the privileged `owners` team in otf.

## Steps

Create an OAuth application in github by following their [step-by-step instructions](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app).

* Set application name to something appropriate, e.g. `otf`
* Set the homepage URL to the URL of your otfd installation (although this is purely informational).
* Set an optional description.
* Set the authorization callback URL to:

    `https://<otfd_install_hostname>/oauth/github/callback`

Once you've registered the application, note the client ID and secret.

Set the following flags when running otfd:

    `--github-client-id=<client_id>`
    `--github-client-secret=<client_secret>`

If you're using github enterprise you'll also need to inform otfd of its hostname:

    `--github-hostname=<hostname>`

Now when you start `otfd` navigate to its URL in your browser and you'll be prompted to login with github:

> ![screenshot](login-with-github.png)
