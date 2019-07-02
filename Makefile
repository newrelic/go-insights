PROJECT_NAME := $(shell basename $(shell pwd))
PROJECT_VER  := $(shell git describe --tags --always --dirty)
GO_PKGS      := $(shell go list ./... | grep -v -e "/vendor/" -e "/example")
NATIVEOS     := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH   := $(shell go version | awk -F '[ /]' '{print $$5}')
SRCDIR       ?= .
BUILD_DIR    := ./bin/
COVERAGE_DIR := ./coverage/
COVERMODE     = atomic
GOTOOLS       = github.com/axw/gocov/gocov \
                github.com/AlekSi/gocov-xml \
                github.com/stretchr/testify/assert \
                github.com/robertkrimen/godocdown/godocdown \
                github.com/golangci/golangci-lint/cmd/golangci-lint

GO_CMD        = go
GODOC         = godocdown
DOC_DIR       = ./docs/
GOLINTER      = golangci-lint

# Determine package dep manager
ifneq ("$(wildcard Gopkg.toml)","")
	VENDOR     = dep
	VENDOR_CMD = ${VENDOR} ensure
	GOTOOLS    += github.com/golang/dep
	GO         = ${GO_CMD}
else ifneq ("$(wildcard Godeps/*)","")
	VENDOR     = godep
	VENDOR_CMD = echo "Not Implemented"
	GOTOOLS    += github.com/tools/godep
	GO         = godep go
else
	VENDOR     = govendor
	VENDOR_CMD = ${VENDOR} sync
	GOTOOLS    += github.com/kardianos/govendor
	GO         = ${VENDOR}
endif

# Determine packages by looking into pkg/*
ifneq ("$(wildcard ${SRCDIR}/pkg/*)","")
	PACKAGES  = $(wildcard ${SRCDIR}/pkg/*)
endif
ifneq ("$(wildcard ${SRCDIR}/internal/*)","")
	PACKAGES += $(wildcard ${SRCDIR}/internal/*)
endif

# Determine commands by looking into cmd/*
COMMANDS = $(wildcard ${SRCDIR}/cmd/*)

GO_FILES := $(shell find $(COMMANDS) $(PACKAGES) -type f -name "*.go")

# Determine binary names by stripping out the dir names
BINS=$(foreach cmd,${COMMANDS},$(notdir ${cmd}))

LDFLAGS='-X main.Version=$(PROJECT_VER)'

all: build

# Humans running make:
build: check-version clean validate test-unit cover-report compile document

# Build command for CI tooling
build-ci: check-version clean validate test compile-only

clean:
	@echo "=== $(PROJECT_NAME) === [ clean            ]: removing binaries and coverage file..."
	@rm -rfv $(BUILD_DIR)/* $(COVERAGE_DIR)/*

tools: check-version
	@echo "=== $(PROJECT_NAME) === [ tools            ]: Installing tools required by the project..."
	@$(GO_CMD) get $(GOTOOLS)

tools-update: check-version
	@echo "=== $(PROJECT_NAME) === [ tools-update     ]: Updating tools required by the project..."
	@$(GO_CMD) get -u $(GOTOOLS)

deps: tools deps-only

deps-only:
	@echo "=== $(PROJECT_NAME) === [ deps             ]: Installing package dependencies required by the project..."
	@echo "=== $(PROJECT_NAME) === [ deps             ]:     Detected '$(VENDOR)'"
	@$(VENDOR_CMD)

validate: deps
	@echo "=== $(PROJECT_NAME) === [ validate         ]: Validating source code running $(GOLINTER)..."
	@$(GOLINTER) run ./...

compile-only: deps-only
	@echo "=== $(PROJECT_NAME) === [ compile          ]: building commands:"
	@for b in $(BINS); do \
		echo "=== $(PROJECT_NAME) === [ compile          ]:     $$b"; \
		BUILD_FILES=`find $(SRCDIR)/cmd/$$b -type f -name "*.go"` ; \
		$(GO) build -ldflags=$(LDFLAGS) -o $(BUILD_DIR)/$$b $$BUILD_FILES ; \
	done

compile: deps compile-only

test: test-deps test-only
test-only: test-unit test-integration

test-unit:
	@echo "=== $(PROJECT_NAME) === [ unit-test        ]: running unit tests..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -tags unit -covermode=$(COVERMODE) -coverprofile $(COVERAGE_DIR)/unit.tmp $(GO_PKGS)

test-integration:
	@echo "=== $(PROJECT_NAME) === [ integration-test ]: running integrtation tests..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -tags integration -covermode=$(COVERMODE) -coverprofile $(COVERAGE_DIR)/integration.tmp $(GO_PKGS)

cover-report:
	@echo "=== $(PROJECT_NAME) === [ cover-report     ]: generating coverage results..."
	@mkdir -p $(COVERAGE_DIR)
	@echo 'mode: $(COVERMODE)' > $(COVERAGE_DIR)/coverage.out
	@cat $(COVERAGE_DIR)/*.tmp | grep -v 'mode: $(COVERMODE)' >> $(COVERAGE_DIR)/coverage.out || true
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "=== $(PROJECT_NAME) === [ cover-report     ]:     $(COVERAGE_DIR)coverage.html"

document:
	@echo "=== $(PROJECT_NAME) === [ documentation    ]: Generating Godoc in Markdown..."
	@for p in $(PACKAGES); do \
		echo "=== $(PROJECT_NAME) === [ documentation    ]:     $$p"; \
		mkdir -p $(DOC_DIR)/$$p ; \
		$(GODOC) $$p > $(DOC_DIR)/$$p/README.md ; \
	done
	@for c in $(COMMANDS); do \
		echo "=== $(PROJECT_NAME) === [ documentation    ]:     $$c"; \
		mkdir -p $(DOC_DIR)/$$c ; \
		$(GODOC) $$c > $(DOC_DIR)/$$c/README.md ; \
	done

check-version:
ifdef GOOS
ifneq "$(GOOS)" "$(NATIVEOS)"
	$(error GOOS is not $(NATIVEOS). Cross-compiling is only allowed for 'clean', 'deps-only' and 'compile-only' targets)
endif
endif
ifdef GOARCH
ifneq "$(GOARCH)" "$(NATIVEARCH)"
	$(error GOARCH variable is not $(NATIVEARCH). Cross-compiling is only allowed for 'clean', 'deps-only' and 'compile-only' targets)
endif
endif

.PHONY: all build clean coverage document document-only document-deps fmt lint vet validate-deps validate-only validate compile-deps compile-only compile test-deps test-unit test-integration test-only test
