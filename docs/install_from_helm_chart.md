# Install from helm chart

You can install the otf server on Kubernetes using the helm chart.

## Instructions

```bash
helm repo add otf https://leg100.github.io/otf-charts
helm upgrade --install otf otf/otf
```

To see all configurable options with detailed comments:

```
helm show values otf/otf
```

!!! note
    The helm chart is maintained in a separate [github repo](https://github.com/leg100/otf-charts).

