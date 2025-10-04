FROM alpine:3.17

# git permits pulling modules via the git protocol
RUN apk add --no-cache git

COPY otfd /usr/local/bin/otfd

ENTRYPOINT ["/usr/local/bin/otfd"]
