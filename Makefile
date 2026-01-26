# Gibson SDK Makefile
# The SDK is a library - no binary to compile, but we build examples and run tests

.PHONY: all bin test test-race test-coverage lint fmt vet tidy clean deps check proto proto-deps proto-clean taxonomy-gen taxonomy-proto generate verify-generated help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Directories
BIN_DIR=bin
EXAMPLES_DIR=examples
PROTO_DIR=api/proto
PROTO_OUT=api/gen
TAXONOMYPB_OUT=$(PROTO_OUT)/taxonomypb
WORKFLOWPB_OUT=$(PROTO_OUT)/workflowpb
TOOLSPB_OUT=$(PROTO_OUT)/toolspb

# Taxonomy generation
TAXONOMY_YAML=taxonomy/core.yaml
# Use local gibson for development (use go run github.com/zero-day-ai/gibson/cmd/taxonomy-gen for production)
GIBSON_DIR=../gibson
TAXONOMY_GEN_CMD=cd $(GIBSON_DIR) && go run ./cmd/taxonomy-gen

# Example binaries to build
EXAMPLES=minimal-agent custom-tool custom-plugin

# Default target
all: test bin

# Build example binaries to bin/
bin: $(BIN_DIR)
	@echo "Building SDK examples..."
	@for example in $(EXAMPLES); do \
		if [ -d "$(EXAMPLES_DIR)/$$example" ]; then \
			echo "  Building $$example..."; \
			cd $(EXAMPLES_DIR)/$$example && $(GOBUILD) -o ../../$(BIN_DIR)/$$example . && cd - > /dev/null; \
		fi; \
	done
	@echo "Build complete: $(BIN_DIR)/"
	@ls -la $(BIN_DIR)/

# Create bin directory
$(BIN_DIR):
	@mkdir -p $(BIN_DIR)

# Run all tests
test:
	@echo "Running SDK tests..."
	$(GOTEST) -v ./...
	@echo "Tests complete"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	@echo "Coverage report:"
	@$(GOCMD) tool cover -func=coverage.out

# Generate coverage HTML report
coverage-html: test-coverage
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage HTML report: coverage.html"

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Tidy modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	@for example in $(EXAMPLES); do \
		if [ -d "$(EXAMPLES_DIR)/$$example" ]; then \
			echo "  Tidying $$example..."; \
			cd $(EXAMPLES_DIR)/$$example && $(GOMOD) tidy && cd - > /dev/null; \
		fi; \
	done

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) ./...
	$(GOMOD) tidy
	@for example in $(EXAMPLES); do \
		if [ -d "$(EXAMPLES_DIR)/$$example" ]; then \
			echo "  Dependencies for $$example..."; \
			cd $(EXAMPLES_DIR)/$$example && $(GOGET) ./... && $(GOMOD) tidy && cd - > /dev/null; \
		fi; \
	done

# Run all checks before commit
check: fmt vet lint test
	@echo "All checks passed!"

