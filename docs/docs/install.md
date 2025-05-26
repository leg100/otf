# Installation

## Requirements

* Linux - the server and agent components are tested on Linux only; the client CLI should work on all platforms.
* PostgreSQL - at least version 12.
* Terraform >= 1.2.0
* An SSL certificate.

## Download

There are three components that can be downloaded:

* `otfd` - the server daemon
* `otf` - the client CLI
* `otf-agent` - the agent daemon

Download them from [Github releases](https://github.com/leg100/otf/releases).

The server and agent components are also available as docker images:

* `leg100/otfd`
* `leg100/otf-agent`

## Install helm chart

You can install OTF onto Kubernetes using helm charts.

Add the helm repository:

```bash
helm repo add otf https://leg100.github.io/otf-charts
```

Then follow instructions for installing the relevant chart:

* [otfd](https://github.com/leg100/otf/blob/master/charts/otfd/README.md)
* [otf-agent](https://github.com/leg100/otf/blob/master/charts/otf-agent/README.md)

## Install from source

You'll need [Go](https://golang.org/doc/install).

Clone the repo, then build and install using the make task:

```bash
git clone https://github.com/leg100/otf
cd otf
make install
```

That'll install the binaries inside your go bin directory (defaults to `$HOME/go/bin`).
