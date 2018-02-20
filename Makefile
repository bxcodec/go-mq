SOURCES := $(shell find . -name '*.go' -type f -not -path './vendor/*'  -not -path '*/mocks/*')

PACKAGE := $(shell go list)
GOOS := $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
OBJ_DIR := $(GOPATH)/pkg/$(GOOS)_$(GOARCH)/$(PACKAGE)

# Dependencies Management
.PHONY: vendor-prepare
vendor-prepare:
	@echo "Installing dep"
	@go get -u -v github.com/golang/dep/cmd/dep

Gopkg.lock: Gopkg.toml
	@dep ensure -update $(DEP_OPTS)

.PHONY: vendor-update
vendor-update:
	@dep ensure -update $(DEP_OPTS)

vendor: Gopkg.lock
	@dep ensure $(DEP_OPTS)

.PHONY: vendor-optimize
vendor-optimize: vendor
	@dep prune

.PHONY: clean-vendor
clean-vendor:
	@rm -rf vendor

# Linter
.PHONY: lint-prepare
lint-prepare:
	@echo "Installing gometalinter"
	@go get -u github.com/alecthomas/gometalinter
	@gometalinter --install

.PHONY: lint
lint: vendor
	@gometalinter --cyclo-over=25 --deadline=2m --vendor ./...

# Testing
.PHONY: test
test: vendor
	@go test -short $(TEST_OPTS) ./...

.PHONY: test-pubsub
test-pubsub: vendor
	@go test -v $(TEST_OPTS) ./pubsub -gcp.project-id "$(GCP_PROJECT)" -gcp.topic-id "$(GCP_TOPIC)" -gcp.subscription-id "$(GCP_SUBSCRIPTION)" -gcp.credentials-file "$(GCP_CREDENTIALS_FILE)"

# Build and Installation
.PHONY: install
install: vendor
	@go install ./...

.PHONY: uninstall
uninstall:
	@echo "Removing binaries and libraries"
	@go clean -i ./...
	@if [ -d $(OBJ_DIR) ]; then \
		rm -rf $(OBJ_DIR); \
	fi
