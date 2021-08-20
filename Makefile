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

bin/jq:
	script/bindown install $(notdir $@)

HANDCRAFTED_REV := 082e94edadf89c33db0afb48889c8419a2cb46a9
bin/handcrafted: bin/gobin
	GOBIN=${CURDIR}/bin \
	bin/gobin github.com/willabides/handcrafted@$(HANDCRAFTED_REV)

GOFUMPT_REV := v0.1.0
bin/gofumpt: bin/gobin
	GOBIN=${CURDIR}/bin \
	bin/gobin mvdan.cc/gofumpt@$(GOFUMPT_REV)

GO_FUZZ_REV := b1f3d6f4ef4e0fab65fa66f9191e6b115ad34f31
bin/go-fuzz-build: bin/gobin
	GOBIN=${CURDIR}/bin \
	go install github.com/dvyukov/go-fuzz/go-fuzz-build@$(GO_FUZZ_REV)

bin/go-fuzz: bin/gobin
	GOBIN=${CURDIR}/bin \
	go install github.com/dvyukov/go-fuzz/go-fuzz@$(GO_FUZZ_REV)

BENCHSTAT_REV := 40a54f11e90963acb1c431127af77c095654c32d
bin/benchstat:
	GOBIN=${CURDIR}/bin \
	go install golang.org/x/perf/cmd/benchstat@$(BENCHSTAT_REV)
