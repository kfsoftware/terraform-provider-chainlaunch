.PHONY: build install generate clean test fmt lint

# Build the provider
build:
	go build -o terraform-provider-chainlaunch

# Install the provider locally
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/chainlaunch/chainlaunch/1.0.0/$$(go env GOOS)_$$(go env GOARCH)
	mv terraform-provider-chainlaunch ~/.terraform.d/plugins/registry.terraform.io/chainlaunch/chainlaunch/1.0.0/$$(go env GOOS)_$$(go env GOARCH)/

# Generate API client from swagger.yaml
generate:
	@echo "Installing go-swagger..."
	go install github.com/go-swagger/go-swagger/cmd/swagger@latest
	@echo "Generating API client..."
	~/go/bin/swagger generate client -f swagger.yaml -t internal/generated
	@echo "Client generated successfully!"

# Clean build artifacts
clean:
	rm -f terraform-provider-chainlaunch
	rm -rf internal/generated/client
	rm -rf internal/generated/models
	go clean

# Run unit tests
test:
	go test ./internal/provider -v -short

# Run unit tests only
test-unit:
	go test ./internal/provider -v -short

# Run integration/acceptance tests
test-integration:
	TF_ACC=1 go test ./internal/provider -v -timeout 30m

# Run E2E tests
test-e2e:
	cd test/e2e && ./run-tests.sh

# Run all tests
test-all: test-unit test-integration test-e2e

# Generate test coverage
test-coverage:
	go test ./internal/provider -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code (Go + Terraform)
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting Terraform files..."
	terraform fmt -recursive .
	@echo "Formatting complete!"

# Check if code is formatted
fmt-check:
	@echo "Checking Go formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Go files need formatting. Run 'make fmt'" && exit 1)
	@echo "Checking Terraform formatting..."
	@terraform fmt -check -recursive . || (echo "Terraform files need formatting. Run 'make fmt'" && exit 1)
	@echo "All files are properly formatted!"

# Lint code (Go)
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install: brew install golangci-lint" && exit 1)
	@bash scripts/lint.sh

# Lint Terraform files
lint-tf:
	@echo "Running tflint..."
	@which tflint > /dev/null || (echo "tflint not installed. Install: brew install tflint" && exit 1)
	@tflint --init
	@tflint --recursive

# Validate Terraform examples
validate:
	@echo "Validating Terraform examples..."
	@for dir in examples/*/; do \
		echo "Validating $$dir..."; \
		(cd "$$dir" && terraform init -backend=false > /dev/null 2>&1 && terraform validate) || echo "⚠️  $$dir validation failed"; \
	done
	@echo "Validation complete!"

# Run all checks (format, lint, validate)
check: fmt-check lint validate
	@echo "✅ All checks passed!"

# Run code quality checks (like pre-commit but manual)
check-code:
	@bash scripts/check-code.sh

# Download dependencies
deps:
	go mod download
	go mod tidy

# Initialize go modules
init:
	go mod init github.com/chainlaunch/terraform-provider-chainlaunch
	go mod tidy

# Run the provider in debug mode
debug:
	go build -gcflags="all=-N -l" -o terraform-provider-chainlaunch
	dlv exec ./terraform-provider-chainlaunch -- -debug

# Generate documentation using tfplugindocs
docs:
	@echo "Generating documentation..."
	@which tfplugindocs > /dev/null || (echo "Installing tfplugindocs..." && go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest)
	tfplugindocs generate --provider-name chainlaunch --rendered-provider-name "Chainlaunch" --rendered-website-dir docs
	@echo "Documentation generated successfully!"

# Validate documentation
docs-validate:
	@echo "Validating documentation..."
	@which tfplugindocs > /dev/null || (echo "Installing tfplugindocs..." && go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest)
	tfplugindocs validate
	@echo "Documentation validation complete!"

# Preview documentation locally
docs-preview:
	@echo "Starting documentation preview..."
	@which glow > /dev/null || (echo "Installing glow..." && brew install glow)
	@cd docs && glow -p .

# Install Git hooks
hooks:
	@bash hooks/install.sh

# Help target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Building:"
	@echo "  build            - Build the provider binary"
	@echo "  install          - Install the provider locally"
	@echo "  clean            - Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  test             - Run unit tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-integration - Run integration/acceptance tests"
	@echo "  test-e2e         - Run end-to-end tests"
	@echo "  test-all         - Run all tests"
	@echo "  test-coverage    - Generate test coverage report"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt              - Format Go and Terraform code"
	@echo "  fmt-check        - Check if code is formatted"
	@echo "  lint             - Lint Go code (requires golangci-lint)"
	@echo "  lint-tf          - Lint Terraform files (requires tflint)"
	@echo "  validate         - Validate Terraform examples"
	@echo "  check            - Run all checks (fmt-check + lint + validate)"
	@echo "  check-code       - Run code quality checks (manual pre-commit)"
	@echo ""
	@echo "Development:"
	@echo "  generate         - Generate API client from swagger.yaml"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  hooks            - Install Git pre-commit hooks"
	@echo "  debug            - Run provider in debug mode"
	@echo "  docs             - Generate provider documentation"
	@echo "  docs-validate    - Validate provider documentation"
	@echo "  docs-preview     - Preview documentation locally with glow"
	@echo "  help             - Show this help message"
