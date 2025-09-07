GOCMD=go
GOBUILD=$(GOCMD) build
PATH := "${CURDIR}/bin:$(PATH)"

.PHONY: gobuildcache

bin/golangci-lint:
	script/bindown install $(notdir $@)

bin/shellcheck:
	script/bindown install $(notdir $@)

bin/gobin:
	script/bindown install $(notdir $@)

bin/mockgen:
	script/bindown install $(notdir $@)

GOFUMPT_REV := v0.1.0
bin/gofumpt: bin/gobin
	GOBIN=${CURDIR}/bin \
	bin/gobin mvdan.cc/gofumpt@$(GOFUMPT_REV)

BENCHSTAT_REV := 40a54f11e90963acb1c431127af77c095654c32d
bin/benchstat:
	GOBIN=${CURDIR}/bin \
	go install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_REV)
