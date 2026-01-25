# Executors

An *executor* is responsible for executing jobs, i.e. plans and applies. There are two types of executor:

* `fork`
* `kubernetes`

The executor is set with the [`--executor`](config/flags.md#-executor) flag. The flag is applicable to both `otfd` and `otf-agent`, which determines how jobs scheduled to that process are executed. For example, you could set `otfd` to use `fork` but if a job is scheduled to run on an `otf-agent` then the job will use whatever executor it has set.


## Fork

Fork is the default executor. It [forks](https://en.wikipedia.org/wiki/Fork_(system_call)) terraform (or tofu) processes as children of the `otfd` or `otf-agent` process.

The maximum number of forked processes is set with the [--concurrency](config/flags.md#-concurrency) flag.

!!! note
    If you want to scale processes beyond a single host with the `fork` executor then you can either run `otfd` on other hosts to form an OTF cluster, or run `otf-agent` on other hosts. See [runners](runners.md) for more details on agents.

## Kubernetes

The kubernetes executor executes jobs on kubernetes. Each job is run as a [kubernetes job](https://kubernetes.io/docs/concepts/workloads/controllers/job/).

This executor is only functional when `otfd` or `otf-agent` is deployed via the [helm charts](https://github.com/leg100/otf/tree/master/charts) to a kubernetes cluster.

There are a number of flags that customise the jobs:

* [`--kubernetes-request-cpu`](config/flags.md#-kubernetes-request-cpu)
* [`--kubernetes-request-memory`](config/flags.md#-kubernetes-request-memory)
* [`--kubernetes-ttl-after-finish`](config/flags.md#-kubernetes-ttl-after-finish)

It's advisable to provide a persistent volume for the cache. Otherwise the terraform or tofu binary is downloaded at the beginning of every job. See the helm chart settings to enable the persistent volume claim. You will need to make available a persistent volume that supports the [ReadWriteMany](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes) access mode.

