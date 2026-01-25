# Executors

An *executor* is responsible for running the `plan` and `apply` jobs of a run. There are two executors:

* `fork`
* `kubernetes`

## Fork

Fork is the default executor. It [forks](https://en.wikipedia.org/wiki/Fork_(system_call)) terraform (or tofu) processes as children of the `otfd` or `otf-agent` process.

It requires no further configuration. The maximum number of forked processes is set with the [--concurrency](config/flags.md#-concurrency) flag.

!!! note
    If you want to scale processes beyond a single host with the `fork` executor then you can either run `otfd` on other hosts to form an OTF cluster, or run `otf-agent` on other hosts. See [runners](runners.md) for more details on agents.

## Kubernetes

The kubernetes executor runs plans and apply phases on kubernetes.

An *engine* is the program responsible for executing run commands like `plan` and `apply`. OTF provides support for two engines:

* `terraform`
* `tofu`

The default engine is `terraform`. This can be overridden with the `otfd` flag [`--default-engine`](config/flags.md#-default-engine).

!!! warning
    If you're running more than one instance of `otfd`, take care to set this flag to the same value on each instance. Doing otherwise will lead to unpredictable results.

When you create a workspace, it'll use the default engine. You can override the engine for a workspace in its settings.

When you create a run OTF will download the workspace's engine if it hasn't already been downloaded. The engine binaries are downloaded to the directory specified by the flag [`--engine-bins-dir`](config/flags.md#-engine-bins-dir).
