FROM alpine:3.14

# otfd binary is expected to be copied from the PWD because that is where goreleaser builds it and
# there does not appear to be a way to customise a different location
COPY otfd /usr/local/bin/otfd

ENTRYPOINT ["/usr/local/bin/otfd"]
