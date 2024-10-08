BRANCH    ?= $(shell git rev-parse --abbrev-ref HEAD)
BUILDTIME ?= $(shell date '+%Y-%m-%d@%H:%M:%S')
BUILDUSER ?= $(shell id -un)
REPO      ?= github.com/ricoberger/aks-state-exporter
REVISION  ?= $(shell git rev-parse HEAD)
VERSION   ?= $(shell git describe --tags)

.PHONY: build
build:
	@go build -ldflags "-X ${REPO}/pkg/version.Version=${VERSION} \
		-X ${REPO}/pkg/version.Revision=${REVISION} \
		-X ${REPO}/pkg/version.Branch=${BRANCH} \
		-X ${REPO}/pkg/version.BuildUser=${BUILDUSER} \
		-X ${REPO}/pkg/version.BuildDate=${BUILDTIME}" \
		-o ./bin/aks-state-exporter .;

.PHONY: vet
vet:
	@go vet ./...

.PHONY: test
test:
	@go test ./...
