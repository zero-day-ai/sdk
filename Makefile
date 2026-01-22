# Gibson SDK Makefile
# The SDK is a library - no binary to compile, but we build examples and run tests

.PHONY: all bin test test-race test-coverage lint fmt vet tidy clean deps check proto proto-deps proto-clean help

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
GRAPHRAGPB_OUT=$(PROTO_OUT)/graphragpb
WORKFLOWPB_OUT=$(PROTO_OUT)/workflowpb
TOOLSPB_OUT=$(PROTO_OUT)/toolspb

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
	@echo "  Generating graphrag.proto..."
	@protoc --proto_path=$(PROTO_DIR) --proto_path=$(PROTOBUF_DIR) \
		--go_out=$(GRAPHRAGPB_OUT) --go_opt=paths=source_relative \
		$(PROTO_DIR)/graphrag.proto
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
	@rm -rf $(PROTO_OUT)/graphragpb/*.pb.go
	@rm -rf $(PROTO_OUT)/workflowpb/*.pb.go
	@rm -rf $(PROTO_OUT)/toolspb/*.pb.go
	@rm -rf $(PROTO_OUT)/proto/*.pb.go

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
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Note: The SDK is a library. 'make bin' builds the example applications."
