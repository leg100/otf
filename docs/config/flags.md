# Flags

## `--concurrency`

* System: `otfd`, `otf-agent`
* Default: 5

Sets the number of workers that can process runs concurrently.

## `--google-jwt-audience`

* System: `otfd`
* Default: ""

The Google JWT audience claim for validation. If unspecified then the audience
claim is not validated. See the [Google IAP](/auth/providers/iap#verification) document for more details.

## `--hostname`

* System: `otfd`
* Default: `localhost:8080` or `--address` if specified.

Sets the hostname that clients can use to access the OTF cluster. This value is
used within links sent to various clients, including:

* The `terraform` CLI when it is streaming logs for a remote `plan` or `apply`.
* Pull requests on VCS providers, e.g. the link beside the status check on a
Github pull request.

It is highly advisable to set this flag in a production deployment.

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

Set a site token for authenticating with the [`site-admin`](/auth#site-admins) user, e.g.:

```bash
otfd --site-token=643f57a1016cdde7e7e39914785d36d61fd
```

The token cannot be longer than 64 characters and you should use a cryptographically secure random number generator, for example using `openssl`:

```bash
openssl rand -hex 32
```

The default or an empty string disables use of a site token.
