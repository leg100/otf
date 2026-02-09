LD_FLAGS = '-s -w -X github.com/leg100/otf/internal.Version=edge'

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: go-tfe-tests
go-tfe-tests: image-otfd compose-up
	./hack/go-tfe-tests.bash

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build \
		-ldflags $(LD_FLAGS) \
		-o ./_build/linux/amd64/ \
		./cmd/otfd ./cmd/otf-job ./cmd/otf-agent

.PHONY: install
install:
	go install ./...

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
compose-up: image-otfd
	docker compose -f docker-compose.testing.yml up -d --wait --wait-timeout 60

# Remove docker compose stack
.PHONY: compose-rm
compose-rm:
	docker compose -f docker-compose.testing.yml rm -sf

# Run postgresql via docker compose
.PHONY: postgres
postgres:
	docker compose -f docker-compose.testing.yml up -d postgres

# Run staticcheck metalinter recursively against code
.PHONY: lint
lint:
	go tool staticcheck ./...

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...

# Build docker images
.PHONY: images
images: build
	make -j image-otfd image-agent image-job

.PHONY: image-otfd
image-otfd: build
	docker build -f Dockerfile -t leg100/otfd:edge --target otfd _build/

.PHONY: image-agent
image-agent: build
	docker build -f Dockerfile -t leg100/otf-agent:edge --target otf-agent _build/

.PHONY: image-job
image-job: build
	docker build -f Dockerfile -t leg100/otf-job:edge --target otf-job _build/

# Build and load edge images into kubernetes kind
.PHONY: load
load: images
	make -j load-otfd load-agent load-job

.PHONY: load-otfd
load-otfd:
	kind load docker-image leg100/otfd:edge

.PHONY: load-agent
load-agent:
	kind load docker-image leg100/otf-agent:edge

.PHONY: load-job
load-job:
	kind load docker-image leg100/otf-job:edge

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
	go generate ./internal/ui/paths
	go tool goimports -w ./internal/ui/paths

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
	go tool templ generate --watch --proxybind 0.0.0.0 --proxy="https://localhost:8080" --open-browser=false --cmd="make live/run"

live/run:
	go run -ldflags $(LD_FLAGS) ./cmd/otfd/main.go

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
# the remaining processes, so it's important it is placed last.
live:
	make -j live/tailwind live/sync_assets live/templ

generate-templates:
	go tool templ generate

check-no-diff: paths actions generate-templates helm-docs
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

.PHONY: helm-docs
helm-docs:
	go tool helm-docs -c ./charts -u


.PHONY: helm-dependency-update
helm-dependency-update: helm-dependency-update-otfd helm-dependency-update-otf-agent

.PHONY: helm-dependency-update-otfd
helm-dependency-update-otfd:
	helm dependency update ./charts/otfd

.PHONY: helm-dependency-update-otf-agent
helm-dependency-update-otf-agent:

.PHONY: helm-lint
helm-lint:
	./hack/helm-lint.sh
