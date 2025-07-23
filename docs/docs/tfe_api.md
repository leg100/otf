# TFE API Compatiblity

OTF implements much of the [Terraform Enterprise API](https://developer.hashicorp.com/terraform/enterprise/api-docs). That means it also supports the use of the [TFE terraform provider](https://registry.terraform.io/providers/hashicorp/tfe/latest/docs?product_intent=terraform) to manage resources on OTF.

As such you can rely on the Hashicorp documentation for using the API and the provider with OTF. However, OTF doesn't implement all of the API endpoints, and for those endpoints it does support there can be differences in the request and response payloads. This document serves to record those differences where they exist.

## OAuth Clients [[Docs]](https://developer.hashicorp.com/terraform/enterprise/api-docs/oauth-clients)

(The OAuth Clients API essentially is an API for what are called *VCS providers* in both TFE and OTF).

The create endpoint requires two parameters: `http-url` and `api-url`. OTF does too. However it only uses the `http-url` parameter for a VCS provider. This URL is the homepage or base URL of your VCS provider. OTF doesn't use the `api-url` parameter (which is documented as the base URL of the API of the provider) and instead it automatically infers it from the `http-url` parameter. .e.g for the public Github provider the `http-url` would be `https://github.com` and from this OTF automatically sets the API URL to be `https://api.github.com`.


The create endpoint also takes a `service-provider` parameter. This can be set to any number of values such as `gitlab_community_edition` or `gitlab_enterprise_edition`. From this value OTF works out which "kind" of VCS provider to create (just as TFE does). However OTF makes no distinction between "sub-kinds": e.g. `gitlab_community_edition` or `gitlab_enterprise_edition` are both deemed by OTF to be the `github` kind and makes no further distinction.

!!! note

    Github *App* installations are treatly differently in OTF and TFE. In TFE they're handled separately from "VCS providers": you create an installation at the admin level and then you're free to use that installation across organizations (I mgight be wrong on the exact details here...). Whereas in OTF they're treated as a VCS provider. That means you should be able to create a VCS provider for a Github App installation via the OAuth Clients create endpoint but that isn't yet currently implemented. (Raise an issue if you would like this).
