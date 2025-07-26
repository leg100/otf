# Forgejo

OTF can use Forgejo/Gitea for both authentication (OIDC) and as a VCS provider.  Here's a setup guide.

## Compatibility

At time of writing (May 2025), the APIs of Gitea and Forgejo are largely compatible, allowing OTF to work with both.  There are only some small differences in naming; OTF uses the webhook type "gitea", which is supported by both.

This guide is tested to work with Forgejo 11.0.1, and Gitea 1.24.0.

# Authentication

OTF's general [OIDC](https://docs.otf.ninja/auth/providers/oidc/) instructions apply.  This document only provides some forgejo-specific details.

## Setting up Forgejo/Gitea

Some examples exist in [the forgejo documentation](https://forgejo.org/docs/latest/user/oauth2-provider/#examples).

Forgejo is set up by going to the "Applications" tab of:

1. User settings → Applications → Manage OAuth2 applications
2. An organization page → Settings → Applications
3. Site administration → Integrations → Applications

The differences between these options are who configures/manages it,
and who can log in through it.

Set it up with the following fields:

* The Application Name can be anything.
* The Redirect URI should be set as described in the OTF OIDC instructions.
* The "Confidential client" box should be checked.

It will generate a client ID and client secret, to be given to OTF (below).

## Setting up OTF

The following OTF parameters make sense:

* `--oidc-name` can be anything.  It is never used.
* `--oidc-issuer-url` is the URL of the forgejo server, with a trailing slash.  Example: `https://forgejo.example.com/`.
* `--oidc-scopes` should be `openid,profile`.
* `--oidc-client-id` is the client ID value provided by forgejo.
* `--oidc-client-secret` is the client secret value provided by forgejo.
* `--oidc-username-claim=sub` tells it to use the `sub` field for usernames.

!!! note
    `--oidc-username-claim=sub` is the only secure setting for the Forgejo OIDC Provider, because users are able to modify their own fullname and email fields.  Recent versions of Gitea have a bit more control over that, see the `USER_DISABLED_FEATURES` setting in [the admin section](https://docs.gitea.com/administration/config-cheat-sheet#admin-admin).
    Unfortunately, in Gitea/Forgejo, the `sub` field returns a numeric ID, not a username, which means `--site-admins` must also be specified numerically.

If all goes well, OTF's web UI should redirect you to log in using forgejo.


# VCS

## Requirements

For now, only one instance of forgejo is supported, and its hostname is specified like `--forgejo-hostname=forgejo.example.com`.  It is assumed that the forgejo instance is running TLS on port 443, and that its certificate was signed by a CA which is trusted by OTF.

You will need a personal access token for a user on that Forgejo instance.  It can be either your own user, or a dedicated service account.

The user needs repository administration privileges, as these are necessary to install webhooks.

The personal access token needs the following permissions:

* repository read and write
* user read

## Setup

In OTF, VCS providers are set up within an organization.  Select (or create) an organization, go to the VCS Providers tab, and click New Forgejo VCS Provider (Personal Token).  Give it a name, and paste in the token.

Once the VCS provider is created, you can attach it to a workspace.  Go to the workspace's Settings menu, click "Connect to VCS", select the VCS provider, and select a git repo or type it in.  This will install a webhook, setting up OTF to receive updates for pushes and pull requests.

To verify that it all works, you can go to the repo's settings page, to the Webhooks tab, select the webhook it installed, and click "Test delivery" at the bottom of the page.  If all goes well, OTF will receive the webhook, create a Run, check out the default branch, run `terraform plan` on it.  When you click on the Run, it will show you the log.
