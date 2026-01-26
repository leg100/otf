# Caching

During the course of a run a number of artifacts may be downloaded from the internet. This can add up to a significant use of your bandwidth, as well as placing a burden on the upstream artifact providers, which may trigger rate limiting. Furthermore, it can slow your runs down while they wait for downloads to complete.

OTF by default provides limited caching of artifacts to skip unnecessary downloads. Engine binaries, i.e. the `terraform` and `tofu` binaries, are cached to the local filesystem. The [`--engine-bins-dir`](config/flags.md#-engine-bins-dir) flag sets the destination directory.

Provider plugins are notoriously tricky to cache. By default they are not cached. Terraform and Tofu allow you to enable a cache by setting the [`TF_PLUGIN_CACHE_DIR`](https://developer.hashicorp.com/terraform/cli/config/config-file#provider-plugin-cache) environment variable to a directory. With OTF you can either set this environment variable, or you can set the [`--plugin-cache`](config/flags.md#-plugin-cache) and [`--plugin-cache-dir`](config/flags.md#-plugin-cache-dir) flags, which take precedence.

However, there are caveats with this cache. It is only effective once the [lock file](https://developer.hashicorp.com/terraform/language/files/dependency-lock) contains checksums for the plugin (another reason why you should always check in your lock files to version control).

A more problematic caveat is that only Tofu (versions 1.10.0 and higher) [supports concurrent use of the cache](https://opentofu.org/docs/cli/config/config-file/#provider-plugin-cache); Terraform doesn't support concurrent use at all. If you enable the plugin cache in OTF and you have multiple terraform init processes operating concurrently, then you're liable to run into fatal errors.

And even then, the concurrency support in Tofu is dependent upon operating system and filesystem support for [file locks](https://en.wikipedia.org/wiki/File_locking). On a linux machine or VM this should be fine. If you're using the [kubernetes executor](executors.md#kubernetes) and you're using a persistent volume for caching then the volume plugin must support file locks. NFS does provide support, so using something like [Filestore for GKE](https://docs.cloud.google.com/filestore/docs/filestore-for-gke) would work. But object storage, such as S3 and GCS, does not.
