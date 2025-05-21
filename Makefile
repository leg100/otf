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
	pnpm exec tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css

.PHONY: tailwind-watch
tailwind-watch:
	pnpm exec tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css --watch

.PHONY: test
test:
	OTF_INTEGRATION_TEST_ENABLE_LOGGER=true go test ./... -v -parallel 1 -failfast -count 10

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

.PHONY: install-goimports
install-goimports:
	go get -tool golang.org/x/tools/cmd/goimports@v0.32.0

# Generate path helpers
.PHONY: paths
paths: install-goimports
	go generate ./internal/http/html/paths
	go tool goimports -w ./internal/http/html/paths
	go tool goimports -w ./internal/http/html/components/paths

.PHONY: install-stringer
install-stringer:
	go get -tool golang.org/x/tools/cmd/stringer@v0.32.0

# Re-generate RBAC action strings
.PHONY: actions
actions: install-stringer
	go tool stringer -type Action ./internal/authz

.PHONY: debug
debug:
	dlv debug --headless --api-version=2 --listen=127.0.0.1:4300 ./cmd/otfd/main.go

.PHONY: connect
connect:
	dlv connect 127.0.0.1:4300 .

.PHONY: install-playwright
install-playwright:
	go get -tool github.com/playwright-community/playwright-go/cmd/playwright@v0.5200.0

.PHONY: playwright-ubuntu
install-playwright-ubuntu: install-playwright
	go tool playwright install chromium --with-deps

.PHONY: playwright-arch
install-playwright-arch: install-playwright
	go tool playwright install chromium

# run templ generation in watch mode to detect all .templ files and
# re-create _templ.txt files on change, then send reload event to browser.
# Default url: http://localhost:7331
live/templ: install-templ
	go tool templ generate --watch --proxy="https://localhost:8080" --open-browser=false --cmd="go run ./cmd/otfd/main.go"

# run tailwindcss to generate the styles.css bundle in watch mode.
live/tailwind:
	tailwindcss -i ./internal/http/html/static/css/input.css -o ./internal/http/html/static/css/output.css --minify --watch

# watch for any js or css change in the assets/ folder, then reload the browser via templ proxy.
live/sync_assets:
	go run github.com/air-verse/air@v1.61.7 \
	--build.cmd "templ generate --notify-proxy" \
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

install-templ:
	go get -tool github.com/a-h/templ/cmd/templ@v0.3.865

generate-templates: install-templ
	go tool templ generate

check-no-diff: paths actions generate-templates
	git diff --exit-code
