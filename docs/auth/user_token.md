# User Tokens

A user can generate API tokens. The token shares the same permissions as the user.

To manage your tokens, go to **Profile > Tokens**.

API tokens are not only used for programmatic access but for authenticating `terraform` and `otf`. For example, you can use `terraform login` to store a token on your workstation:

```bash
terraform login <otf hostname>
```

And follow the instructions. The token is persisted to a local credentials file for use by both `terraform` and `otf`.
