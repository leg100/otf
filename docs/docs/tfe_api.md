# TFE API Compatiblity

OTF implements much of the [Terraform Enterprise API](https://developer.hashicorp.com/terraform/enterprise/api-docs). That means it also supports the use of the [TFE terraform provider](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs?product_intent=terraform) to manage resources on OTF.

As such you can rely on the Hashicorp documentation for using the API and the provider with OTF. However, OTF doesn't implement all of the API endpoints, and for those endpoints it does support there can be differences in the request and response payloads. This document serves to record those differences where they exist.

## OAuth Clients

https://developer.hashicorp.com/terraform/enterprise/api-docs/oauth-clients

OAuth Clients map to VCS providers in OTF.

The create endpoint requires two parameters: `http-url` and `api-url`. OTF does too. However it only uses the `api-url` parameter for a VCS provider. OTF treats this as the base URL for the provider. e.g. for the public Github API, it would be `https://api.github.com`.

