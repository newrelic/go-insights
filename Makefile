PROJECT_NAME := $(shell basename $(shell pwd))
GO_PKGS      := $(shell go list ./... | grep -v -e "/vendor/" -e "/example")
GO_FILES     := $(shell find ./ -type f -name "*.go")
NATIVEOS     := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH   := $(shell go version | awk -F '[ /]' '{print $$5}')
SRCDIR       ?= .
BUILD_DIR    := ./bin/
COVERAGE_DIR := ./coverage/
GOTOOLS       = github.com/kardianos/govendor \
                gopkg.in/alecthomas/gometalinter.v2 \
                github.com/axw/gocov/gocov \
                github.com/AlekSi/gocov-xml \
                github.com/stretchr/testify/assert \
                github.com/robertkrimen/godocdown/godocdown \


GO           = govendor
GODOC        = godocdown
GOMETALINTER = gometalinter.v2
GOVENDOR     = govendor

# Determine packages by looking into pkg/*
PACKAGES=$(wildcard ${SRCDIR}/pkg/*)

# Determine commands by looking into cmd/*
COMMANDS=$(wildcard ${SRCDIR}/cmd/*)

# Determine binary names by stripping out the dir names
BINS=$(foreach cmd,${COMMANDS},$(notdir ${cmd}))

#ifeq (${COMMANDS},)
#  $(error Could not determine COMMANDS, set SRCDIR or run in source dir)
#endif
#ifeq (${BINS},)
#  $(error Could not determine BINS, set SRCDIR or run in source dir)
#endif


all: build

build: check-version clean validate test coverage compile document

clean:
	@echo "=== $(PROJECT_NAME) === [ clean            ]: removing binaries and coverage file..."
	@rm -rfv $(BUILD_DIR)/* $(COVERAGE_DIR)/*

tools: check-version
	@echo "=== $(PROJECT_NAME) === [ tools            ]: Installing tools required by the project..."
	@$(GO) get $(GOTOOLS)
	@$(GOMETALINTER) --install

tools-update: check-version
	@echo "=== $(PROJECT_NAME) === [ tools-update     ]: Updating tools required by the project..."
	@$(GO) get -u $(GOTOOLS)
	@$(GOMETALINTER) --install

deps: tools deps-only

deps-only:
	@echo "=== $(PROJECT_NAME) === [ deps             ]: Installing package dependencies required by the project..."
	@$(GOVENDOR) sync

validate: deps
	@echo "=== $(PROJECT_NAME) === [ validate         ]: Validating source code running gometalinter..."
	@$(GOMETALINTER) --config=.gometalinter.json ./...

compile-only: deps-only
	@echo "=== $(PROJECT_NAME) === [ compile          ]: building commands:"
	@for b in $(BINS); do \
		echo "=== $(PROJECT_NAME) === [ compile          ]:     $$b"; \
		BUILD_FILES=`find $(SRCDIR)/cmd/$$b -type f -name "*.go"` ; \
		$(GO) build -o $(BUILD_DIR)/$$b $$BUILD_FILES ; \
	done

compile: deps compile-only

coverage:
	@echo "=== $(PROJECT_NAME) === [ coverage         ]: generating coverage results..."
	@rm -rf $(COVERAGE_DIR)/*
	@for d in $(GO_PKGS); do \
		pkg=`basename $$d` ;\
		$(GO) test -coverprofile $(COVERAGE_DIR)/$$pkg.tmp $$d ;\
	done
	@echo 'mode: set' > $(COVERAGE_DIR)/coverage.out
# || true to ignore grep return code if no matches (i.e. no tests written...)
	@cat $(COVERAGE_DIR)/*.tmp | grep -v 'mode: set' >> $(COVERAGE_DIR)/coverage.out || true
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

test-unit:
	@echo "=== $(PROJECT_NAME) === [ unit-test        ]: running unit tests..."
	@$(GO) test -tags unit $(GO_PKGS)

test-integration:
	@echo "=== $(PROJECT_NAME) === [ integration-test ]: running integrtation tests..."
	@$(GO) test -tags integration $(GO_PKGS)

document:
	@echo "=== $(PROJECT_NAME) === [ documentation    ]: Generating Godoc in Markdown..."
	@for p in $(PACKAGES); do \
		echo "=== $(PROJECT_NAME) === [ documentation    ]:     $$p"; \
		$(GODOC) $$p > $$p/README.md ; \
	done

test-only: test-unit test-integration
test: test-deps test-only

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
