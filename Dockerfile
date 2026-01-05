# This is a multi-stage build with layers optimized for efficient caching and
# to minimize the final image sizes.
#
# Use --target to choose which image to build:
# docker build --tag otfd --target otfd .
# docker build --tag otf-agent --target otf-agent .
# docker build --tag otf-job --target otf-job .

# STAGE: builder Used for building the binaries. This stage/layer will not be
# included in the final images, as only the compiled binaries are copied from
# it.
FROM golang:alpine3.23 AS builder

WORKDIR /app

# Build cache optimization: use cache mounts
# https://docs.docker.com/build/cache/optimize/#use-cache-mounts Mounting
# /etc/apk/cache with type=cache lets the image build system handle apk's cache
# for us, without leaving any cache files in the image itself.
RUN --mount=type=cache,target=/etc/apk/cache \
  apk add make git

# Unused files/folders are filtered out with the .dockerignore file.
COPY . .

# Compile the binaries
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  make build

# STAGE: base
# This stage contains the files/packages that are used by the final images.
FROM alpine:3.23 AS base

# Create non-root user and group, which will be used in the final image. We
# can't switch to it now, because the change only lasts until the end of the
# stage.
RUN adduser -D -H -u 4096 otf

# git permits pulling modules via the git protocol
RUN --mount=type=cache,target=/etc/apk/cache \
  apk add git

# STAGE: otfd
# Final stage that takes the `base` stage and the `otfd` binary
FROM base AS otfd
COPY --from=builder /app/_build/otfd /usr/local/bin/
USER otf
ENTRYPOINT ["/usr/local/bin/otfd"]


# STAGE: otf-agent
# Final stage that takes the `base` stage and the `otf-agent` binary
FROM base AS otf-agent
COPY --from=builder /app/_build/otf-agent /usr/local/bin/
USER otf
ENTRYPOINT ["/usr/local/bin/otf-agent"]

# STAGE: otf-job
# Final stage that takes the `base` stage and the `otf-job` binary
FROM base AS otf-job
COPY --from=builder /app/_build/otf-job /usr/local/bin/
USER otf
ENTRYPOINT ["/usr/local/bin/otf-job"]
