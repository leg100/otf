# This is a multi-stage build.
#
# Use --target to choose which image to build:
# docker build --tag otfd --target otfd .
# docker build --tag otf-agent --target otf-agent .
# docker build --tag otf-job --target otf-job .

# STAGE: base
# This stage contains the files/packages that are used by the final images.
FROM --platform=$BUILDPLATFORM alpine:3.23.3 AS base

# Build cache optimization: use cache mounts
# https://docs.docker.com/build/cache/optimize/#use-cache-mounts Mounting
# /etc/apk/cache with type=cache lets the image build system handle apk's cache
# for us, without leaving any cache files in the image itself.

# To Do: 
#   - Update base image after CVEs are fixed
#   - Remove apk update && apk upgrade (When the CVEs are fixed, this is no longer relevant)
RUN --mount=type=cache,target=/etc/apk/cache \
  apk update && \
  apk upgrade --no-cache && \
  apk add --no-cache --upgrade git gcompat

# Create non-root user and group, which will be used in the final image. We
# can't switch to it now, because the change only lasts until the end of the
# stage.
RUN adduser -D -u 4096 otf

ARG TARGETPLATFORM

# STAGE: otfd
# Final stage that takes the `base` stage and the `otfd` binary
FROM base AS otfd
COPY $TARGETPLATFORM/otfd /usr/local/bin/
USER otf
ENTRYPOINT ["/usr/local/bin/otfd"]

# STAGE: otf-agent
# Final stage that takes the `base` stage and the `otf-agent` binary
FROM base AS otf-agent
COPY $TARGETPLATFORM/otf-agent /usr/local/bin/
USER otf
ENTRYPOINT ["/usr/local/bin/otf-agent"]

# STAGE: otf-job
# Final stage that takes the `base` stage and the `otf-job` binary
FROM base AS otf-job
COPY $TARGETPLATFORM/otf-job /usr/local/bin/
USER otf
ENTRYPOINT ["/usr/local/bin/otf-job"]
