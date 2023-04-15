# Plugin cache

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
