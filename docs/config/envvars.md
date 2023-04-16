# Environment variables

OTF can be configured from environment variables. Arguments can be converted to the equivalent env var by prefixing
it with `OTF_`, replacing all `-` with `_`, and upper-casing it. For example:

- `--secret` becomes `OTF_SECRET`
- `--site-token` becomes `OTF_SITE_TOKEN`

Env variables can be suffixed with `_FILE` to tell OTF to read the values from a file. This is useful for container
environments where secrets are often mounted as files.
