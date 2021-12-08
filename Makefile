VERSION = $(shell git describe --tags --dirty --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
RANDOM_SUFFIX := $(shell cat /dev/urandom | tr -dc 'a-z0-9' | head -c5)
IMAGE_NAME = otf
IMAGE_TAG ?= $(VERSION)-$(RANDOM_SUFFIX)
LD_FLAGS = " \
	-X 'github.com/leg100/otf.Version=$(VERSION)'	\
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
	./hack/harness.bash ./hack/go-tfe-tests.bash

.PHONY: e2e
e2e: build
	./hack/harness.bash go test ./e2e -failfast

.PHONY: unit
unit:
	go test $$(go list ./... | grep -v e2e)

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
	set -e ;\
	ZIP_FILE=$$(tempfile --prefix=otf --suffix=.zip) ;\
	RELEASE_URL=$$(curl -s https://api.github.com/repos/leg100/otf/releases/latest | \
		jq -r '.assets[] | select(.name | test(".*_linux_amd64.zip$$")) | .browser_download_url') ;\
	curl -Lo $$ZIP_FILE $$RELEASE_URL ;\
	unzip -o -d $(GOBIN) $$ZIP_FILE ;\
	}


# Run staticcheck metalinter recursively against code
.PHONY: lint
lint:
	staticcheck ./...

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

# Push docker image
.PHONY: push
push: image
	docker tag $(IMAGE_NAME):latest $(IMAGE_TARGET)
	docker push $(IMAGE_TARGET)