# Proto generation
proto-deps:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto: proto-deps
	@echo "Generating Go code from proto files..."
	@mkdir -p $(GRAPHRAGPB_OUT) $(WORKFLOWPB_OUT) $(TOOLSPB_OUT) $(PROTO_OUT)/proto
	$(eval PROTOBUF_DIR := $(shell go list -m -f '{{.Dir}}' google.golang.org/protobuf))
	@echo "  Generating taxonomy.proto..."
	@mkdir -p $(TAXONOMYPB_OUT)
	@if [ -f "$(PROTO_DIR)/taxonomy.proto" ]; then \
		protoc --proto_path=$(PROTO_DIR) --proto_path=$(PROTOBUF_DIR) \
			--go_out=$(TAXONOMYPB_OUT) --go_opt=paths=source_relative \
			$(PROTO_DIR)/taxonomy.proto; \
	else \
		echo "    taxonomy.proto not found - run 'make taxonomy-gen' first"; \
	fi
	@echo "  Generating workflow.proto..."
	@protoc --proto_path=$(PROTO_DIR) --proto_path=$(PROTOBUF_DIR) \
		--go_out=$(WORKFLOWPB_OUT) --go_opt=paths=source_relative \
		$(PROTO_DIR)/workflow.proto
	@echo "  Generating tool protos..."
	@protoc --proto_path=$(PROTO_DIR) --proto_path=$(PROTOBUF_DIR) \
		--go_out=$(TOOLSPB_OUT) --go_opt=paths=source_relative \
		$(PROTO_DIR)/tools/*.proto
	@if [ -d "$(TOOLSPB_OUT)/tools" ]; then \
		mv $(TOOLSPB_OUT)/tools/*.pb.go $(TOOLSPB_OUT)/ 2>/dev/null || true; \
		rmdir $(TOOLSPB_OUT)/tools 2>/dev/null || true; \
	fi
	@echo "  Generating common protos..."
	@for proto in $(PROTO_DIR)/*.proto; do \
		if [ "$$(basename $$proto)" != "graphrag.proto" ] && [ "$$(basename $$proto)" != "workflow.proto" ]; then \
			echo "    Generating from $$(basename $$proto)..."; \
			protoc --proto_path=$(PROTO_DIR) --proto_path=$(PROTOBUF_DIR) \
				--go_out=$(PROTO_OUT)/proto --go_opt=paths=source_relative \
				--go-grpc_out=$(PROTO_OUT)/proto --go-grpc_opt=paths=source_relative \
				$$proto; \
		fi \
	done
	@echo "Proto generation complete"

proto-clean:
	@echo "Cleaning generated proto files..."
	@rm -rf $(PROTO_OUT)/taxonomypb/*.pb.go
	@rm -rf $(PROTO_OUT)/workflowpb/*.pb.go
	@rm -rf $(PROTO_OUT)/toolspb/*.pb.go
	@rm -rf $(PROTO_OUT)/proto/*.pb.go

# Taxonomy generation from YAML
taxonomy-gen:
	@echo "Generating taxonomy from YAML..."
	@mkdir -p $(TAXONOMYPB_OUT) graphrag/domain graphrag/validation graphrag/query graphrag/taxonomy
	@cd $(GIBSON_DIR) && go run ./cmd/taxonomy-gen \
		--base ../sdk/$(TAXONOMY_YAML) \
		--output-proto ../sdk/$(PROTO_DIR)/taxonomy.proto \
		--output-domain ../sdk/graphrag/domain/domain_generated.go \
		--output-validators ../sdk/graphrag/validation/validators_generated.go \
		--output-constants ../sdk/graphrag/constants_generated.go \
		--output-query ../sdk/graphrag/query/query_generated.go \
		--output-helpers ../sdk/graphrag/helpers_generated.go \
		--package domain
	@echo "Generating relationships mapping..."
	@cd $(GIBSON_DIR) && go run ./cmd/taxonomy-gen \
		--base ../sdk/$(TAXONOMY_YAML) \
		--output-relationships ../sdk/graphrag/taxonomy/relationships_generated.go \
		--package taxonomy
	@echo "Formatting generated files..."
	@gofmt -w graphrag/domain/domain_generated.go \
		graphrag/validation/validators_generated.go \
		graphrag/constants_generated.go \
		graphrag/query/query_generated.go \
		graphrag/helpers_generated.go \
		graphrag/taxonomy/relationships_generated.go
	@echo "Taxonomy generation complete"

# Generate taxonomy proto
taxonomy-proto: taxonomy-gen proto-deps
	@echo "Generating Go code from taxonomy.proto..."
	@mkdir -p $(TAXONOMYPB_OUT)
	$(eval PROTOBUF_DIR := $(shell go list -m -f '{{.Dir}}' google.golang.org/protobuf))
	@protoc --proto_path=$(PROTO_DIR) --proto_path=$(PROTOBUF_DIR) \
		--go_out=$(TAXONOMYPB_OUT) --go_opt=paths=source_relative \
		$(PROTO_DIR)/taxonomy.proto
	@echo "Taxonomy proto generation complete"

# Full generate: YAML -> Proto -> Go code
generate: taxonomy-gen taxonomy-proto
	@echo "All generation complete!"

# Verify that generated files are up to date
# Used in CI to catch uncommitted generated code changes
verify-generated:
	@echo "Verifying generated files are up to date..."
	@echo "  Backing up current generated files..."
	@mkdir -p .tmp/verify
	@cp graphrag/helpers_generated.go .tmp/verify/helpers_generated.go.bak 2>/dev/null || true
	@cp graphrag/taxonomy/relationships_generated.go .tmp/verify/relationships_generated.go.bak 2>/dev/null || true
	@cp graphrag/domain/domain_generated.go .tmp/verify/domain_generated.go.bak 2>/dev/null || true
	@cp graphrag/validation/validators_generated.go .tmp/verify/validators_generated.go.bak 2>/dev/null || true
	@cp graphrag/constants_generated.go .tmp/verify/constants_generated.go.bak 2>/dev/null || true
	@cp graphrag/query/query_generated.go .tmp/verify/query_generated.go.bak 2>/dev/null || true
	@cp $(PROTO_DIR)/taxonomy.proto .tmp/verify/taxonomy.proto.bak 2>/dev/null || true
	@echo "  Running generation..."
	@$(MAKE) taxonomy-gen > /dev/null 2>&1
	@echo "  Comparing generated files..."
	@DIFF_FOUND=0; \
	for file in \
		graphrag/helpers_generated.go \
		graphrag/taxonomy/relationships_generated.go \
		graphrag/domain/domain_generated.go \
		graphrag/validation/validators_generated.go \
		graphrag/constants_generated.go \
		graphrag/query/query_generated.go \
		$(PROTO_DIR)/taxonomy.proto; do \
		backup=".tmp/verify/$$(basename $$file).bak"; \
		if [ -f "$$backup" ]; then \
			if ! diff -q "$$file" "$$backup" > /dev/null 2>&1; then \
				echo "  ERROR: $$file differs from generated version"; \
				echo "  Run 'make generate' and commit the changes"; \
				DIFF_FOUND=1; \
			fi; \
		fi; \
	done; \
	echo "  Restoring backup files..."; \
	cp .tmp/verify/helpers_generated.go.bak graphrag/helpers_generated.go 2>/dev/null || true; \
	cp .tmp/verify/relationships_generated.go.bak graphrag/taxonomy/relationships_generated.go 2>/dev/null || true; \
	cp .tmp/verify/domain_generated.go.bak graphrag/domain/domain_generated.go 2>/dev/null || true; \
	cp .tmp/verify/validators_generated.go.bak graphrag/validation/validators_generated.go 2>/dev/null || true; \
	cp .tmp/verify/constants_generated.go.bak graphrag/constants_generated.go 2>/dev/null || true; \
	cp .tmp/verify/query_generated.go.bak graphrag/query/query_generated.go 2>/dev/null || true; \
	cp .tmp/verify/taxonomy.proto.bak $(PROTO_DIR)/taxonomy.proto 2>/dev/null || true; \
	rm -rf .tmp/verify; \
	if [ $$DIFF_FOUND -eq 1 ]; then \
		exit 1; \
	fi
	@echo "Generated files are up to date!"

# Help target
help:
	@echo "Gibson SDK - Makefile Targets"
	@echo ""
	@echo "  make bin           - Build example binaries to bin/"
	@echo "  make test          - Run all tests"
	@echo "  make test-race     - Run tests with race detection"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make coverage-html - Generate HTML coverage report"
	@echo "  make lint          - Run golangci-lint"
	@echo "  make fmt           - Format Go code"
	@echo "  make vet           - Run go vet"
	@echo "  make tidy          - Tidy go modules"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make check         - Run all checks (fmt, vet, lint, test)"
	@echo "  make proto         - Generate Go code from proto files"
	@echo "  make proto-deps    - Install protoc plugins"
	@echo "  make proto-clean   - Remove generated proto files"
	@echo "  make taxonomy-gen  - Generate taxonomy from YAML (proto, domain, validators, helpers)"
	@echo "  make taxonomy-proto- Generate Go code from taxonomy.proto"
	@echo "  make generate      - Full generation: YAML -> Proto -> Go"
	@echo "  make verify-generated - Verify generated files are up to date (CI check)"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Note: The SDK is a library. 'make bin' builds the example applications."
