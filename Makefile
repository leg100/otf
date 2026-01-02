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
	docker compose -f docker-compose.testing.yml up -d --wait --wait-timeout 60

# Remove docker compose stack
.PHONY: compose-rm
compose-rm:
	docker compose -f docker-compose.testing.yml rm -sf

# Run postgresql via docker compose
.PHONY: postgres
postgres:
	docker compose -f docker-compose.testing.yml up -d postgres

# Install staticcheck linter
.PHONY: install-linter
install-linter:
	go get -tool honnef.co/go/tools/cmd/staticcheck@2025.1.1

# Run staticcheck metalinter recursively against code
.PHONY: lint
lint: install-linter
	go tool staticcheck ./...

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
image:
	docker build -f Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) -t $(IMAGE_NAME):latest --target otfd .

# Build and load image into k8s kind
.PHONY: load
load: image
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG)

# Build docker image for otf-agent
.PHONY: image-agent
image-agent:
	docker build -f Dockerfile -t $(IMAGE_NAME_AGENT):$(IMAGE_TAG) -t $(IMAGE_NAME_AGENT):latest --target otf-agent .

# Build and load otf-agent image into k8s kind
.PHONY: load-agent
load-agent: image-agent
	kind load docker-image $(IMAGE_NAME_AGENT):$(IMAGE_TAG)

# Install pre-commit
.PHONY: install-pre-commit
install-pre-commit:
	pip install pre-commit==3.2.2
	pre-commit install

# Install DB migration tool
.PHONY: install-migrator
install-migrator:
	go install github.com/jackc/tern@latest

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
	go tool goimports -w ./internal/http/html/paths
	go tool goimports -w ./internal/http/html/components/paths

# Re-generate RBAC action strings
.PHONY: actions
actions:
	go tool stringer -type Action ./internal/authz

.PHONY: debug
debug:
	dlv debug --headless --api-version=2 --listen=127.0.0.1:4300 ./cmd/otfd/main.go

.PHONY: connect
connect:
	dlv connect 127.0.0.1:4300 .

.PHONY: playwright-ubuntu
install-playwright-ubuntu:
	go tool playwright install chromium --with-deps

.PHONY: playwright-arch
install-playwright-arch:
	go tool playwright install chromium

# run templ generation in watch mode to detect all .templ files and
# re-create _templ.txt files on change, then send reload event to browser.
# Default url: https://localhost:7331
live/templ:
	go tool templ generate --watch --proxy="https://localhost:8080" --open-browser=false --cmd="go run ./cmd/otfd/main.go"

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css --minify --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
live/sync_assets:
	go run github.com/air-verse/air@v1.61.7 \
	--build.cmd "go tool templ generate --notify-proxy" \
	--build.bin "true" \
	--build.delay "100" \
	--build.include_dir "internal/http/html/static" \
	--build.include_ext "js,css,svg"

# start watch processes in parallel.
#
# NOTE: for some reason, if live/templ is placed first in the list it blocks
# the remaining processes, so it's important it is placed first.
live:
	make -j live/tailwind live/sync_assets live/templ

generate-templates:
	go tool templ generate

check-no-diff: paths actions generate-templates
	git diff --exit-code

.PHONY: deploy-otfd
deploy-otfd:
	helm upgrade -i --create-namespace -n otfd-test -f ./charts/otfd/test-values.yaml otfd ./charts/otfd --wait

.PHONY: test-otfd
test-otfd: deploy-otfd
	helm test -n otfd-test otfd

.PHONY: bump-chart-version
bump-chart-version:
	yq -i '.version |= (split(".") | .[-1] |= ((. tag = "!!int") + 1) | join("."))' ./charts/${CHART}/Chart.yaml
