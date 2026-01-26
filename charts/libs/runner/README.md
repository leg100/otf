# runner

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: library](https://img.shields.io/badge/Type-library-informational?style=flat-square) ![AppVersion: 1.16.0](https://img.shields.io/badge/AppVersion-1.16.0-informational?style=flat-square)

A Helm chart library for otf runner resources and configuration.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cacheVolume.accessModes | list | `["ReadWriteMany"]` | Persistent volume access modes. |
| cacheVolume.enabled | bool | `false` | Enable persistent volume for cache. |
| cacheVolume.size | string | `"100Gi"` | Persistent volume size. |
| cacheVolume.storageClass | string | `nil` | Persistent volume storage class. # If defined, storageClassName: <storageClass> # If set to "-", storageClassName: "", which disables dynamic provisioning # If undefined (the default) or set to null, no storageClassName spec is # set, choosing the default provisioner. |
| concurrency | int | `nil` | Set the number of runs that can be processed concurrently. See [docs](https://docs.otf.ninja/config/flags/#-concurrency). |
| executor | string | `""` | The executor to use. See [docs](https://docs.otf.ninja/config/flags/#-executor) |
| kubernetesTTLAfterFinish | string | `nil` | Delete finished kubernetes jobs after this duration. |
| pluginCache | bool | `nil` | Enable shared provider plugin cache for terraform providers. Note this is only concurrency safe in opentofu 1.10.0 and greater. |

