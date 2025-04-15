# Flags

## `--address`

* System: `otfd`, `otf-agent`
* Default: `localhost:8080`

Sets the listening address of an `otfd` node.

Set the port to an empty string or to `0` to choose a random available port.

Set the address to an empty string to listen on all interfaces. For example, the
following listens on all interfaces using a random port:

```
otfd --address :0
```

## `--applying-timeout`

* System: `otfd`
* Default: `24h`

Sets the amount of time a run is permitted to be in the `applying` state before it is canceled.

## `--cache-expiry`

* System: `otfd`
* Default: `10 minutes`

Set the TTL for cache entries.

## `--cache-size`

* System: `otfd`
* Default: `0` (unlimited)

Cache size in MB. The cache is stored in RAM. Default is `0` which means it'll use an unlimited amount of RAM.

It is recommended that you set this to an appropriate size in a production
deployment, taking into consideration the [cache expiry](#-cache-expiry).

## `--concurrency`

* System: `otfd`, `otf-agent`
* Default: 5

Sets the number of workers that can process runs concurrently.

## `--dev-mode`

* System: `otfd`
* Default: `false`

Enables developer mode:

1. Static files are loaded from disk rather than from those embedded within the `otfd` binary.
2. Enables [livereload](https://github.com/livereload/livereload-js).

This means you can make changes to CSS, templates, etc, and you automatically see the changes in the browser in real-time.

If developer mode were disabled, you would need to re-build the `otfd` binary and then manually reload the page in your browser.

!!! note
    Ensure you have cloned the git repository to your local filesystem and that you have started `otfd` from the root of the repository, otherwise it will not be able to locate the static files.

## `--github-client-id`

* System: `otfd`
* Default: ""

Github OAuth Client ID. Set this flag along with [--github-client-secret](#-github-client-secret) to enable [Github authentication](../auth/providers/github.md).

## `--github-client-secret`

* System: `otfd`
* Default: ""

Github OAuth client secret. Set this flag along with [--github-client-id](#-github-client-id) to enable [Github authentication](../auth/providers/github.md).

## `--gitlab-client-id`

* System: `otfd`
* Default: ""

Gitlab OAuth Client ID. Set this flag along with [--gitlab-client-secret](#-gitlab-client-secret) to enable [Gitlab authentication](../auth/providers/gitlab.md).

## `--gitlab-client-secret`

* System: `otfd`
* Default: ""

Gitlab OAuth client secret. Set this flag along with [--gitlab-client-id](#-gitlab-client-id) to enable [Gitlab authentication](../auth/providers/gitlab.md).

## `--google-jwt-audience`

* System: `otfd`
* Default: ""

The Google JWT audience claim for validation. If unspecified then the audience
claim is not validated. See the [Google IAP](../auth/providers/iap.md#verification) document for more details.

## `--hostname`

* System: `otfd`
* Default: `localhost:8080` or `--address` if specified.

Sets the hostname that clients can use to access the OTF cluster. This value is
used within links sent to various clients, including:

* The `terraform` CLI when it is streaming logs for a remote `plan` or `apply`.
* Pull requests on VCS providers, e.g. the link beside the status check on a
Github pull request.

It is highly advisable to set this flag in a production deployment.

## `--webhook-hostname`

* System: `otfd`
* Default: `localhost:8080` or `--address` if specified.

Sets the hostname that VCS providers can use to access the OTF webhooks.

## `--log-format`

* System: `otfd`, `otf-agent`
* Default: `default`

Set the logging format. Can be one of:

* `default`: human-friendly, not easy to parse, writes to stderr
* `text`: sequence of key=value pairs, writes to stdout
* `json`: json format, writes to stdout

## `--max-config-size`

* System: `otfd`
* Default: `104865760` (10MiB)

Maximum permitted configuration upload size. This refers to the size of the (compressed) configuration tarball that `terraform` uploads to OTF at the start of a remote plan/apply.

## `--oidc-client-id`

* System: `otfd`
* Default: ""

OIDC Client ID. Set this flag along with [--oidc-client-secret](#-oidc-client-secret) to enable [OIDC authentication](../auth/providers/oidc.md).

## `--oidc-client-secret`

* System: `otfd`
* Default: ""

OIDC Client Secret. Set this flag along with [--oidc-client-id](#-oidc-client-id) to enable [OIDC authentication](../auth/providers/oidc.md).

## `--oidc-issuer-url`

* System: `otfd`
* Default: ""

OIDC Issuer URL for OIDC authentication.

## `--oidc-name`

* System: `otfd`
* Default: ""

User friendly OIDC name - this is the name of the OIDC provider shown on the login prompt on the web UI.

## `--oidc-scopes`

* System: `otfd`
* Default: [openid,profile]

OIDC scopes to request from OIDC provider.

## `--oidc-username-claim`

* System: `otfd`
* Default: "name"

OIDC claim for mapping to an OTF username. Must be one of `name`, `email`, or `sub`.

## `--planning-timeout`

* System: `otfd`
* Default: `2h`

Sets the amount of time a run is permitted to be in the `planning` state before it is canceled.

## `--restrict-org-creation`

* System: `otfd`
* Default: false

Restricts the ability to create organizations to users possessing the site admin role. By default _any_ user can create organizations.

## `--sandbox`

* System: `otfd`
* Default: false

Enable sandbox box; isolates `terraform apply` using [bubblewrap](https://github.com/containers/bubblewrap) for additional security.

## `--secret`

* **Required**
* System: `otfd`
* Default: ""

Hex-encoded 16-byte secret for performing cryptographic work. You should use a cryptographically secure random number generator, e.g. `openssl`:

```bash
> openssl rand -hex 16
6b07b57377755b07cf61709780ee7484
```

!!! note
    The secret is required. It must be exactly 16 bytes in size, and it must be hex-encoded.

## `--site-admins`

* System: `otfd`
* Default: []

Promote users to the role of site admin. Specify their usernames, separated by a
comma. For example:

```
otfd --site-admins bob@example.com,alice@example.com
```

Users are automatically created if they don't exist already.

## `--site-token`

* System: `otfd`
* Default: ""

The site token for authenticating with the built-in [`site-admin`](../auth/site_admins.md) user, e.g.:

```bash
otfd --site-token=643f57a1016cdde7e7e39914785d36d61fd
```

The default, an empty string, disables the site admin account.

## `--url`

* System: `otf-agent`, `otf`
* Default: `https://localhost:8080`

Specifies the URL of `otfd` to connect to. You must include the scheme, which is either `https://` or `http://`.

## `--v`, `-v`

* System: `otfd`, `otf-agent`
* Default: `0`

Set logging verbosity. The higher the number the more verbose the logs. Each number translates to a `level` log field like so:

|verbosity|level|
|-|-|
|0|INFO|
|1|DEBUG|
|2|DEBUG-1|
|3|DEBUG-2|
|n|DEBUG-(n+1)|
