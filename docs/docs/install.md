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

## Install using docker compose

You can install and run OTF and postgres in a container using [docker compose](https://docs.docker.com/compose/).

First clone the repo:

```
git clone https://github.com/leg100/otf
cd otf
```

Populate an `.env` file with a secret and site token:

```
cat > .env <<EOF
OTF_SECRET=6b07b57377755b07cf61709780ee7484
OTF_SITE_TOKEN=my-site-token
EOF
```

!!! note
    The secret must be a hex-encoded 16-byte array. Generate using `openssl rand -hex 16`.

Then create and start the containers:

```
docker compose up
```

Login to the web app at `http://localhost:8080` and use the site token configured above to login.

!!! warning
    Use at your own risk. This exposes port 8080 on all interfaces, using plaintext HTTP. It also hardcodes the postgres account credentials.

## Install helm chart

You can install OTF onto Kubernetes using helm charts.

Add the helm repository:

```bash
helm repo add otf https://leg100.github.io/otf-charts
```

Then follow instructions for installing the relevant chart:

* [otfd](https://github.com/leg100/otf/blob/master/charts/otfd/README.md)
* [otf-agent](https://github.com/leg100/otf/blob/master/charts/otf-agent/README.md)

## Install using `go`

You'll need [Go](https://golang.org/doc/install). Run:

```
go install github.com/leg100/otf/cmd/otfd@latest
```

That'll install the latest `otfd` binary into your go bin directory (defaults to `$HOME/go/bin`).

See the [quickstart](quickstart.md) for configuring and running `otfd` locally.
