# Helm Chart for `otf-agent`

![Version: 0.1.20](https://img.shields.io/badge/Version-0.1.20-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.5.14](https://img.shields.io/badge/AppVersion-0.5.14-informational?style=flat-square)

Installs the [otf agent](https://docs.otf.ninja/runners/).

## Usage

First, follow the instructions in the [docs](https://docs.otf.ninja/install/#install-from-source) to add the helm repository.

Then ensure:

* You have a running deployment of `otfd`.
* You have generated an agent token (see the [docs](https://docs.otf.ninja/runners/)).

Once you have these, deploy the chart with the URL of the `otfd` deployment and the agent token, e.g.:

```bash
helm install otf-agent otf/otf-agent \
    --set url=https://otf.acme.corp \
    --set token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9wb29sX2lkIjoiYXBvb2wtNWI5MDQ0M2VkODJlZjc2OSIsImlhdCI6MTcyOTUzMDMwMywia2luZCI6ImFnZW50X3Rva2VuIiwic3ViIjoiYXQtWDZLdjVuQVE4Y1NQQ3lvZCJ9._wziD3FlGC2xdF4Ss_sf-igcagrgrhUmM5AFJGrwQso
```

Alternatively you can deploy the token via a secret:

```bash
# create secret
kubectl create secret generic agent-token \
    --from-literal token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZ2VudF9wb29sX2lkIjoiYXBvb2wtNWI5MDQ0M2VkODJlZjc2OSIsImlhdCI6MTcyOTUzMDMwMywia2luZCI6ImFnZW50X3Rva2VuIiwic3ViIjoiYXQtWDZLdjVuQVE4Y1NQQ3lvZCJ9._wziD3FlGC2xdF4Ss_sf-igcagrgrhUmM5AFJGrwQso
# deploy chart
helm template otf-agent otf/otf-agent
    --set url=https://otf.acme.corp
    --set tokenFromSecret.name=agent-token,tokenFromSecret.key=token
```

Check the deploy succeeded:

```bash
kubectl get pod -l app.kubernetes.io/name=otf-agent
NAME                        READY   STATUS    RESTARTS   AGE
otf-agent-b8968d764-v6rbk   1/1     Running   0          3h29m
```

And check the logs to confirm it has registered itself with the server:

```bash
kubectl logs -f otf-agent-b8968d764-v6rbk
2024/10/21 17:06:01 INFO starting agent version=v0.2.4-15-g029e5255-dirty
2024/10/21 17:06:01 INFO registered successfully agent.id=agent-xuY1t2oeMMTjyaWz agent.server=false agent.status=idle agent.ip_
address=10.244.0.18 agent.pool_id=apool-5b90443ed82ef769
2024/10/21 17:06:01 INFO waiting for next job
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| extraEnvs | list | `[]` | Extra environment variables to be passed to the deployment. |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"leg100/otf-agent"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| logging.format | string | `nil` | Logging format: default, text, or json. See [docs](https://docs.otf.ninja/config/flags/#-log-format) |
| logging.verbosity | int | `nil` | Logging verbosity, the higher the number the more verbose the logs. See [docs](https://docs.otf.ninja/config/flags/#-v). |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| rbac.create | bool | `true` |  |
| replicaCount | int | `1` | Number of agents to deploy to cluster. Note: each agent shares the same token and therefore belongs to the same pool. To deploy agents belonging to different pools you'll need to deploy multiple charts. |
| resources | object | `{}` |  |
| securityContext | object | `{}` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| sidecars | list | `[]` | Additional sidecar containers to run alongside the main otf-agent container |
| token | string | `nil` | Token to authenticate the agent. Either this or `tokenFromSecret` must be specified. |
| tokenFromSecret | object | `nil` | Source token from a secret. Either this or `token` must be specified. |
| tolerations | list | `[]` |  |
| topologySpreadConstraints | list | `[]` |  |
| url | string | `nil` | URL of the OTF server to connect to. Must begin with `https://` or `http://`. Required. |
| volumeMounts | list | `[]` | Additional volume mounts for the main otf-agent container |
| volumes | list | `[]` | Additional volumes to make available to the pod |
| runner.cacheVolume.accessModes | list | `["ReadWriteMany"]` | Persistent volume access modes. |
| runner.cacheVolume.enabled | bool | `false` | Enable persistent volume for cache. |
| runner.cacheVolume.size | string | `"100Gi"` | Persistent volume size. |
| runner.cacheVolume.storageClass | string | `nil` | Persistent volume storage class. # If defined, storageClassName: <storageClass> # If set to "-", storageClassName: "", which disables dynamic provisioning # If undefined (the default) or set to null, no storageClassName spec is # set, choosing the default provisioner. |
| runner.concurrency | int | `nil` | Set the number of runs that can be processed concurrently. See [docs](https://docs.otf.ninja/config/flags/#-concurrency). |
| runner.executor | string | `""` | The executor to use. See [docs](https://docs.otf.ninja/config/flags/#-executor) |
| runner.kubernetesTTLAfterFinish | string | `nil` | Delete finished kubernetes jobs after this duration. |
| runner.pluginCache | bool | `nil` | Enable shared provider plugin cache for terraform providers. Note this is only concurrency safe in opentofu 1.10.0 and greater. |

