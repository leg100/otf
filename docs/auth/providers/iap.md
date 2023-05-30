# Google IAP

OTF supports deployment using [Google's Identity-Aware Proxy](https://cloud.google.com/iap). Deploy an OTF cluster to Google Cloud (GCP) and enable IAP to authenticate users accessing the cluster. Only authenticated requests reach the cluster and each request contains information about the user. OTF verifies the requests and checks the user exists. If the user does not exist an account is created.

<figure markdown>
![IAP-GKE deployment](/images/iap-load-balancer.png)
    <figcaption>IAP deployment with GCP Compute Engine / GKE (image sourced from [Google Cloud documentation](https://cloud.google.com/iap/docs/concepts-overview))</figcaption>
</figure>

## Verification

OTF checks each incoming request for the presence of a [signed IAP header](https://cloud.google.com/iap/docs/signed-headers-howto). If present then it verifies the header's signed token to verify it originated from Google IAP and that it has not expired.

You can also configure OTF to validate the **audience** token claim. Validating the audience checks OTF is indeed the intended recipient of the request. Follow [Google's instructions](https://cloud.google.com/iap/docs/signed-headers-howto#iap_validate_jwt-go) for retrieving the audience string. Then set the [--iap-google-jwt-audience](../../../config/flags/#-google-jwt-audience) `otfd` flag accordingly, e.g.:

```
otfd --google-jwt-audience /projects/project_number/apps/my_project_id
```

It is recommended you set this flag, especially for a production deployment.

## Authentication

Authentication is delegated to IAP. From the [Google Cloud documentation](https://cloud.google.com/iap/docs/concepts-overview):

> ...IAP checks the user's browser credentials. If none exist, the user is redirected to an OAuth 2.0 Google Account sign-in flow that stores a token in a browser cookie for future sign-ins... ...If the request credentials are valid, the authentication server uses those credentials to get the user's identity (email address and user ID). The authentication server then uses the identity to check the user's IAM role and check if the user is authorized to access the resource.

## Authorization

IAP permits restricting _which_ users can access the cluster (ibid):

> After authentication, IAP applies the relevant IAM policy to check if the user is authorized to access the requested resource. If the user has the IAP-secured Web App User role on the Google Cloud console project where the resource exists, they're authorized to access the application

Whereas OTF remains responsible for determining _what_ users can access, i.e. you assign users to teams and set team permissions to allow access to organizations and workspaces, etc.
