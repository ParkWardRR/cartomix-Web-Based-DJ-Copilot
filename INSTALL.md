# Installation Guide

## One-Command Install

```bash
# Clone and install everything
git clone https://github.com/cartomix/algiers.git
cd algiers
make install
```

That's it! The installer handles everything automatically.

---

## Quick Start

After installation:

```bash
# Start all services
make run-stack

# Open in browser
open http://localhost:5173
```

---

## System Requirements

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| **macOS** | 13 (Ventura) | 14+ (Sonoma) |
| **Chip** | Any Mac | Apple Silicon (M1-M5) |
| **RAM** | 8 GB | 16 GB |
| **Storage** | 2 GB | 5 GB (for audio files) |

### Software Dependencies

These are installed automatically by `make install`:

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.24+ | Backend engine |
| **Swift** | 6+ | Audio analyzer |
| **Node.js** | 22+ | Web UI |
| **Buf** | Latest | Protobuf generation |
| **FFmpeg** | Latest | Screenshots (optional) |

---

## Manual Installation

If you prefer to install manually:

### 1. Install Homebrew (if needed)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### 2. Install Dependencies

```bash
# Required
brew install go node

# Optional (for protobuf and screenshots)
brew install bufbuild/buf/buf ffmpeg webp
```

### 3. Install Xcode Command Line Tools

```bash
xcode-select --install
```

### 4. Clone the Repository

```bash
git clone https://github.com/cartomix/algiers.git
cd algiers
```

### 5. Build Components

```bash
# Go engine
go mod download
go build -o algiers-engine ./cmd/engine

# Swift analyzer
cd analyzer-swift
swift build -c release
cd ..

# Web UI
cd web
npm install
npm run build
cd ..
```

### 6. Run

```bash
# Start engine (Terminal 1)
./algiers-engine

# Start web dev server (Terminal 2)
cd web && npm run dev
```

Open http://localhost:5173

---

## Make Commands

Run `make help` to see all available commands:

### Installation
| Command | Description |
|---------|-------------|
| `make install` | First-time setup (install deps + build) |
| `make install-deps` | Install dependencies only |

### Build
| Command | Description |
|---------|-------------|
| `make build` | Build all components |
| `make build-engine` | Build Go engine |
| `make build-analyzer` | Build Swift analyzer |
| `make build-web` | Build web UI |
| `make proto` | Regenerate protobuf stubs |

### Run
| Command | Description |
|---------|-------------|
| `make run-stack` | Start all services (recommended) |
| `make run-engine` | Start Go engine |
| `make run-analyzer` | Start Swift analyzer |
| `make run-web` | Start Vite dev server |
| `make dev` | Show development instructions |

### Test
| Command | Description |
|---------|-------------|
| `make test` | Run all tests |
| `make test-go` | Run Go unit tests |
| `make test-swift` | Run Swift tests |
| `make test-e2e` | Run Playwright E2E tests |
| `make test-coverage` | Generate coverage report |

### Other
| Command | Description |
|---------|-------------|
| `make lint` | Run all linters |
| `make screenshots` | Capture UI screenshots |
| `make doctor` | Check system health |
| `make version` | Show version info |
| `make clean` | Remove build artifacts |

---

## Ports

| Service | Port | Protocol |
|---------|------|----------|
| Web UI | 5173 | HTTP |
| Engine (HTTP) | 8080 | HTTP REST |
| Engine (gRPC) | 50051 | gRPC |
| Analyzer | 9090 | HTTP/gRPC |

### Custom Ports

```bash
# Change engine ports
ENGINE_PORT=9080 GRPC_PORT=9051 make run-engine

# Change data directory
ALGIERS_DATA_DIR=/path/to/data make run-engine
```

---

## Troubleshooting

### "command not found: go/swift/node"

Ensure Homebrew is in your PATH:
```bash
eval "$(/opt/homebrew/bin/brew shellenv)"
```

Add to `~/.zshrc` to make permanent.

### Port already in use

Check what's using the port:
```bash
lsof -i :8080
```

Kill the process or use a different port.

### Swift build fails

Ensure Xcode CLI tools are installed:
```bash
xcode-select --install
```

### Web build fails (TypeScript errors)

Clear node modules and reinstall:
```bash
cd web
rm -rf node_modules package-lock.json
npm install
```

### Make command not found

On macOS, make should be available. If not:
```bash
xcode-select --install
```

---

## Updating

```bash
git pull
make build
```

---

## Uninstalling

```bash
# Remove build artifacts
make clean

# Remove data directory (optional)
rm -rf ~/.algiers

# Remove the repository
cd ..
rm -rf algiers
```

---

## Support

- **Issues**: https://github.com/cartomix/algiers/issues
- **Discussions**: https://github.com/cartomix/algiers/discussions

---

## License

Blue Oak Model License 1.0.0
