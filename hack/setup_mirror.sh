# Setup filesystem mirror for both terraform and tofu. This speeds up the
# integration tests massively.
cd mirror/
terraform providers mirror .
tofu providers mirror .
