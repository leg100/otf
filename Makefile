VERSION = $(shell git describe --tags --dirty --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
RANDOM_SUFFIX := $(shell cat /dev/urandom | tr -dc 'a-z0-9' | head -c5)
IMAGE_NAME = leg100/otfd
IMAGE_TAG ?= $(VERSION)-$(RANDOM_SUFFIX)
GOOSE_DBSTRING=postgres:///otf
LD_FLAGS = " \
    -s -w \
	-X 'github.com/leg100/otf.Version=$(VERSION)' \
	-X 'github.com/leg100/otf.Commit=$(GIT_COMMIT)'	\
	-X 'github.com/leg100/otf.Built=$(shell date +%s)'	\
	" \

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: go-tfe-tests
go-tfe-tests: build
	./hack/go-tfe-tests.bash

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
	ZIP_FILE=$$(tempfile --prefix=otf --suffix=.zip) ;\
	RELEASE_URL=$$(curl -s https://api.github.com/repos/leg100/otf/releases/latest | \
		jq -r '.assets[] | select(.name | test("otfd_.*_linux_amd64.zip$$")) | .browser_download_url') ;\
	curl -Lo $$ZIP_FILE $$RELEASE_URL ;\
	unzip -o -d $(GOBIN) $$ZIP_FILE otfd ;\
	}

# Run postgresql in a container
.PHONY: postgres
postgres:
	docker compose -f docker-compose-postgres.yml up -d

# Stop and remove postgres container
.PHONY: postgres-rm
postgres-rm:
	docker compose -f docker-compose-postgres.yml rm -sf

# Run squid caching proxy in a container
.PHONY: squid
squid:
	docker run --rm --name squid -t -d -p 3128:3128 -v $(PWD)/integration/fixtures:/etc/squid/certs leg100/squid:0.2

# Stop squid container
.PHONY: squid-stop
squid-stop:
	docker stop --signal INT squid

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
		--query-glob 'sql/queries/*.sql' \
		--output-dir sql/pggen \
		--go-type 'text=github.com/jackc/pgtype.Text' \
		--go-type 'int4=int' \
		--go-type 'bool=bool' \
		--go-type 'bytea=[]byte' \
		--acronym url \
		--acronym sha \
		--acronym json \
		--acronym vcs \
		--acronym http \
		--acronym tls \
		--acronym hcl
	goimports -w ./sql/pggen
	go fmt ./sql/pggen

# Install DB migration tool
.PHONY: install-goose
install-goose:
	@sh -c "which goose > /dev/null || go install github.com/pressly/goose/v3/cmd/goose@latest"

# Migrate SQL schema to latest version
.PHONY: migrate
migrate: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./sql/migrations up

# Redo SQL schema migration
.PHONY: migrate-redo
migrate-redo: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./sql/migrations redo

# Rollback SQL schema by one version
.PHONY: migrate-rollback
migrate-rollback: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./sql/migrations down

# Get SQL schema migration status
.PHONY: migrate-status
migrate-status: install-goose
	GOOSE_DBSTRING=$(GOOSE_DBSTRING) GOOSE_DRIVER=postgres goose -dir ./sql/migrations status

# Run docs server with live reload
.PHONY: serve-docs
serve-docs:
	mkdocs serve -a localhost:9999

# Create tunnel between local server and cloudflare - useful for testing
# webhooks, e.g. a github webhook sending events to local server.
.PHONY: tunnel
tunnel:
	cloudflared tunnel run otf

# Generate path helpers
.PHONY: paths
paths:
	go generate ./http/html/paths
	goimports -w ./http/html/paths

# Re-generate RBAC action strings
.PHONY: actions
actions:
	stringer -type Action ./rbac

# Install staticcheck linter
.PHONY: install-linter
install-linter:
	go install honnef.co/go/tools/cmd/staticcheck@latest
