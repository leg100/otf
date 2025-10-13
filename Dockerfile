# This is a multi-stage build with layers optimized for efficient caching and to minimize the final image sizes.
# 
# Use --target to choose which image to build:
# docker build --tag otfd --target otfd .
# docker build --tag otf-agent --target otf-agent .


# STAGE: builder
# Used for building the `otfd` and `otf-agent` binaries. This stage/layer will not be included in the final images, as
# only the compiled binaries are copied from it.
FROM golang:1.25.0-alpine3.22 AS builder

WORKDIR /app

# Build cache optimization: use cache mounts
# https://docs.docker.com/build/cache/optimize/#use-cache-mounts
# Mounting /etc/apk/cache with type=cache lets the image build system handle apk's cache for us,
# without leaving any cache files in the image itself.
RUN --mount=type=cache,target=/etc/apk/cache \
  apk add make=~4.4.1 git=~2.49.1

# Unused files/folders are filtered out with the .dockerignore file.
COPY . .

# Compile the binaries
# Chmod 0555 ensures that the binaries don't have potentially dangerous permissions, like setuid and world writable.
RUN --mount=type=cache,target=/go/pkg/mod \
  make build && chmod 0555 _build/*


# STAGE: base
# This stage contains the files/packages that are used by the final images.
FROM alpine:3.22 AS base

# Create non-root user and group, which will be used in the final image.
# We can't switch to it now, because the change only lasts until the end of the stage.
RUN adduser -D -H -u 4096 otf

# git permits pulling modules via the git protocol
# tini manages init and signal handling
RUN --mount=type=cache,target=/etc/apk/cache \
  apk add git=~2.49.1 tini=~0.19.0


# STAGE: otfd
# Final stage that takes the `base` stage and the `otfd` binary
FROM base AS otfd
COPY --from=builder /app/_build/otfd /usr/local/bin/
USER otf
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/otfd"]


# STAGE: otf-agent
# Final stage that takes the `base` stage and the `otf-agent` binary
FROM base AS otf-agent
COPY --from=builder /app/_build/otf-agent /usr/local/bin/
USER otf
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/otf-agent"]
