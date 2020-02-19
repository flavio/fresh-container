GOMOD ?= on
GO ?= GO111MODULE=$(GOMOD) go

#Don't enable mod=vendor when GOMOD is off or else go build/install will fail
GOMODFLAG ?=-mod=vendor
ifeq ($(GOMOD), off)
GOMODFLAG=
endif

#retrieve go version details for version check
GO_VERSION     := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
GO_VERSION_MAJ := $(shell echo $(GO_VERSION) | cut -f1 -d'.')
GO_VERSION_MIN := $(shell echo $(GO_VERSION) | cut -f2 -d'.')

GOFMT ?= gofmt
GO_MD2MAN ?= go-md2man
LN = ln
RM = rm

BINPATH       := $(abspath ./bin)
GOBINPATH     := $(shell $(GO) env GOPATH)/bin
COMMIT        := $(shell git rev-parse HEAD)
BUILD_DATE    := $(shell date +%Y%m%d)
# TAG can be provided as an envvar (provided in the .spec file)
TAG           ?= $(shell git describe --tags --exact-match HEAD 2> /dev/null)
# CLOSEST_TAG can be provided as an envvar (provided in the .spec file)
CLOSEST_TAG   ?= $(shell git describe --tags)
# VERSION is inferred from CLOSEST_TAG
# It accepts tags of type `vX.Y.Z`, `vX.Y.Z-(alpha|beta|rc|...)` and produces X.Y.Z
VERSION       := $(shell echo $(CLOSEST_TAG) | sed -E 's/v(([0-9]\.?)+).*/\1/')
TAGS          := development
PROJECT_PATH  := github.com/flavio/fresh-container
FRESH_CONTAINER_MONITOR_LDFLAGS  = -ldflags "-X=$(PROJECT_PATH)/pkg/fresh_container.Version=$(VERSION) \
                           -X=$(PROJECT_PATH)/pkg/fresh_container.BuildDate=$(BUILD_DATE) \
                           -X=$(PROJECT_PATH)/pkg/fresh_container.Tag=$(TAG) \
                           -X=$(PROJECT_PATH)/pkg/fresh_container.ClosestTag=$(CLOSEST_TAG)"

FRESH_CONTAINER_MONITOR_DIRS = cmd pkg internal

# go source files, ignore vendor directory
FRESH_CONTAINER_MONITOR_SRCS = $(shell find $(FRESH_CONTAINER_MONITOR_DIRS) -type f -name '*.go')

.PHONY: all
all: install

.PHONY: build
build: go-version-check
	$(GO) build $(GOMODFLAG) $(FRESH_CONTAINER_MONITOR_LDFLAGS) -tags $(TAGS) ./cmd/...

MANPAGES_MD := $(wildcard docs/man/*.md)
MANPAGES    := $(MANPAGES_MD:%.md=%)

docs/man/%.1: docs/man/%.1.md
	$(GO_MD2MAN) -in $< -out $@

.PHONY: docs
docs: $(MANPAGES)

.PHONY: install
install: go-version-check
	$(GO) install $(GOMODFLAG) $(FRESH_CONTAINER_MONITOR_LDFLAGS) -tags $(TAGS) ./cmd/...

.PHONY: clean
clean:
	$(GO) clean -i ./...
	$(RM) -f ./fresh-container
	$(RM) -rf $(BINPATH)

.PHONY: distclean
distclean: clean
	$(GO) clean -i -cache -testcache -modcache ./...

.PHONY: staging
staging:
	make TAGS=staging install

.PHONY: release
release:
	make TAGS=release install

.PHONY: go-version-check
go-version-check:
	@[ $(GO_VERSION_MAJ) -ge 2 ] || \
		[ $(GO_VERSION_MAJ) -eq 1 -a $(GO_VERSION_MIN) -ge 12 ] || (echo "FATAL: Go version should be >= 1.12.x" ; exit 1 ; )

.PHONY: lint
lint: deps
	# explicitly enable GO111MODULE otherwise go mod will fail
	GO111MODULE=on go mod tidy && GO111MODULE=on go mod vendor && GO111MODULE=on go mod verify
	# run go vet
	$(GO) vet ./...
	# run go gmt
	test -z `$(GOFMT) -l $(FRESH_CONTAINER_MONITOR_SRCS)` || { $(GOFMT) -d $(FRESH_CONTAINER_MONITOR_SRCS) && false; }
	# run golangci-lint
	$(BINPATH)/golangci-lint run --verbose --timeout=3m

.PHONY: deps
deps:
	test -f $(BINPATH)/golangci-lint || curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BINPATH) v1.21.0

.PHONY: pre-commit-install
pre-commit-install:
	test -f $(BINPATH)/bin/pre-commit || curl -sfL https://pre-commit.com/install-local.py | HOME=$(BINPATH) python -
	$(BINPATH)/bin/pre-commit install

.PHONY: pre-commit-uninstall
pre-commit-uninstall:
	test -f $(BINPATH)/bin/pre-commit || curl -sfL https://pre-commit.com/install-local.py | HOME=$(BINPATH) python -
	$(BINPATH)/bin/pre-commit uninstall

.PHONY: suse-package
suse-package:
	ci/packaging/suse/rpmfiles_maker.sh "$(VERSION)" "$(TAG)" "$(CLOSEST_TAG)"

.PHONY: suse-changelog
suse-changelog:
	ci/packaging/suse/changelog_maker.sh "$(CHANGES)"

# tests
.PHONY: test
test: test-unit test-bench

.PHONY: test-unit
test-unit:
	$(GO) test $(GOMODFLAG) -coverprofile=coverage.out $(PROJECT_PATH)/{cmd,pkg,internal}/...

.PHONY: test-unit-coverage
test-unit-coverage: test-unit
	$(GO) tool cover -html=coverage.out

.PHONY: test-bench
test-bench:
	$(GO) test $(GOMODFLAG) -bench=. $(PROJECT_PATH)/{cmd,pkg,internal}/...
