default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./internal/provider -v $(TESTARGS) -timeout 120m

# Build the provider
.PHONY: build
build:
	go build -v -o terraform-provider-galaxy .

# Install the provider locally
.PHONY: install
install: build
	go install .

# Generate documentation
.PHONY: docs
docs:
	tfplugindocs generate --provider-name galaxy

# Validate documentation
.PHONY: docs-validate
docs-validate:
	tfplugindocs validate --provider-name galaxy

# Clean up generated files
.PHONY: clean
clean:
	rm -f terraform-provider-galaxy provider

# Format Go code
.PHONY: fmt
fmt:
	go fmt ./...

# Run linter
.PHONY: lint
lint:
	golangci-lint run

# Check and add license headers
.PHONY: license
license:
	@go install github.com/google/addlicense@latest
	@addlicense -check -f .license-header.txt -ignore '**/*.yml' -ignore '**/*.yaml' -ignore '**/*.tf' .

# Add license headers to files missing them
.PHONY: license-add
license-add:
	@go install github.com/google/addlicense@latest
	@addlicense -f .license-header.txt -ignore '**/*.yml' -ignore '**/*.yaml' -ignore '**/*.tf' .

# Run tests
.PHONY: test
test:
	go test -v ./...

# Generate documentation and validate
.PHONY: docs-all
docs-all: docs docs-validate

# Development setup - install provider locally and generate docs
.PHONY: dev-setup
dev-setup: install docs

# Complete testing workflow - build, generate docs, validate, test examples
.PHONY: test-all
test-all: install license lint docs docs-validate testacc

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        Build the provider"
	@echo "  install      Install the provider locally"
	@echo "  test         Run unit tests"
	@echo "  testacc      Run acceptance tests"
	@echo "  docs         Generate documentation"
	@echo "  docs-validate Validate generated documentation"
	@echo "  docs-all     Generate and validate documentation"
	@echo "  test-all     Complete workflow: build + docs + validate + test examples"
	@echo "  fmt          Format Go code"
	@echo "  lint         Run linter"
	@echo "  license      Check license headers"
	@echo "  license-add  Add license headers to files missing them"
	@echo "  clean        Clean up generated files"
	@echo "  dev-setup    Install provider and generate docs"
	@echo "  testacc      Run all acceptance tests"
	@echo "  help         Show this help message"
