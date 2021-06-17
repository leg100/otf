VERSION = $(shell git describe --tags --dirty --always)
GIT_COMMIT = $(shell git rev-parse HEAD)
GOMODULE = github.com/leg100/ots
BUILD_BIN ?= ./_build/ots
LD_FLAGS = " \
	-X '.Version=$(VERSION)'	\
	-X '.Commit=$(GIT_COMMIT)'	\
	" \

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: e2e
e2e: build
	go test -v ./e2e -failfast

.PHONY: unit
unit:
	go test $$(go list ./... | grep -v e2e)

.PHONY: build
build:
	go build -o $(BUILD_BIN) -ldflags $(LD_FLAGS) $(GOMODULE)/cmd/ots
	chmod +x $(BUILD_BIN)

.PHONY: install
install:
	go install -ldflags $(LD_FLAGS) $(GOMODULE)

.PHONY: install-latest-release
install-latest-release:
	{ \
	set -e ;\
	ZIP_FILE=$$(tempfile --prefix=ots --suffix=.zip) ;\
	RELEASE_URL=$$(curl -s https://api.github.com/repos/leg100/ots/releases/latest | \
		jq -r '.assets[] | select(.name | test(".*_linux_amd64.zip$$")) | .browser_download_url') ;\
	curl -Lo $$ZIP_FILE $$RELEASE_URL ;\
	unzip -o -d $(GOBIN) $$ZIP_FILE ;\
	}

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet:
	go vet ./...
