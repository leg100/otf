FROM alpine:3.14

# Terraform version must be kept in sync with otf.DefaultTerraformVersion
ARG TERRAFORM_VERSION=1.0.10

# Install terraform
RUN apk add curl && \
    curl -LOs https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    curl -LOs https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_SHA256SUMS && \
    sed -n "/terraform_${TERRAFORM_VERSION}_linux_amd64.zip/p" terraform_${TERRAFORM_VERSION}_SHA256SUMS | sha256sum -c && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/local/bin && \
    rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    rm terraform_${TERRAFORM_VERSION}_SHA256SUMS

# otfd binary is expected to be copied from the PWD because that is where goreleaser builds it and
# there does not appear to be a way to customise a different location
COPY otfd /usr/local/bin/otfd

ENTRYPOINT ["/usr/local/bin/otfd"]
