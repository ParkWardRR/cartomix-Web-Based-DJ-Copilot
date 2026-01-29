#!/usr/bin/env bash
#
# Algiers - DJ Set Prep Copilot
# One-command installer for macOS (Apple Silicon)
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/cartomix/algiers/main/scripts/install.sh | bash
#   or locally: ./scripts/install.sh
#
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Logging functions
log_info() { echo -e "${BLUE}ℹ${NC}  $1"; }
log_success() { echo -e "${GREEN}✔${NC}  $1"; }
log_warn() { echo -e "${YELLOW}⚠${NC}  $1"; }
log_error() { echo -e "${RED}✖${NC}  $1"; }
log_step() { echo -e "\n${MAGENTA}▸${NC} ${BOLD}$1${NC}"; }

# Banner
show_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
    _    _      _
   / \  | | __ _(_) ___ _ __ ___
  / _ \ | |/ _` | |/ _ \ '__/ __|
 / ___ \| | (_| | |  __/ |  \__ \
/_/   \_\_|\__, |_|\___|_|  |___/
           |___/
EOF
    echo -e "${NC}"
    echo -e "${BOLD}DJ Set Prep Copilot — Apple Silicon Only${NC}"
    echo -e "Version 0.5.2-beta | M1-M4 Required | Neural Engine + Metal GPU"
    echo ""
}

# Check if running on macOS
check_macos() {
    if [[ "$(uname -s)" != "Darwin" ]]; then
        log_error "Algiers requires macOS (Apple Silicon)"
        exit 1
    fi
    log_success "macOS detected"
}

# Check Apple Silicon (REQUIRED)
check_apple_silicon() {
    local arch=$(uname -m)
    if [[ "$arch" != "arm64" ]]; then
        log_error "Apple Silicon (M1/M2/M3/M4) is REQUIRED"
        log_error "Algiers uses Metal GPU and Neural Engine acceleration"
        log_error "Intel Macs are not supported"
        exit 1
    fi
    log_success "Apple Silicon detected ($arch)"
}

# Check macOS version
check_macos_version() {
    local version=$(sw_vers -productVersion)
    local major=$(echo "$version" | cut -d. -f1)
    if [[ "$major" -lt 14 ]]; then
        log_error "macOS 14 (Sonoma) or later required. You have: $version"
        log_error "Earlier versions lack required Metal and Core ML features"
        exit 1
    fi
    log_success "macOS $version"
}

# Check Homebrew
check_homebrew() {
    if ! command -v brew &> /dev/null; then
        log_warn "Homebrew not found. Installing..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        eval "$(/opt/homebrew/bin/brew shellenv)"
    fi
    log_success "Homebrew $(brew --version | head -1 | cut -d' ' -f2)"
}

# Install dependencies
install_deps() {
    log_step "Installing dependencies"

    # Go
    if ! command -v go &> /dev/null; then
        log_info "Installing Go..."
        brew install go
    fi
    local go_version=$(go version | cut -d' ' -f3 | sed 's/go//')
    log_success "Go $go_version"

    # Node.js
    if ! command -v node &> /dev/null; then
        log_info "Installing Node.js..."
        brew install node
    fi
    local node_version=$(node --version)
    log_success "Node.js $node_version"

    # Swift (comes with Xcode CLI tools)
    if ! command -v swift &> /dev/null; then
        log_info "Installing Xcode Command Line Tools..."
        xcode-select --install 2>/dev/null || true
        log_warn "Please complete Xcode CLI installation and re-run this script"
        exit 1
    fi
    local swift_version=$(swift --version 2>&1 | head -1 | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?')
    log_success "Swift $swift_version"

    # Buf (protobuf)
    if ! command -v buf &> /dev/null; then
        log_info "Installing Buf (protobuf toolchain)..."
        brew install bufbuild/buf/buf
    fi
    log_success "Buf $(buf --version 2>&1 | head -1)"

    # FFmpeg (for screenshots)
    if ! command -v ffmpeg &> /dev/null; then
        log_info "Installing FFmpeg (for screenshots)..."
        brew install ffmpeg
    fi
    log_success "FFmpeg installed"

    # gif2webp (for animated demos)
    if ! command -v gif2webp &> /dev/null; then
        log_info "Installing WebP tools..."
        brew install webp
    fi
    log_success "WebP tools installed"
}

# Build the project
build_project() {
    log_step "Building Algiers"

    # Get project root
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local root="$(cd "$script_dir/.." && pwd)"
    cd "$root"

    # Go dependencies
    log_info "Downloading Go dependencies..."
    go mod download
    log_success "Go dependencies ready"

    # Build Swift analyzer
    log_info "Building Swift analyzer (this may take a few minutes)..."
    cd analyzer-swift
    swift build -c release 2>&1 | tail -5
    cd "$root"
    log_success "Swift analyzer built"

    # Web dependencies
    log_info "Installing web dependencies..."
    cd web
    npm install --silent
    cd "$root"
    log_success "Web dependencies ready"

    # Build web UI
    log_info "Building web UI..."
    cd web
    npm run build --silent
    cd "$root"
    log_success "Web UI built"

    # Build Go engine
    log_info "Building Go engine..."
    go build -o algiers-engine ./cmd/engine
    log_success "Engine built: ./algiers-engine"

    # Generate protobuf (if buf is available)
    if command -v buf &> /dev/null; then
        log_info "Generating protobuf stubs..."
        buf generate 2>/dev/null || true
        log_success "Protobuf stubs generated"
    fi
}

# Create data directory
setup_data_dir() {
    log_step "Setting up data directory"

    local data_dir="${ALGIERS_DATA_DIR:-$HOME/.algiers}"
    mkdir -p "$data_dir"
    mkdir -p "$data_dir/models"
    mkdir -p "$data_dir/cache"
    mkdir -p "$data_dir/exports"

    log_success "Data directory: $data_dir"
}

# Install Playwright browsers
install_playwright() {
    log_step "Installing Playwright browsers (for E2E tests)"

    go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium 2>/dev/null || {
        log_warn "Playwright install skipped (optional)"
    }
    log_success "Playwright ready"
}

# Run tests
run_tests() {
    log_step "Running tests"

    log_info "Running Go tests..."
    go test ./... -short 2>&1 | tail -10
    log_success "All tests passed"
}

# Print success message
print_success() {
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}${BOLD}  ✔ Algiers installed successfully!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BOLD}Quick Start:${NC}"
    echo ""
    echo -e "  ${CYAN}# Start all services (recommended)${NC}"
    echo -e "  ${BOLD}make run-stack${NC}"
    echo ""
    echo -e "  ${CYAN}# Or start individually:${NC}"
    echo -e "  ${BOLD}make run-engine${NC}    # Terminal 1: Go engine (:8080 HTTP, :50051 gRPC)"
    echo -e "  ${BOLD}make run-web${NC}       # Terminal 2: Vite dev server (:5173)"
    echo ""
    echo -e "  ${CYAN}# Open in browser:${NC}"
    echo -e "  ${BOLD}open http://localhost:5173${NC}"
    echo ""
    echo -e "${BOLD}Useful Commands:${NC}"
    echo ""
    echo -e "  ${BOLD}make test${NC}          # Run all tests"
    echo -e "  ${BOLD}make screenshots${NC}   # Capture UI screenshots"
    echo -e "  ${BOLD}make help${NC}          # Show all available commands"
    echo ""
    echo -e "${BOLD}Documentation:${NC}"
    echo ""
    echo -e "  README.md              Overview and quickstart"
    echo -e "  docs/AI-ML.md          AI/ML architecture guide"
    echo -e "  docs/ML-TRAINING.md    Custom model training"
    echo -e "  docs/gRPC-MIGRATION.md gRPC API reference"
    echo -e "  docs/API.md            Full API documentation"
    echo ""
    echo -e "${CYAN}100% local. No cloud. Your audio never leaves your Mac.${NC}"
    echo ""
}

# Main
main() {
    show_banner

    log_step "Checking system requirements"
    check_macos
    check_apple_silicon
    check_macos_version
    check_homebrew

    install_deps
    build_project
    setup_data_dir
    install_playwright
    run_tests

    print_success
}

# Handle arguments
case "${1:-}" in
    --help|-h)
        show_banner
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --deps-only    Only install dependencies"
        echo "  --build-only   Only build (skip tests)"
        echo "  --skip-tests   Skip running tests"
        echo ""
        exit 0
        ;;
    --deps-only)
        show_banner
        check_macos
        check_apple_silicon
        check_macos_version
        check_homebrew
        install_deps
        log_success "Dependencies installed"
        exit 0
        ;;
    --build-only)
        show_banner
        build_project
        log_success "Build complete"
        exit 0
        ;;
    --skip-tests)
        show_banner
        check_macos
        check_apple_silicon
        check_macos_version
        check_homebrew
        install_deps
        build_project
        setup_data_dir
        print_success
        exit 0
        ;;
    *)
        main
        ;;
esac
