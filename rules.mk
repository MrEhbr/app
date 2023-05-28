.PHONY: _default_entrypoint
_default_entrypoint: help

##
## Common helpers
##

rwildcard = $(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
check-program = $(foreach exec,$(1),$(if $(shell PATH="$(PATH)" which $(exec)),,$(error "No $(exec) in PATH")))
my-filter-out = $(foreach v,$(2),$(if $(findstring $(1),$(v)),,$(v)))
novendor = $(call my-filter-out,vendor/,$(1))

##
## Golang
##

ifndef GOPKG
ifneq ($(wildcard go.mod),)
GOPKG = $(shell sed '/module/!d;s/^omdule\ //' go.mod)
endif
endif
ifdef GOPKG
CGO_ENABLED ?= 0
GO ?= go
GOPATH ?= $(HOME)/go
GO_TEST_OPTS ?= -test.timeout=1m
WHAT = "./..."
VERSION ?= $(shell git describe --exact-match --tags 2> /dev/null || git symbolic-ref -q --short HEAD)
VCS_REF ?= $(shell git rev-parse HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_TAGS ?= ""
GO_LINT_OPTS ?= --verbose
GO_LDFLAGS ?= "-s -w -X main.version=$(VERSION) -X main.commit=$(VCS_REF) -X main.date=$(BUILD_DATE) -X main.builtBy=$(shell whoami)"
GO_INSTALL_OPTS ?= -asmflags=all=-trimpath=$(GOPKG) -gcflags=all=-trimpath=$(GOPKG) -ldflags $(GO_LDFLAGS) -tags $(GO_TAGS)
GO_BUILD_OPTS ?= -asmflags=all=-trimpath=$(GOPKG) -gcflags=all=-trimpath=$(GOPKG) -ldflags $(GO_LDFLAGS) -tags $(GO_TAGS)

GOMOD_DIRS ?= $(sort $(call novendor,$(dir $(call rwildcard,*,*/go.mod go.mod))))
GOCOVERAGE_FILE ?= ./coverage.txt

ifdef GOBINS
.PHONY: go.install
go.install:
	@set -e; for dir in $(GOBINS); do ( set -xe; \
	  cd $$dir; \
	  CGO_ENABLED=$(CGO_ENABLED) $(GO) install $(GO_INSTALL_OPTS) . \
	); done
INSTALL_STEPS += go.install

.PHONY: go.release
go.release:
	$(call check-program, goreleaser)
	goreleaser --snapshot --skip-publish --rm-dist
	@echo -n "Do you want to release? [y/N] " && read ans && \
	  if [ $${ans:-N} = y ]; then set -xe; goreleaser --rm-dist; fi
RELEASE_STEPS += go.release
endif

.PHONY: go.unittest
go.unittest:
	@echo "mode: atomic" > /tmp/gocoverage
	@set -e; for dir in $(GOMOD_DIRS); do (set -e; (set -xe; \
	  cd $$dir; \
	  $(GO) test $(WHAT) $(GO_TEST_OPTS) -cover -coverprofile=/tmp/profile.out -covermode=atomic -race); \
	  if [ -f /tmp/profile.out ]; then \
		cat /tmp/profile.out | sed "/mode: atomic/d" >> /tmp/gocoverage; \
		rm -f /tmp/profile.out; \
	  fi); done
	@mv /tmp/gocoverage $(GOCOVERAGE_FILE)

.PHONY: go.checkdoc
go.checkdoc:
	go doc $(first $(GOMOD_DIRS))

.PHONY: go.coverfunc
go.coverfunc: go.unittest
	go tool cover -func=$(GOCOVERAGE_FILE) | grep -v .pb.go: | grep -v .pb.gw.go:
	@echo "coverage report: $(GOCOVERAGE_FILE)"

.PHONY: go.lint
go.lint:
	@set -e; for dir in $(GOMOD_DIRS); do ( set -xe; \
	  cd $$dir; \
	  golangci-lint run $(GO_LINT_OPTS) $(WHAT); \
	); done

.PHONY: go.tidy
go.tidy:
	@# tidy dirs with go.mod files
	@set -e; for dir in $(GOMOD_DIRS); do ( set -xe; \
	  cd $$dir; \
	  $(GO)	mod tidy; \
	); done

.PHONY: go.download
go.download:
	@# download all deps
	@set -e; for dir in $(GOMOD_DIRS); do ( set -xe; \
	  cd $$dir; \
	  $(GO)	mod download; \
	); done

.PHONY: go.generate
go.generate:
	@set -e; for dir in $(GOMOD_DIRS); do ( set -xe; \
	  cd $$dir; \
	  $(GO) generate -x $(WHAT); \
	); done
GENERATE_STEPS += go.generate

.PHONY: go.depaware-update
go.depaware-update: go.tidy
	@# gen depaware for bins
	@set -e; for dir in $(GOBINS); do ( set -xe; \
	  cd $$dir; \
	  $(GO) run github.com/tailscale/depaware --update . \
	); done
	@# tidy unused depaware deps if not in a tools_test.go file
	@set -e; for dir in $(GOMOD_DIRS); do ( set -xe; \
	  cd $$dir; \
	  $(GO)	mod tidy; \
	); done

.PHONY: go.depaware-check
go.depaware-check: go.tidy
	@# gen depaware for bins
	@set -e; for dir in $(GOBINS); do ( set -xe; \
	  cd $$dir; \
	  $(GO) run github.com/tailscale/depaware --check . \
	); done


.PHONY: go.build
go.build:
	@set -xe; \
	  for dir in $(GOBINS); do ( \
			CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GO_BUILD_OPTS) -o .bin/`basename $$dir`/`basename $$dir`  $$dir/main.go; \
	  ); done

BUILD_STEPS += go.build

.PHONY: go.bump-deps
go.bumpdeps:
	@set -e; for dir in $(GOMOD_DIRS); do ( set -xe; \
	  cd $$dir; \
	  $(GO)	get -u $(WHAT); \
	); done

.PHONY: go.fmt
go.fmt:
	@set -e; for dir in $(GOMOD_DIRS); do ( set -e; \
	  cd $$dir; \
	  $(GO) run mvdan.cc/gofumpt -extra -w -l `go list -f '{{.Dir}}' $(WHAT) | grep -v mocks` \
	); done

VERIFY_STEPS += go.depaware-check
BUILD_STEPS += go.build
BUMPDEPS_STEPS += go.bumpdeps go.depaware-update
TIDY_STEPS += go.tidy
LINT_STEPS += go.lint
UNITTEST_STEPS += go.unittest
FMT_STEPS += go.fmt

# FIXME: disabled, because currently slow
# new rule that is manually run sometimes, i.e. `make pre-release` or `make maintenance`.
# alternative: run it each time the go.mod is changed
#GENERATE_STEPS += go.depaware-update
endif

##
## proto
##

.PHONY: proto.lint
proto.lint:
	buf check lint

# remote is what we run when testing in most CI providers
# this does breaking change detection against our remote git repository
#
.PHONY: proto.lint-remote
proto.lint-remote: $(BUF)
	buf check lint
	buf check breaking --against-input .git#branch=master


.PHONY: proto.generate
proto.generate: $(BUF)
	./scripts/protoc_gen.bash \
		"--proto_path=$(PROTO_PATH)" \
		$(patsubst %,--proto_include_path=%,$(PROTO_INCLUDE_PATHS)) \
		"--plugin_name=go" \
		"--plugin_out=$(PROTOC_GEN_GO_OUT)" \
		"--plugin_opt=$(PROTOC_GEN_GO_OPT)"
ifneq ($(wildcard buf.yml),)
GENERATE_STEPS += proto.generate
endif

ifneq ($(wildcard buf.yml),)
LINT_STEPS += proto.lint
endif

##
## Gitattributes
##

ifneq ($(wildcard .gitattributes),)
.PHONY: _linguist-kept
_linguist-kept:
	@git check-attr linguist-vendored $(shell git check-attr linguist-generated $(shell find . -type f | grep -v .git/) | grep unspecified | cut -d: -f1) | grep unspecified | cut -d: -f1 | sort

.PHONY: _linguist-ignored
_linguist-ignored:
	@git check-attr linguist-vendored linguist-ignored `find . -not -path './.git/*' -type f` | grep '\ set$$' | cut -d: -f1 | sort -u
endif

##
## Docker
##

docker_build = docker build \
	  --build-arg VCS_REF=$(VCS_REF) \
	  --build-arg BUILD_DATE=$(BUILD_DATE) \
	  --build-arg VERSION=$(VERSION) \
	  -t "$2:$(VERSION)" \
	  -f "$1" .

docker_push =	docker push "$1:$(VERSION)"

ifndef DOCKERFILE_PATH
DOCKERFILE_PATH = ./Dockerfile
endif
ifndef DOCKER_IMAGE
ifneq ($(wildcard Dockerfile),)
DOCKER_IMAGE = $(notdir $(PWD))
endif
endif
ifdef DOCKER_IMAGE
ifneq ($(DOCKER_IMAGE),none)
.PHONY: docker.build
docker.build:
	$(call check-program, docker)
	$(call docker_build,$(DOCKERFILE_PATH),$(DOCKER_IMAGE))

.PHONY: docker.push
docker.push:
	$(call check-program, docker)
	$(call docker_push,$(DOCKER_IMAGE))

endif
endif

##
## Common
##

TEST_STEPS += $(UNITTEST_STEPS)
TEST_STEPS += $(LINT_STEPS)
TEST_STEPS += $(TIDY_STEPS)

ifneq ($(strip $(TEST_STEPS)),)
.PHONY: test
test: $(PRE_TEST_STEPS) $(TEST_STEPS)
endif

ifdef INSTALL_STEPS
.PHONY: install
install: $(PRE_INSTALL_STEPS) $(INSTALL_STEPS)
endif

ifdef UNITTEST_STEPS
.PHONY: unittest
unittest: $(PRE_UNITTEST_STEPS) $(UNITTEST_STEPS)
endif

ifdef LINT_STEPS
.PHONY: lint
lint: $(PRE_LINT_STEPS) $(LINT_STEPS)
endif

ifdef TIDY_STEPS
.PHONY: tidy
tidy: $(PRE_TIDY_STEPS) $(TIDY_STEPS)
endif

ifdef BUILD_STEPS
.PHONY: build
build: $(PRE_BUILD_STEPS) $(BUILD_STEPS)
endif

ifdef VERIFY_STEPS
.PHONY: verify
verify: $(PRE_VERIFY_STEPS) $(VERIFY_STEPS)
endif

ifdef RELEASE_STEPS
.PHONY: release
release: $(PRE_RELEASE_STEPS) $(RELEASE_STEPS)
endif

ifdef BUMPDEPS_STEPS
.PHONY: bumpdeps
bumpdeps: $(PRE_BUMDEPS_STEPS) $(BUMPDEPS_STEPS)
endif

ifdef FMT_STEPS
.PHONY: fmt
fmt: $(PRE_FMT_STEPS) $(FMT_STEPS)
endif

ifdef GENERATE_STEPS
.PHONY: generate
generate: $(PRE_GENERATE_STEPS) $(GENERATE_STEPS)
endif

.PHONY: help
help::
	@echo "General commands:"
	@[ "$(BUILD_STEPS)" != "" ]     && echo "  build"     	    || true
	@[ "$(BUMPDEPS_STEPS)" != "" ]  && echo "  bumpdeps"  	    || true
	@[ "$(FMT_STEPS)" != "" ]       && echo "  fmt"       	    || true
	@[ "$(GENERATE_STEPS)" != "" ]  && echo "  generate"  	    || true
	@[ "$(INSTALL_STEPS)" != "" ]   && echo "  install"   	    || true
	@[ "$(LINT_STEPS)" != "" ]      && echo "  lint"      	    || true
	@[ "$(RELEASE_STEPS)" != "" ]   && echo "  release"   	    || true
	@[ "$(TEST_STEPS)" != "" ]      && echo "  test"      	    || true
	@[ "$(TIDY_STEPS)" != "" ]      && echo "  tidy"      	    || true
	@[ "$(UNITTEST_STEPS)" != "" ]  && echo "  unittest"  	    || true
	@[ "$(VERIFY_STEPS)" != "" ]    && echo "  verify"    	    || true
	@[ "$(DOCKER_IMAGE)" != "" ]    && echo "  docker.build"    || true
	@[ "$(DOCKER_IMAGE)" != "" ]    && echo "  docker.push"     || true
	@# FIXME: list other commands

print-% : ; $(info $* is a $(flavor $*) variable set to [$($*)]) @true
