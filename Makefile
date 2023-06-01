VERSION = $(shell git describe --tags --dirty --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
RANDOM_SUFFIX := $(shell cat /dev/urandom | tr -dc 'a-z0-9' | head -c5)
IMAGE_NAME = leg100/otfd
IMAGE_TAG ?= $(VERSION)-$(RANDOM_SUFFIX)
GOOSE_DBSTRING=postgres:///otf
LD_FLAGS = " \
    -s -w \
	-X 'github.com/leg100/otf/internal.Version=$(VERSION)' \
	-X 'github.com/leg100/otf/internal.Commit=$(GIT_COMMIT)'	\
	-X 'github.com/leg100/otf/internal.Built=$(shell date +%s)'	\
	" \

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# go-tfe-tests runs API tests - before it does that, it builds the otfd docker
# image and starts up otfd and postgres using docker compose, and then the
# tests are run against it.
#
# NOTE: two batches of tests are run:
# (1) using the forked repo
# (2) using the upstream repo, for tests against new features, like workspace tags
.PHONY: go-tfe-tests
go-tfe-tests: image compose-up go-tfe-tests-forked go-tfe-tests-upstream

.PHONY: go-tfe-tests-forked
go-tfe-tests-forked:
	./hack/go-tfe-tests.bash

.PHONY: go-tfe-tests-upstream
go-tfe-tests-upstream:
	./hack/go-tfe-tests-upstream.bash

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	CGO_ENABLED=0 go build -o _build/ -ldflags $(LD_FLAGS) ./...
	chmod -R +x _build/*

.PHONY: install
install:
	go install -ldflags $(LD_FLAGS) ./...

.PHONY: install-latest-release
install-latest-release:
	{ \
	set -ex ;\
	ZIP_FILE=$$(mktemp) ;\
	RELEASE_URL=$$(curl -s https://api.github.com/repos/leg100/otf/releases/latest | \
		jq -r '.assets[] | select(.name | test("otfd_.*_linux_amd64.zip$$")) | .browser_download_url') ;\
	curl -Lo $$ZIP_FILE $$RELEASE_URL ;\
	unzip -o -d $(GOBIN) $$ZIP_FILE otfd ;\
	}

# Run docker compose stack
.PHONY: compose-up
compose-up: image
	docker compose up -d

# Remove docker compose stack
.PHONY: compose-rm
compose-rm:
	docker compose rm -sf

# Run postgresql via docker compose
.PHONY: postgres
postgres:
	docker compose up -d postgres

# Run squid via docker compose
.PHONY: squid
squid:
	docker compose up -d squid

# Run staticcheck metalinter recursively against code
.PHONY: lint
lint:
	go list ./... | grep -v pggen | xargs staticcheck

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# Build docker image
.PHONY: image
image: build
	docker build -f Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) -t $(IMAGE_NAME):latest ./_build

# Build and load image into k8s kind
.PHONY: load
load: image
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG)

# Install pre-commit
.PHONY: install-pre-commit
install-pre-commit:
	pip install pre-commit==3.2.2
	pre-commit install

# Install sql code generator
.PHONY: install-pggen
install-pggen:
	@sh -c "which pggen > /dev/null || go install github.com/leg100/pggen/cmd/pggen@latest"

# Generate sql code
.PHONY: sql
sql: install-pggen
	pggen gen go \
		--postgres-connection "dbname=otf" \
		--query-glob 'internal/sql/queries/*.sql' \
		--output-dir ./internal/sql/pggen \
		--go-type 'text=github.com/jackc/pgtype.Text' \
		--go-type 'int4=int' \
		--go-type 'int8=int' \
		--go-type 'bool=bool' \
		--go-type 'bytea=[]byte' \
		--acronym url \
		--acronym sha \
		--acronym json \
		--acronym vcs \
		--acronym http \
		--acronym tls \
		--acronym hcl
	goimports -w ./internal/sql/pggen
	go fmt ./internal/sql/pggen

# Install DB migration tool
.PHONY: install-goose
install-goose:
	@sh -c "which goose > /dev/null || go install github.com/pressly/goose/v3/cmd/goose@latest"

# Migrate SQL schema to latest version
.PHONY: migrate
migrate: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations up

# Redo SQL schema migration
.PHONY: migrate-redo
migrate-redo: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations redo

# Rollback SQL schema by one version
.PHONY: migrate-rollback
migrate-rollback: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations down

# Get SQL schema migration status
.PHONY: migrate-status
migrate-status: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./internal/sql/migrations status

# Run docs server with live reload
.PHONY: serve-docs
serve-docs:
	mkdocs serve -a localhost:9999

.PHONY: doc-screenshots
doc-screenshots: # update documentation screenshots
	OTF_DOC_SCREENSHOTS=true go test ./internal/integration/... -count 1

# Create tunnel between local server and cloudflare - useful for testing
# webhooks, e.g. a github webhook sending events to local server.
.PHONY: tunnel
tunnel:
	cloudflared tunnel run otf

# Generate path helpers
.PHONY: paths
paths:
	go generate ./internal/http/html/paths
	goimports -w ./internal/http/html/paths

# Re-generate RBAC action strings
.PHONY: actions
actions:
	stringer -type Action ./internal/rbac

# Install staticcheck linter
.PHONY: install-linter
install-linter:
	go install honnef.co/go/tools/cmd/staticcheck@latest
