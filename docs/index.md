# otf

**otf** is an open-source alternative to Terraform Enterprise:

* Full Terraform CLI integration
* Remote execution mode: plans and applies run on servers
* Agent execution mode: plans and applies run on agents
* Remote state backend: state stored in PostgreSQL
* SSO signin: github and gitlab supported
* Team-based authorization: syncs your github teams / gitlab roles
* Compatible with much of the Terraform Enterprise/Cloud API
* Minimal dependencies: requires only PostgreSQL
* Stateless: horizontally scale servers in pods on Kubernetes, etc

!!! Note
    The server and agent have only been tested on linux.

## Quick install

These steps will get you started with running the server on your local system.

Download a [release](https://github.com/leg100/otf/releases) of the server component, `otfd`. The release is a zip file. Extract the `otfd` binary to your current directory.

Ensure you have access to a postgres server. otf by default assumes postgres is running locally, accessible via a domain socket in `/var/run/postgresql`, and defaults to using a database named `otf`. You need to create the database first:

```bash
createdb otf
```

otfd requires a secret for creating cryptographic signatures. It should be up to 64 characters long and you should use a cryptographically secure random number generator, e.g.:

```bash
> openssl rand -hex 32
56789f6076a66323643f57a1016cdde7e7e39914785d36d61fdd8b9a30081f14
```

To get up and running quickly, we'll use the **site admin** account. This account has complete privileges and should only be used for administrative tasks rather than day-to-day usage. To use the account you need to set a token, which can any combination of characters. Make a note of this.

Now start the otf daemon with both the secret and the token:

```bash
> ./otfd --secret=my-secret --token=my-token
2022-10-30T20:06:10Z INF started cache max_size=0 ttl=10m0s
2022-10-30T20:06:10Z INF successfully connected component=database path=postgres:///otf?host=/var/run/postgresql
2022-10-30T20:06:10Z INF goose: no migrations to run. current version: 20221017170815 compone
nt=database
2022-10-30T20:06:10Z INF started server address=[::]:8080 ssl=false
```

You have now successfully installed `otfd` and confirmed you can start `otfd` with minimal configuration. Proceed to create your first organization.

### Create organization

You can navigate to the web app in your browser:

[http://localhost:8080](http://localhost:8080)

Note it announces you have 'no authenticators configured'. The normal method of login is to use SSO signin, via Github etc, but in this quickstart we're using the site admin account. Click on the 'site admin' link in the bottom right, and use your token to login.

Go to 'organizations' and click the button `New Organization`. Give the organization a name and create.

### Run Terraform

!!! Note
    The terraform CLI will be connecting to the server and it expects to make a verified SSL connection. Therefore we need to configure SSL first.

Generate a self-signed SSL certificate and key:

```bash
openssl req -x509 -newkey rsa:4096 -sha256 -keyout key.pem -out cert.crt -days 365 -nodes -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost'
```
    
Ensure your system trusts the generated cert. For example, on Ubuntu based systems:

```bash
sudo cp cert.crt /usr/local/share/ca-certificates
sudo update-ca-certificates
```

Now return to the terminal in which `otfd` is running. You'll need to kill it and start it again, this time with SSL enabled:
    
```bash
> ./otfd --secret=my-secret --token=my-token --ssl --cert-file=cert.crt --key-file=key.pem
```

Terraform needs to use your token to authenticate with `otfd`:

```bash
terraform login localhost:8080
```

Enter `yes` to proceed.

!!! warning
    You'll notice `terraform login` opens a browser window. However it ignores the port, thereby failing to open the correct page on the server. Once you properly deploy the server on a non-custom port this won't be a problem.

Ignore the browser window it has opened and enter your token at the terminal prompt. You should receive confirmation of success:

```
Success! Logged in to Terraform Enterprise (localhost:8080)
```

Now we'll write some terraform configuration. Configure the terraform backend and define a resource:

```bash
cat > main.tf <<EOF
terraform {
  backend "remote" {
    hostname = "localhost:8080"
    organization = "default"

    workspaces {
      name = "dev"
    }
  }
}

resource "null_resource" "quickstart" {}
EOF
```

Initialize terraform:

```bash
terraform init
```

Run a plan:

```bash
terraform plan
```

That starts a run on the server. You can click on the link to the run to view status and logs.

And apply:

```bash
terraform apply
```

This starts another run on the server. Again you can click on the link to see logs.

You have reached the end of this quickstart guide. Have a look at the remainder of the documentation to further complete the installation of otf, to setup SSO, run agents, etc.
