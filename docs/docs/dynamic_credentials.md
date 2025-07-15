# Dynamic Provider Credentials

OTF supports dynamic provider credentials, avoiding the need for static credentials. Furthermore they let you use your cloud provider's auth tools to scope permissions based on the properties of your run, including the organizatino, workspace and run phase.

## How they work

Essentially they rely on first establishing a trust relationship between OTF and your cloud provider. You can define IAM policies to specify organizations or workspaces that are allowed to access specific cloud resources.

Once that initial setup is complete, for each plan or apply the following steps are carried out:

1. OTF generates a token and configures your cloud provider to send that token when it authenticates.
2. The cloud provider receives the token and sends a request to OTF to retrieve the public key to verify the token.
3. The cloud provider upon success sends temporary credentials to OTF.
4. Those credentials are then used as the plan or apply proceeds and makes API calls to your cloud provider.
5. The credentials are discarded upon completion.

## Setup

The setup steps are not simple and differ according to cloud provider. OTF fully\* implements Terraform Cloud's dynamic provider credentials implementation, using the exact same environment variables. You can therefore rely entirely on [its documentation](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials) and follow its instructions to setup dynamic provider credentials for your cloud provider.

Thus far the following cloud providers are supported in OTF:

* GCP
* AWS
* Azure

There are a few pre-requisites specific to OTF you need to first carry out:

1. Generate a public key pair:

```
openssl genrsa -out key.pem 4096
openssl rsa -in key.pem -pubout -out public.pem
```

2. Configure `otfd` with public key pair:

```
otfd --public-key-path public.pem --private-key-path key.pem
```

3. Ensure external access to metadata endpoints:

In order to verify signed JWTs, cloud platforms must have network access to the following static OIDC metadata endpoints within OTF:

`/.well-known/openid-configuration` - standard OIDC metadata.
`/.well-known/jwks` - OTF`s public key(s) that cloud platforms use to verify the authenticity of tokens that claim to come from OTF.

!!! note

Not all cloud providers have this requirement. For example, GCP permits uploading the key:

OIDC JWKS files that are directly uploaded to Google Cloud. By using this method, the endpoint doesn't need to be publicly accessible. The x5c and x5t fields inside the JWK aren't supported and must be removed before uploading

## Notes

\* There are some minor differences where OTF divergges from the Terraform Cloud documentation:

* The token's subject in TFC has the format: `organization:<org>:project:<project>:workspace:<workspace>:run-phase:<phase>` Whereas OTF does not have support for projects. Therefore the subject format is `organization:<org>:workspace:<workspace>:run-phase:<phase>`.
* The example terraform configurations reference projects, but OTF does not support projects. Therefore you'll need to amend the terraform configurations accordingly.
