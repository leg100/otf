# Engines

An *engine* is the program responsible for executing run commands like `plan` and `apply`. OTF provides support for two engines:

* `terraform`
* `tofu`

The default engine is `terraform`. This can be overridden with the `otfd` flag [`--default-engine`](config/flags.md#-default-engine).

!!! warning
    If you're running more than one instance of `otfd`, take care to set this flag to the same value on each instance. Doing otherwise will lead to unpredictable results.

When you create a workspace, it'll use the default engine. You can override the engine for a workspace in its settings.

When you create a run OTF will download the workspace's engine if it hasn't already been downloaded. The engine binaries are downloaded to the directory specified by the flag [`--engine-bins-dir`](config/flags.md#-engine-bins-dir).
