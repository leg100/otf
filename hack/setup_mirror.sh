cd mirror/
terraform providers mirror .
tofu providers mirror .

cat > mirror.tfrc <<-EOF
provider_installation {
  filesystem_mirror {
    path = "${PWD}"
  }
}
EOF

echo "now set TF_CLI_CONFIG_FILE=${PWD}/mirror.tfrc"
