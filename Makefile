SHELL := /usr/bin/env bash -o pipefail
GOPKG ?= github.com/MrEhbr/app
GOBINS = cmd/analyzer/error_style

include rules.mk

.PHONY: go.build_plugin
go.build_plugin:
	@set -xe; \
	  for dir in $(GOBINS); do ( \
			test -f $$dir/plugin.go && $(GO) build  -buildmode=plugin $(GO_BUILD_OPTS) -o .bin/`basename $$dir`/`basename $$dir`_plugin.so $$dir/plugin.go; \
	  ); done
