# Dynamic Provider Credentials

OTF supports dynamic provider credentials, avoiding the need for static credentials. Furthermore they let you use your cloud provider's auth tools to scope permissions based on the properties of your run, including the organization, workspace and run phase.

## How they work

Essentially they rely on first establishing a trust relationship between OTF and your cloud provider. You can define IAM policies to specify organizations or workspaces that are allowed to access specific cloud resources.

Once that initial setup is complete, for each plan or apply the following steps are carried out:

1. OTF generates a token and configures your cloud provider to send that token when it authenticates.
2. The cloud provider receives the token and sends a request to OTF to retrieve the public key to verify the token.
3. The cloud provider upon success sends temporary credentials to OTF.
4. Those credentials are then used as the plan or apply proceeds and makes API calls to your cloud provider.
5. The credentials are discarded upon completion.

## Setup

The setup steps are not simple and differ according to cloud provider. OTF implements Terraform Cloud's [dynamic provider credentials implementation](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials), using the exact same environment variables, allowing for [configuration of multiple provider blocks](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/specifying-multiple-configurations). You can therefore rely entirely on [its documentation](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials) and follow its instructions to setup dynamic provider credentials for your cloud provider.

Thus far the following cloud providers are supported in OTF:

* GCP
* AWS
* Azure

!!! warning
    Only GCP support has been fully tested by the developers. Please report success or bugs with the AWS and Azure providers on Github Issues / Slack.

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

    * `/.well-known/openid-configuration` - standard OIDC metadata.
    * `/.well-known/jwks` - OTF`s public key(s) that cloud platforms use to verify the authenticity of tokens that claim to come from OTF.

    !!! note
        Not all cloud providers have this requirement. For example, GCP permits [uploading the key](https://cloud.google.com/iam/docs/workload-identity-federation-with-other-providers#manage-oidc-keys). If you want to do this, the key can be retrieved from the `./.well-known/jwks` path above, e.g.:

            curl https://localhost:8080/.well-known/jwks -o key-to-upload.json

        The benefit of this approach is that you don't need to expose the endpoints above publicly.

4. When following the provider specific documentation, you'll be prompted to enter an issuer, which is the URL of your OTF installation. The URL's hostname must match the value of the `--hostname` flag.

### Differences

There are some minor differences where OTF diverges from the Terraform Cloud documentation, mainly around Terraform Cloud projects, which OTF does not support:

* The token's subject in TFC has the format: `organization:<org>:project:<project>:workspace:<workspace>:run-phase:<phase>` Whereas OTF does not have support for projects. Therefore the subject format is `organization:<org>:workspace:<workspace>:run-phase:<phase>`.
* Their example terraform configurations reference projects, but OTF does not support projects. Therefore you'll need to amend the terraform configurations accordingly if you decide to use them.

