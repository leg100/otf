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

## `--default-engine`

* System: `otfd`
* Default: `terraform`

Specifies the default engine for new workspaces. Specify either `terraform` or `tofu`.

## `--delete-configs-after`

* System: `otfd`
* Default: `0`

Deletes configs older than the specified age. Specifying `0` disables config deletion.

A config is the tarball of terraform configuration usually created for each run (retrying a run re-uses the existing run's config). Over a long period of time it can consume a lot of database disk space and the only other way to delete configs is to delete the parent workspace.

Deleting a config also deletes any runs that use that config.

Note that the only valid time units are `s`, `m`, and `h`. To specify longer periods of time you need to perform the necessary arithmetric, e.g. for 180 days, 180 x 24, which is `4320h`.

## `--delete-runs-after`

* System: `otfd`
* Default: `0`

Deletes runs older than the specified age. Specifying `0` disables run deletion.

Deleting a run does not delete its associated config. To delete both the run and the config use `--delete-configs-after` instead.

Note that the only valid time units are `s`, `m`, and `h`. To specify longer periods of time you need to perform the necessary arithmetric, e.g. for 180 days, 180 x 24, which is `4320h`.


## `--engine-bins-dir`

* System: `otfd`, `otf-agent`
* Default: `/tmp/otf-engine-bins`

Sets the directory in which engine binaries are downloaded.

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
* Default: the value of `--address`

Sets the hostname advertised to external clients, for example:

* The hostname within the link beside the status check on a GitHub pull request.
* The hostname to which to send webhook events to trigger runs when a workspace is connected to a GitHub repository (see `--webhook-hostname` below.

It is advisable to set this flag in a production deployment. Otherwise it defaults to the listening address set with `--address` which is unlikely to be accessible to external clients.

## `--webhook-hostname`

* System: `otfd`
* Default: the value of `--hostname`

Overrides `--hostname` specifically for webhooks. This is useful if you want to set a separate firewalled inbound route for VCS providers (such as GitHub) via which to send their webhook events.

## `--executor`

* System: `otfd`, `otf-agent`
* Default: `process`

Specifies how runs should be executed.

By default it is set to `process`, which means executables such as `terraform` are forked as child processes of `otfd` (or `otf-agent` if the workspace is set to use an agent).

If set to `kubernetes` then for each plan and apply a Kubernetes job is created. Executables such as `terraform` are then forked as child processes in the job pod.

## `--kubernetes-job-image`

* System: `otfd`, `otf-agent`
* Default: `leg100/otf-job:<current version>`

The container image for kubernetes jobs. The default image tag is set to the version of the `otfd` or `otf-agent` process that creates the job.

## `--kubernetes-job-namespace`

* System: `otfd`, `otf-agent`
* Default: `default`

Kubernetes namespace in which jobs are created.

## `--kubernetes-job-url`

* System: `otfd`
* Default: see below

Specifies the URL that kubernetes jobs use to connect to `otfd`.

The kubernetes executor creates a kubernetes job for each plan and apply. The job needs to connect to `otfd` to execute each plan and apply, so it needs to use a valid URL that allows it to connect to `otfd`.

By default the URL is set to use the value of `--hostname` for the host and `https` for the scheme, e.g. if `--hostname=otf.example.org` is set then the URL is set to `https://example.org/`.

However, if `otfd` is deployed in a kubernetes cluster using the [helm chart](https://github.com/leg100/otf-charts) then the the chart's kubernetes service name, port and namespace is used for the URL, with `http` for the scheme, e.g. if the service name is `otfd`, the service port is `8080`, and the namespace is `otf`, then the URL is set to `http://otfd.otf:8080`. Note that this should be a valid accessible URL as long as the kubernetes job is running on the same cluster as `otfd`.

!!! note
    This flag is not valid in `otf-agent`, which instead instructs the kubernetes job to use the value of the `--url` flag to connect to `otfd`.

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
