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

Github OAuth Client ID. Set this flag along with [--github-client-secret](#-github-client-secret) to enable [Github authentication](../../auth/providers/github).

## `--github-client-secret`

* System: `otfd`
* Default: ""

Github OAuth client secret. Set this flag along with [--github-client-id](#-github-client-id) to enable [Github authentication](../../auth/providers/github).

## `--gitlab-client-id`

* System: `otfd`
* Default: ""

Gitlab OAuth Client ID. Set this flag along with [--gitlab-client-secret](#-gitlab-client-secret) to enable [Gitlab authentication](../../auth/providers/gitlab).

## `--gitlab-client-secret`

* System: `otfd`
* Default: ""

Gitlab OAuth client secret. Set this flag along with [--gitlab-client-id](#-gitlab-client-id) to enable [Gitlab authentication](../../auth/providers/gitlab).

## `--google-jwt-audience`

* System: `otfd`
* Default: ""

The Google JWT audience claim for validation. If unspecified then the audience
claim is not validated. See the [Google IAP](../../auth/providers/iap#verification) document for more details.

## `--hostname`

* System: `otfd`
* Default: `localhost:8080` or `--address` if specified.

Sets the hostname that clients can use to access the OTF cluster. This value is
used within links sent to various clients, including:

* The `terraform` CLI when it is streaming logs for a remote `plan` or `apply`.
* Pull requests on VCS providers, e.g. the link beside the status check on a
Github pull request.

It is highly advisable to set this flag in a production deployment.

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

## `--plugin-cache`

* System: `otfd`, `otf-agent`
* Default: disabled

Enable the [terraform plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache).

Each plan and apply starts afresh without any provider plugins. They first invoke `terraform init`, which downloads plugins from registries. Given that plugins can be quite large this can use a lot of bandwidth. The terraform [plugin cache](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache) avoids this by caching plugins in a shared directory.

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

The users must exist on the system. Any users that were previously promoted and
are no longer specified with this flag are demoted.

## `--site-token`

* System: `otfd`
* Default: ""

The site token for authenticating with the [`site-admin`](../../auth/site_admin) user, e.g.:

```bash
otfd --site-token=643f57a1016cdde7e7e39914785d36d61fd
```

The default, an empty string, disables the site admin account.

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
