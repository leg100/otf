FROM alpine:3.14

# for sandboxing
RUN apk add bubblewrap

COPY otfd /usr/local/bin/otfd

ENTRYPOINT ["/usr/local/bin/otfd"]
