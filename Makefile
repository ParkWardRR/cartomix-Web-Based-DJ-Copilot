GO_TEST?=go test ./...
SWIFT_TEST?=cd analyzer-swift && swift test
WEB_DIR=web

.PHONY: all test go-test swift-test web-test fixturegen proto proto-lint lint go-lint web-lint build go-build web-build run dev clean run-stack screenshots screenshots-install

# Default target
all: build

# Build targets
build: go-build web-build

go-build:
	go build ./...

web-build:
	npm run build --prefix $(WEB_DIR)

# Test targets
test: go-test swift-test web-test

go-test:
	$(GO_TEST)

swift-test:
	$(SWIFT_TEST)

web-test:
	npm test --prefix $(WEB_DIR) --if-present

# Lint targets
lint: go-lint proto-lint web-lint

go-lint:
	@which golangci-lint > /dev/null || (echo "Install golangci-lint: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

proto-lint:
	buf lint

web-lint:
	npm run lint --prefix $(WEB_DIR) --if-present

# Proto generation
proto:
	buf generate

# Fixture generation
fixturegen:
	go run ./cmd/fixturegen --out ./testdata/audio

# Screenshots (requires web UI running at localhost:5173)
screenshots-install:
	go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium

screenshots:
	@echo "Ensure web UI is running: make run-web"
	@echo "Requires ffmpeg + libwebp: brew install ffmpeg webp"
	go run ./cmd/screenshots -headless=true -out=docs/assets/screens -webp=true -gif=false

screenshots-gif:
	@echo "Ensure web UI is running: make run-web"
	@echo "Requires ffmpeg: brew install ffmpeg"
	go run ./cmd/screenshots -headless=true -out=docs/assets/screens -webp=false -gif=true

screenshots-no-video:
	@echo "Ensure web UI is running: make run-web"
	go run ./cmd/screenshots -headless=true -out=docs/assets/screens -webp=false -gif=false

screenshots-headed:
	@echo "Ensure web UI is running: make run-web"
	go run ./cmd/screenshots -headless=false -out=docs/assets/screens -webp=false -gif=false

verify-export:
	go run ./cmd/exportverify --manifest testdata/audio/set-checksums.txt --dir testdata/audio || echo "Provide your manifest path"

# Development servers
run: run-engine

run-engine:
	go run ./cmd/engine

run-web:
	npm run dev --prefix $(WEB_DIR)

run-stack:
	./scripts/dev-stack.sh

dev:
	@echo "Run these in separate terminals:"
	@echo "  make run-engine"
	@echo "  make run-web"

# Clean targets
clean:
	rm -rf $(WEB_DIR)/dist
	go clean ./...

# Install dependencies
deps: go-deps web-deps

go-deps:
	go mod download

web-deps:
	npm install --prefix $(WEB_DIR)
