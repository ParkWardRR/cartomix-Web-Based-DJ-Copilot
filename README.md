<div align="center">

<img src="docs/assets/logo.svg" alt="Algiers" width="120" />

# Algiers

### The AI-Powered DJ Set Prep Copilot

**Analyze. Match. Plan. Export. — All running locally on your Mac.**

<br/>

[![Download](https://img.shields.io/badge/Download-v1.0--beta-6366f1?style=for-the-badge&logo=apple&logoColor=white)](https://github.com/ParkWardRR/cartomix-Web-Based-DJ-Copilot/releases/download/v1.0-beta/Algiers-v1.0-beta-AppleSilicon.dmg)
[![macOS](https://img.shields.io/badge/macOS%2014+-000000?style=for-the-badge&logo=apple&logoColor=white)](#requirements)
[![Apple Silicon](https://img.shields.io/badge/Apple%20Silicon-M1--M4-000000?style=for-the-badge&logo=apple&logoColor=white)](#requirements)

<br/>

[![License](https://img.shields.io/badge/License-Blue%20Oak%201.0-blue?style=flat-square)](LICENSE)
[![Notarized](https://img.shields.io/badge/Apple%20Notarized-✓-34C759?style=flat-square)](#installation)
[![Swift](https://img.shields.io/badge/Swift-6.0-F05138?style=flat-square&logo=swift&logoColor=white)](#tech-stack)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go&logoColor=white)](#tech-stack)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black)](#tech-stack)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=flat-square&logo=typescript&logoColor=white)](#tech-stack)

<br/>

![Algiers Hero Screenshot](docs/assets/screens/algiers-demo.webp)

<br/>

**100% Local** · **No Cloud** · **Neural Engine Accelerated** · **Private by Design**

</div>

---

## What is Algiers?

Algiers is a **native macOS application** that transforms how you prep DJ sets. Point it at your music folder, and it will:

| Feature | Description |
|---------|-------------|
| **Analyze** | BPM, musical key, energy levels, song sections, cue points |
| **Match** | Find tracks with similar "vibe" using neural network embeddings |
| **Plan** | Optimize set order with AI-explainable transitions |
| **Export** | Rekordbox, Serato, Traktor, M3U8, JSON formats |

Everything runs **100% locally** on your Mac. Your music **never leaves your device**.

---

## Key Features

### Neural-Powered Vibe Matching

OpenL3 neural network creates **512-dimensional embeddings** for each track, capturing the sonic "vibe" beyond just BPM and key:

```
Similar tracks to "Get Lucky":
├─ Lose Yourself to Dance (92% match) — same vibe, same key
├─ Redbone (78% match) — similar vibe, Δ-4 BPM
└─ Midnight City (71% match) — similar energy arc
```

### Intelligent Set Planning

Weighted graph optimization considers multiple factors:

- **Tempo compatibility** — BPM delta and beatgrid alignment
- **Harmonic mixing** — Camelot wheel key relationships
- **Energy flow** — Building and dropping intensity
- **Vibe continuity** — OpenL3 similarity scores

Every transition shows **why** it works:

```
"similar vibe (82%); Δ+2 BPM; key: 8A→9A (compatible); energy +1"
```

### Waveform Section Editing

Edit track sections directly on the waveform canvas:
- **Click + drag** to create new sections
- **Resize handles** on section edges
- **Right-click menu** for section type selection
- **8 section types**: Intro, Build, Drop, Breakdown, Outro, Verse, Chorus, Body

### Batch Operations

Efficient bulk track management:
- **Select All / Select None** with one click
- **Multi-select mode** with checkboxes
- **Batch analyze** multiple tracks simultaneously
- **Keyboard shortcuts** for power users

### Keyboard Shortcuts

| Category | Shortcut | Action |
|----------|----------|--------|
| Navigation | `↓` / `J` | Next track |
| Navigation | `↑` / `K` | Previous track |
| Views | `1` - `5` | Switch between Library, Set Builder, Graph, Settings, Training |
| Selection | `S` | Toggle batch select mode |
| Selection | `⌘A` | Select all tracks |
| Selection | `Space` | Toggle selection on current track |
| Help | `?` | Show keyboard shortcuts |

---

## Screenshots

<table>
<tr>
<td align="center"><b>Library View</b></td>
<td align="center"><b>Set Builder</b></td>
<td align="center"><b>Transition Graph</b></td>
</tr>
<tr>
<td><img src="docs/assets/screens/algiers-library-view.png" width="300"/></td>
<td><img src="docs/assets/screens/algiers-set-builder.png" width="300"/></td>
<td><img src="docs/assets/screens/algiers-graph-view.png" width="300"/></td>
</tr>
</table>

---

## Installation

### Download (Recommended)

**[Download Algiers v1.0-beta](https://github.com/ParkWardRR/cartomix-Web-Based-DJ-Copilot/releases/download/v1.0-beta/Algiers-v1.0-beta-AppleSilicon.dmg)** (~21 MB)

1. Open the DMG
2. Drag **Algiers** to Applications
3. Double-click to launch

> **Note:** The app is signed and notarized by Apple — no Gatekeeper warnings.

### Build from Source

```bash
git clone https://github.com/ParkWardRR/cartomix-Web-Based-DJ-Copilot.git
cd cartomix-Web-Based-DJ-Copilot
bash scripts/build-and-notarize.sh
open build/Algiers.app
```

---

## Requirements

| Requirement | Minimum |
|-------------|---------|
| **macOS** | 14.0 (Sonoma) or later |
| **Chip** | Apple Silicon (M1 / M2 / M3 / M4) |
| **RAM** | 8 GB |
| **Storage** | 500 MB |

> **Note:** Intel Macs are not supported. Algiers requires Metal GPU and Neural Engine for ML inference.

---

## Architecture

Algiers bundles three optimized components in a single app:

```
Algiers.app/
├── MacOS/Algiers               # SwiftUI app shell (WebView + process management)
├── Helpers/
│   ├── algiers-engine          # Go HTTP server + set planner + SQLite storage
│   └── analyzer-swift          # Apple Silicon audio analyzer (vDSP + Core ML)
└── Resources/
    ├── Models/
    │   └── OpenL3.mlpackage    # Core ML model for vibe matching (~15MB)
    └── web/                    # React frontend (production build)
```

### Data Flow

```
┌─────────────┐     HTTP      ┌────────────────┐     gRPC      ┌─────────────────┐
│   React UI  │ ◄──────────► │  Go Engine     │ ◄───────────► │  Swift Analyzer │
│  (WebView)  │    REST API   │  (HTTP:8080)   │    Proto3     │  (gRPC:50052)   │
└─────────────┘               └────────────────┘               └─────────────────┘
                                     │                                  │
                                     ▼                                  ▼
                              ┌──────────────┐                ┌──────────────────┐
                              │   SQLite DB  │                │   Core ML Model  │
                              │  (WAL mode)  │                │   (Neural Engine)│
                              └──────────────┘                └──────────────────┘
```

---

## Tech Stack

### Frontend

| Technology | Version | Purpose |
|------------|---------|---------|
| **React** | 19 | UI framework with concurrent rendering |
| **TypeScript** | 5.6 | Type-safe development |
| **Vite** | 7.3 | Build tooling with HMR |
| **Framer Motion** | 12 | Fluid animations |
| **D3.js** | 7 | Force-directed transition graph |
| **Zustand** | 5 | State management |

### Backend

| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.24 | HTTP server, set planner, graph algorithms |
| **SQLite** | 3.46 | Local database with WAL mode |
| **gRPC** | 1.65 | High-performance analyzer communication |
| **Protocol Buffers** | 3 | Binary serialization |

### Analyzer

| Technology | Purpose |
|------------|---------|
| **Swift 6** | Native Apple Silicon performance |
| **Accelerate vDSP** | FFT, spectrograms, signal processing |
| **Core ML** | Neural Engine inference for OpenL3 |
| **AVFoundation** | Audio decode (FLAC, AAC, MP3, WAV, AIFF) |
| **Apple Sound Analysis** | Section boundary detection |

---

## Apple Silicon Optimization

Algiers is built specifically for Apple Silicon's unified architecture:

| Engine | Framework | Use Case | Performance |
|--------|-----------|----------|-------------|
| **Neural Engine** | Core ML | OpenL3 embeddings | ~5ms/window |
| **GPU** | Metal | Spectrograms, onset detection | 10x faster than CPU |
| **CPU** | Accelerate vDSP | FFT, key detection, beatgrid | SIMD optimized |
| **Media Engine** | AVFoundation | Audio decode | Hardware accelerated |

**Unified Memory Architecture** means zero-copy data flow between all engines — no data transfer overhead.

### Performance Benchmarks

| Operation | M1 | M2 Pro | M3 Max |
|-----------|-------|--------|--------|
| Full track analysis | ~8s | ~5s | ~3s |
| Vibe embedding | ~2s | ~1.5s | ~1s |
| 100-track set planning | ~200ms | ~150ms | ~100ms |

---

## Audio Analysis

### Beatgrid Detection

- **Multi-pass tempo detection** with confidence scoring
- **Dynamic tempo maps** for tracks with tempo changes
- **Downbeat alignment** for accurate phrase mapping
- **Phase-locked beatgrid** at millisecond precision

### Key Detection

- **Krumhansl-Schmuckler** pitch class profiling
- **Camelot notation** output (1A-12B)
- **Open Key** notation support
- **Confidence scoring** for ambiguous keys

### Section Tagging

Automatic detection of DJ-relevant sections:

| Section | Description |
|---------|-------------|
| **Intro** | Opening with minimal elements |
| **Build** | Rising energy and tension |
| **Drop** | Peak energy moment |
| **Breakdown** | Energy reduction, melodic focus |
| **Outro** | Closing section for mixing out |

### Cue Point Suggestions

Up to 8 suggested cue points:

| Cue Type | Purpose |
|----------|---------|
| **Load** | Recommended load point |
| **FirstDownbeat** | First phrase downbeat |
| **Drop** | Main drop entry |
| **Breakdown** | Breakdown entry |
| **Build** | Build-up entry |
| **OutroStart** | Outro mixing point |
| **SafetyLoop** | Emergency loop zone |

---

## Export Formats

| Format | Output Files |
|--------|--------------|
| **Rekordbox** | `DJ_PLAYLISTS.xml` with cues, tempo, key, color coding |
| **Serato** | Binary `.crate` file + cues CSV |
| **Traktor** | NML v19 with `CUE_V2` markers |
| **Generic** | M3U8 playlist, JSON analysis, CSV cues |
| **Bundle** | tar.gz archive with all formats |

---

## Privacy

- Audio files are **never uploaded** anywhere
- Analysis runs **100% locally** on your Mac
- **No telemetry**, no analytics, no cloud sync
- App works completely **offline**
- Database stored in `~/Library/Application Support/Algiers/`

---

## Development

### Project Structure

```
.
├── Algiers/                    # Xcode project (SwiftUI wrapper)
├── analyzer-swift/             # Swift audio analyzer
│   ├── Sources/
│   │   └── AudioAnalyzer/      # Core analysis engine
│   └── Package.swift
├── cmd/engine/                 # Go HTTP server
│   ├── main.go
│   ├── handlers/               # REST API handlers
│   ├── planner/                # Set optimization
│   └── storage/                # SQLite repository
├── web/                        # React frontend
│   ├── src/
│   │   ├── components/         # UI components
│   │   ├── hooks/              # Custom hooks
│   │   └── store.ts            # Zustand state
│   └── package.json
├── scripts/
│   └── build-and-notarize.sh   # Full build pipeline
└── docs/
    └── assets/                 # Screenshots, logo
```

### Development Mode

```bash
# Terminal 1: Swift analyzer
cd analyzer-swift
swift build -c release
.build/release/analyzer-swift serve --port 50052 --proto grpc

# Terminal 2: Go engine
go run ./cmd/engine --analyzer-addr localhost:50052

# Terminal 3: React frontend
cd web
npm install
npm run dev
```

### Building for Distribution

```bash
# Full build with code signing and notarization
bash scripts/build-and-notarize.sh

# Output: build/Algiers-v1.0-beta-AppleSilicon.dmg
```

---

## Roadmap

### Completed

- [x] Standalone macOS app with code signing
- [x] Apple notarization for Gatekeeper
- [x] Intro wizard for first-run onboarding
- [x] OpenL3 neural vibe matching
- [x] Rekordbox/Serato/Traktor export
- [x] Drag-and-drop folder import
- [x] Real-time analysis progress indicator
- [x] Modern UI with gradient styling
- [x] Waveform-based section editing
- [x] Batch operations (select all, analyze all)
- [x] Keyboard shortcuts for power users

### Planned

- [ ] Live audio preview with crossfade simulation
- [ ] Custom ML model training on your preferences
- [ ] Harmonic mixing assistant
- [ ] Set history and statistics
- [ ] Apple Music library integration

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

---

## License

Blue Oak Model License 1.0.0. See [LICENSE](LICENSE).

---

<div align="center">

### Built for DJs who want to prep smarter, not harder.

<br/>

*Made with Swift, Go, React, and way too much coffee.*

<br/>

**[Download Algiers](https://github.com/ParkWardRR/cartomix-Web-Based-DJ-Copilot/releases/download/v1.0-beta/Algiers-v1.0-beta-AppleSilicon.dmg)**

</div>
