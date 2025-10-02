FROM golang:1.25.0-alpine3.22 AS builder

WORKDIR /app

# build cache optimization, use cache mounts
# https://docs.docker.com/build/cache/optimize/#use-cache-mounts
# hadolint ignore=DL3019
RUN --mount=type=cache,target=/etc/apk/cache \
  apk add make=~4.4.1 git=~2.49.1

# Unused files/folders are filtered out with the .dockerignore file.
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
  make build


FROM alpine:3.22 AS base
# Create non-root user and group, which will be used in the final image.
# We can't switch to it now, because the change only lasts until the end of the stage.
RUN adduser -D -H -u 4096 otf

# bubblewrap is for sandboxing, and git permits pulling modules via
# the git protocol
# tini manages init and signal handling
# hadolint ignore=DL3019
RUN --mount=type=cache,target=/etc/apk/cache \
  apk add bubblewrap=~0.11.0 git=~2.49.1 tini=~0.19.0

# Two minimal stages allow us to easily build separate images that have just the required binary.
#
# Use --target to choose which image to build:
# docker build --tag otfd --target otfd .
# docker build --tag otf-agent --target otf-agent .
#
# With `COPY --chmod=0555` we ensure that the binary doesn't have something dangerous, like setuid and world writable.
FROM base AS otfd
COPY --chmod=0555 --from=builder /app/_build/otfd /usr/local/bin/
USER otf
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/otfd"]

FROM base AS otf-agent
COPY --chmod=0555 --from=builder /app/_build/otf-agent /usr/local/bin/
USER otf
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/otf-agent"]
