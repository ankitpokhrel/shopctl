##############
# Build vars #
##############

# The 'GIT_COMMIT' variable is used to retrieve the commit hash of the current commit (HEAD).
#
# If there are any uncommitted changes, a stash entry will be created with `git stash create`,
# and the commit hash of the latest commit (HEAD) will be used instead.
GIT_COMMIT ?= $(shell { git stash create; git rev-parse HEAD; } | head -n 1)

# https://reproducible-builds.org/docs/source-date-epoch/
#
# The 'SOURCE_DATE_EPOCH' variable is used to set the timestamp of the commit referenced by 'GIT_COMMIT'.
# It ensures that builds are reproducible by using a consistent, fixed timestamp for the source code.
export SOURCE_DATE_EPOCH ?= $(shell git show -s --format="%ct" $(GIT_COMMIT))

VERSION ?= $(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match)
VERSION_PKG = github.com/ankitpokhrel/shopctl/internal/version
export LDFLAGS += -X $(VERSION_PKG).GitCommit=$(GIT_COMMIT)
export LDFLAGS += -X $(VERSION_PKG).SourceDateEpoch=$(SOURCE_DATE_EPOCH)
export LDFLAGS += -X $(VERSION_PKG).Version=$(VERSION)
export LDFLAGS += -s
export LDFLAGS += -w

export CGO_ENABLED ?= 0
export GOCACHE ?= $(CURDIR)/.gocache

.PHONY: all
all: build

.PHONY: deps
deps:
	go mod vendor -v

.PHONY: build
build: deps
	go build -ldflags='$(LDFLAGS)' ./...

.PHONY: install
install:
	CGO_ENABLED=1 go install -ldflags='$(LDFLAGS)' ./cmd/...

.PHONY: schema
schema:
	SHOPIFY_ACCESS_TOKEN=${SHOPIFY_ACCESS_TOKEN} go run ./cmd/introspect -s ${SHOPIFY_STORE} -t Product -p schema > ./schema/product.go
	gofmt -s -w ./schema/product.go

.PHONY: lint
lint:
	@if ! command -v golangci-lint > /dev/null 2>&1; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b "$$(go env GOPATH)/bin" v1.62.0 ; \
	fi
	golangci-lint run ./...

.PHONY: test
test:
	@go clean -testcache
	CGO_ENABLED=1 go test -race ./...

.PHONY: coverage
coverage:
	go test -cover ./...

.PHONY: dev
dev:
	@docker build -t shopctl:latest .

.PHONY: exec
exec:
	@docker exec -it shopctl sh

.PHONY: ci
ci: lint test

.PHONY: clean
clean:
	go clean -x ./...

.PHONY: distclean
distclean:
	go clean -x -cache -testcache -modcache
