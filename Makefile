# ============================================================================
# Algiers - DJ Set Prep Copilot
# Makefile for building, testing, and running the application
# ============================================================================

.DEFAULT_GOAL := help

# Configuration
GO_TEST ?= go test ./...
SWIFT_DIR = analyzer-swift
WEB_DIR = web
DATA_DIR ?= $(HOME)/.algiers
ENGINE_PORT ?= 8080
GRPC_PORT ?= 50051
ANALYZER_PORT ?= 9090

# Binaries
ENGINE_BIN = algiers-engine
ANALYZER_BIN = $(SWIFT_DIR)/.build/release/analyzer-swift

# Colors for help
CYAN := \033[36m
YELLOW := \033[33m
GREEN := \033[32m
RESET := \033[0m

.PHONY: all build build-engine build-analyzer build-web test test-go test-swift test-web test-e2e lint proto deps clean help install run run-stack run-engine run-analyzer run-web dev screenshots

# ============================================================================
# Main Targets
# ============================================================================

## help: Show this help message
help:
	@echo ""
	@echo "$(CYAN)Algiers$(RESET) - DJ Set Prep Copilot"
	@echo "============================================"
	@echo ""
	@echo "$(YELLOW)Installation:$(RESET)"
	@grep -E '^## install' $(MAKEFILE_LIST) | sed -E 's/^## /  make /' | sed -E 's/: /\t/' | column -t -s $$'\t'
	@echo ""
	@echo "$(YELLOW)Build:$(RESET)"
	@grep -E '^## (build|proto)' $(MAKEFILE_LIST) | sed -E 's/^## /  make /' | sed -E 's/: /\t/' | column -t -s $$'\t'
	@echo ""
	@echo "$(YELLOW)Run:$(RESET)"
	@grep -E '^## (run|dev)' $(MAKEFILE_LIST) | sed -E 's/^## /  make /' | sed -E 's/: /\t/' | column -t -s $$'\t'
	@echo ""
	@echo "$(YELLOW)Test:$(RESET)"
	@grep -E '^## test' $(MAKEFILE_LIST) | sed -E 's/^## /  make /' | sed -E 's/: /\t/' | column -t -s $$'\t'
	@echo ""
	@echo "$(YELLOW)Other:$(RESET)"
	@grep -E '^## (lint|screenshots|clean|fixturegen)' $(MAKEFILE_LIST) | sed -E 's/^## /  make /' | sed -E 's/: /\t/' | column -t -s $$'\t'
	@echo ""
	@echo "$(GREEN)Quick Start:$(RESET)"
	@echo "  make install      # First-time setup"
	@echo "  make run-stack    # Start all services"
	@echo "  open http://localhost:5173"
	@echo ""

## install: First-time setup (install all dependencies and build)
install:
	@chmod +x scripts/install.sh
	@./scripts/install.sh

## install-deps: Install only dependencies (no build)
install-deps:
	@chmod +x scripts/install.sh
	@./scripts/install.sh --deps-only

# ============================================================================
# Build Targets
# ============================================================================

## build: Build all components (engine + analyzer + web)
build: build-engine build-analyzer build-web
	@echo "$(GREEN)✔ All components built$(RESET)"

## build-engine: Build the Go engine binary
build-engine:
	@echo "Building Go engine..."
	@go build -o $(ENGINE_BIN) ./cmd/engine
	@echo "$(GREEN)✔ Engine built: ./$(ENGINE_BIN)$(RESET)"

## build-analyzer: Build the Swift analyzer (release mode)
build-analyzer:
	@echo "Building Swift analyzer..."
	@cd $(SWIFT_DIR) && swift build -c release
	@echo "$(GREEN)✔ Analyzer built: $(ANALYZER_BIN)$(RESET)"

## build-web: Build the web UI for production
build-web:
	@echo "Building web UI..."
	@npm run build --prefix $(WEB_DIR)
	@echo "$(GREEN)✔ Web UI built: $(WEB_DIR)/dist$(RESET)"

## proto: Regenerate protobuf stubs
proto:
	@echo "Generating protobuf stubs..."
	@buf generate
	@echo "$(GREEN)✔ Protobuf stubs generated$(RESET)"

## proto-lint: Lint protobuf definitions
proto-lint:
	@buf lint

# ============================================================================
# Run Targets
# ============================================================================

## run-stack: Start all services in one command (recommended)
run-stack:
	@chmod +x scripts/dev-stack.sh
	@./scripts/dev-stack.sh

## run-engine: Start the Go engine server
run-engine:
	@echo "Starting engine on :$(ENGINE_PORT) (HTTP) and :$(GRPC_PORT) (gRPC)..."
	@ALGIERS_DATA_DIR=$(DATA_DIR) go run ./cmd/engine

## run-analyzer: Start the Swift analyzer server
run-analyzer:
	@echo "Starting Swift analyzer on :$(ANALYZER_PORT)..."
	@$(ANALYZER_BIN) serve --port $(ANALYZER_PORT)

## run-web: Start the Vite dev server
run-web:
	@echo "Starting web dev server on :5173..."
	@npm run dev --prefix $(WEB_DIR)

## dev: Show instructions for development mode
dev:
	@echo ""
	@echo "$(CYAN)Development Mode$(RESET)"
	@echo "================"
	@echo ""
	@echo "Option 1: Run everything in one terminal"
	@echo "  $(GREEN)make run-stack$(RESET)"
	@echo ""
	@echo "Option 2: Run services separately (better for debugging)"
	@echo "  Terminal 1: $(GREEN)make run-engine$(RESET)    # Go engine"
	@echo "  Terminal 2: $(GREEN)make run-analyzer$(RESET)  # Swift analyzer"
	@echo "  Terminal 3: $(GREEN)make run-web$(RESET)       # Vite dev server"
	@echo ""
	@echo "Then open: $(CYAN)http://localhost:5173$(RESET)"
	@echo ""

# ============================================================================
# Test Targets
# ============================================================================

## test: Run all tests (Go + E2E)
test: test-go test-e2e
	@echo "$(GREEN)✔ All tests passed$(RESET)"

## test-go: Run Go unit tests
test-go:
	@echo "Running Go tests..."
	@$(GO_TEST)

## test-swift: Run Swift unit tests
test-swift:
	@echo "Running Swift tests..."
	@cd $(SWIFT_DIR) && swift test

## test-web: Run web tests (if configured)
test-web:
	@npm test --prefix $(WEB_DIR) --if-present

## test-e2e: Run Playwright E2E tests (requires running services)
test-e2e:
	@echo "Running E2E tests..."
	@go test ./internal/e2e/... -v

## test-golden: Run golden comparison tests
test-golden:
	@echo "Running golden tests..."
	@go test ./internal/exporter/... -v -run Golden

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✔ Coverage report: coverage.html$(RESET)"

# ============================================================================
# Lint Targets
# ============================================================================

## lint: Run all linters
lint: lint-go lint-proto lint-web
	@echo "$(GREEN)✔ All linters passed$(RESET)"

lint-go:
	@echo "Linting Go code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)⚠ golangci-lint not found. Install: brew install golangci-lint$(RESET)"; \
	fi

lint-proto:
	@echo "Linting protobuf..."
	@buf lint

lint-web:
	@echo "Linting web code..."
	@npm run lint --prefix $(WEB_DIR) --if-present

# ============================================================================
# Screenshots & Documentation
# ============================================================================

## screenshots: Capture UI screenshots (requires running web UI)
screenshots: screenshots-webp

screenshots-webp:
	@echo "Capturing screenshots (WebP animation)..."
	@echo "$(YELLOW)Ensure web UI is running: make run-web$(RESET)"
	@go run ./cmd/screenshots -headless=true -out=docs/assets/screens -webp=true -gif=false

screenshots-gif:
	@echo "Capturing screenshots (GIF animation)..."
	@go run ./cmd/screenshots -headless=true -out=docs/assets/screens -webp=false -gif=true

screenshots-static:
	@echo "Capturing screenshots (static only)..."
	@go run ./cmd/screenshots -headless=true -out=docs/assets/screens -webp=false -gif=false

screenshots-headed:
	@echo "Capturing screenshots (visible browser)..."
	@go run ./cmd/screenshots -headless=false -out=docs/assets/screens -webp=false -gif=false

screenshots-install:
	@echo "Installing Playwright browsers..."
	@go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium

# ============================================================================
# Utility Targets
# ============================================================================

## fixturegen: Generate test audio fixtures
fixturegen:
	@echo "Generating test fixtures..."
	@go run ./cmd/fixturegen --out ./testdata/audio
	@echo "$(GREEN)✔ Fixtures generated: ./testdata/audio$(RESET)"

## verify-export: Verify export checksums
verify-export:
	@go run ./cmd/exportverify --manifest testdata/audio/set-checksums.txt --dir testdata/audio

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(WEB_DIR)/dist
	@rm -rf $(SWIFT_DIR)/.build
	@rm -f $(ENGINE_BIN)
	@rm -f coverage.out coverage.html
	@go clean ./...
	@echo "$(GREEN)✔ Clean complete$(RESET)"

## deps: Install all dependencies
deps: deps-go deps-web
	@echo "$(GREEN)✔ All dependencies installed$(RESET)"

deps-go:
	@echo "Downloading Go dependencies..."
	@go mod download

deps-web:
	@echo "Installing web dependencies..."
	@npm install --prefix $(WEB_DIR)

## version: Show component versions
version:
	@echo "Algiers v0.5.1-beta"
	@echo ""
	@echo "Components:"
	@echo "  Go:     $$(go version | cut -d' ' -f3)"
	@echo "  Swift:  $$(swift --version 2>&1 | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?')"
	@echo "  Node:   $$(node --version)"
	@echo "  Buf:    $$(buf --version 2>&1 | head -1)"
	@echo ""

## doctor: Check system health
doctor:
	@echo "$(CYAN)Algiers Health Check$(RESET)"
	@echo "===================="
	@echo ""
	@echo "System:"
	@echo "  OS:     $$(sw_vers -productName) $$(sw_vers -productVersion)"
	@echo "  Arch:   $$(uname -m)"
	@echo ""
	@echo "Dependencies:"
	@printf "  Go:     "; command -v go > /dev/null && echo "$(GREEN)✔$(RESET) $$(go version | cut -d' ' -f3)" || echo "$(RED)✖ not found$(RESET)"
	@printf "  Swift:  "; command -v swift > /dev/null && echo "$(GREEN)✔$(RESET) $$(swift --version 2>&1 | head -1 | grep -oE '[0-9]+\.[0-9]+')" || echo "$(RED)✖ not found$(RESET)"
	@printf "  Node:   "; command -v node > /dev/null && echo "$(GREEN)✔$(RESET) $$(node --version)" || echo "$(RED)✖ not found$(RESET)"
	@printf "  Buf:    "; command -v buf > /dev/null && echo "$(GREEN)✔$(RESET) $$(buf --version 2>&1 | head -1)" || echo "$(YELLOW)⚠ not found$(RESET)"
	@printf "  FFmpeg: "; command -v ffmpeg > /dev/null && echo "$(GREEN)✔$(RESET) installed" || echo "$(YELLOW)⚠ not found (optional)$(RESET)"
	@echo ""
	@echo "Ports:"
	@printf "  :$(ENGINE_PORT) (HTTP):  "; lsof -i :$(ENGINE_PORT) > /dev/null 2>&1 && echo "$(YELLOW)in use$(RESET)" || echo "$(GREEN)available$(RESET)"
	@printf "  :$(GRPC_PORT) (gRPC):  "; lsof -i :$(GRPC_PORT) > /dev/null 2>&1 && echo "$(YELLOW)in use$(RESET)" || echo "$(GREEN)available$(RESET)"
	@printf "  :5173 (Vite):  "; lsof -i :5173 > /dev/null 2>&1 && echo "$(YELLOW)in use$(RESET)" || echo "$(GREEN)available$(RESET)"
	@echo ""
