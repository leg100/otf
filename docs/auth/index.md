# Authentication

OTF provides a variety of mechanisms for authenticating users and clients.

## Identity Providers

Authentication of users to the Web UI is delegated to an _identity provider_. Support currently exists for:

* Github OAuth
* Gitlab OAuth
* [Google IAP](/auth/providers/iap)

User authentication is delegated to identity providers. S

Users sign into OTF primarily via an SSO provider. Support currently exists for:

* Github
* Gitlab

Alternatively, an administrator can sign into OTF using a Site Admin token. This should only be used ad-hoc, e.g. to investigate issues.

## Site Admin Token

Set a hardcoded site token providing access to the built-in `site-admin` user. See the documentation for the [`--site-token` flag](/config/flags#-site-token).

!!! note
    Keep the token secure. Anyone with access to the token has complete access to OTF.

You can sign into the web app using the token. Use the link found in the bottom right corner of the login page.

You can also configure the `otf` client CLI and the `terraform` CLI to use this token:

```bash
terraform login <otf hostname>
```

And enter the token when prompted. It'll be persisted to a local credentials file.

!!! note
    This is recommended only for one-off administrative and testing purposes. You should use an [identity provider](#identity-providers) in most cases.
