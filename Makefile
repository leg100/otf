VERSION = $(shell git describe --tags --dirty --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
RANDOM_SUFFIX := $(shell cat /dev/urandom | tr -dc 'a-z0-9' | head -c5)
IMAGE_NAME = leg100/otfd
IMAGE_NAME_AGENT = leg100/otf-agent
IMAGE_TAG ?= $(VERSION)-$(RANDOM_SUFFIX)
DBSTRING=postgres:///otf
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

.PHONY: go-tfe-tests
go-tfe-tests: image compose-up
	./hack/go-tfe-tests.bash

.PHONY: watch
watch: tailwind-watch modd

.PHONY: modd
modd:
	+modd

.PHONY: tailwind
tailwind:
	npx tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css

.PHONY: tailwind-watch
tailwind-watch:
	+npx tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css --watch

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
	docker compose up -d --wait --wait-timeout 60

# Remove docker compose stack
.PHONY: compose-rm
compose-rm:
	docker compose rm -sf

# Run postgresql via docker compose
.PHONY: postgres
postgres:
	docker compose up -d postgres

# Run staticcheck metalinter recursively against code
.PHONY: lint
lint:
	go list ./... | grep -v github.com/leg100/otf/internal/sql/sqlc | xargs staticcheck

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

# Build docker image for otf-agent
.PHONY: image-agent
image-agent: build
	docker build -f ./Dockerfile.agent -t $(IMAGE_NAME_AGENT):$(IMAGE_TAG) -t $(IMAGE_NAME_AGENT):latest ./_build

# Build and load otf-agent image into k8s kind
.PHONY: load-agent
load-agent: image-agent
	kind load docker-image $(IMAGE_NAME_AGENT):$(IMAGE_TAG)

# Install pre-commit
.PHONY: install-pre-commit
install-pre-commit:
	pip install pre-commit==3.2.2
	pre-commit install

# Install sql code generator
.PHONY: install-sqlc
install-sqlc:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate sql code and register table types with pgx
.PHONY: sql
sql:
	sqlc generate && go generate ./internal/sql

# Install DB migration tool
.PHONY: install-migrator
install-migrator:
	go install github.com/jackc/tern@latest

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
