<div align="center">

<img src="docs/assets/logo.svg" alt="Algiers" width="100" />

# Algiers

**Standalone macOS app for DJ set prep — analyze, match, plan, export.**

[![Download](https://img.shields.io/badge/Download-v0.9--beta-blue?style=for-the-badge)](https://github.com/ParkWardRR/cartomix-Web-Based-DJ-Copilot/releases/download/v0.9-beta/Algiers-v0.9-beta-AppleSilicon.dmg)
[![macOS](https://img.shields.io/badge/macOS%2014+-000000?style=for-the-badge&logo=apple&logoColor=white)](#requirements)
[![Apple Silicon](https://img.shields.io/badge/Apple%20Silicon-M1--M4-000000?style=for-the-badge&logo=apple&logoColor=white)](#requirements)
[![Notarized](https://img.shields.io/badge/Apple%20Notarized-✓-34C759?style=for-the-badge)](#installation)

<br/>

![Algiers Demo](docs/assets/screens/algiers-demo.webp)

**100% local · No cloud · Neural Engine + Metal accelerated · Private by design**

</div>

---

## What is Algiers?

A native macOS app that helps you prep DJ sets. Point it at your music folder, and it will:

- **Analyze** — BPM, key, energy, sections, cue points
- **Match** — Find tracks with similar "vibe" using ML embeddings
- **Plan** — Optimize set order with explainable transitions
- **Export** — Rekordbox, Serato, Traktor formats

Everything runs locally on your Mac. Your music never leaves your device.

---

## Installation

### Download (Recommended)

**[⬇️ Download Algiers v0.9-beta](https://github.com/ParkWardRR/cartomix-Web-Based-DJ-Copilot/releases/download/v0.9-beta/Algiers-v0.9-beta-AppleSilicon.dmg)** (~21 MB)

1. Open the DMG
2. Drag **Algiers** to Applications
3. Double-click to launch

The app is signed and notarized by Apple — no Gatekeeper warnings.

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
| **macOS** | 14.0 (Sonoma) |
| **Chip** | Apple Silicon (M1/M2/M3/M4) |
| **RAM** | 8 GB |
| **Storage** | 500 MB |

> **Note:** Intel Macs are not supported. Algiers requires Metal GPU and Neural Engine for ML inference.

---

## Features

### Audio Analysis
- **Beatgrid** — Tempo detection with dynamic tempo maps
- **Key Detection** — Krumhansl-Schmuckler algorithm with Camelot notation
- **Energy** — 1-10 scale with per-section curves
- **Sections** — Intro, Build, Drop, Breakdown, Outro detection
- **Cue Points** — Up to 8 beat-aligned suggestions
- **Loudness** — EBU R128 broadcast standard (LUFS, LU, true peak)

### Vibe Matching
OpenL3 neural network creates 512-dimensional "vibe" embeddings for each track:

```
Similar tracks to "Get Lucky":
├─ Lose Yourself to Dance (92% match) — same vibe, same key
├─ Redbone (78% match) — similar vibe, Δ-4 BPM
└─ Midnight City (71% match) — similar energy arc
```

### Set Planning
Weighted graph optimization considers:
- Tempo compatibility (BPM delta)
- Key compatibility (Camelot wheel)
- Energy flow (building/dropping)
- Vibe continuity (OpenL3 similarity)

Every transition shows **why** it works:
```
"similar vibe (82%); Δ+2 BPM; key: 8A→9A (compatible); energy +1"
```

### Export Formats
| Format | Output |
|--------|--------|
| **Rekordbox** | DJ_PLAYLISTS XML with cues, tempo, key |
| **Serato** | Binary .crate + cues CSV |
| **Traktor** | NML v19 with CUE_V2 markers |
| **Generic** | M3U8, JSON, CSV, tar.gz bundle |

---

## Screenshots

| Library | Set Builder | Graph |
|:---:|:---:|:---:|
| ![Library](docs/assets/screens/algiers-library-view.png) | ![Set Builder](docs/assets/screens/algiers-set-builder.png) | ![Graph](docs/assets/screens/algiers-graph-view.png) |

---

## How It Works

Algiers bundles three components in one app:

```
Algiers.app/
├── MacOS/Algiers           # SwiftUI app shell
├── Helpers/
│   ├── algiers-engine      # Go HTTP server + set planner
│   └── analyzer-swift      # Apple Silicon audio analyzer
└── Resources/
    ├── Models/OpenL3.mlpackage  # Core ML model for vibe matching
    └── web/                     # React frontend
```

The analyzer uses Apple's Accelerate framework (vDSP) for DSP and Core ML for ML inference on the Neural Engine. No cloud services, no telemetry.

---

## Apple Silicon Optimization

Algiers is built specifically for Apple Silicon:

| Engine | Framework | Use Case |
|--------|-----------|----------|
| **Neural Engine** | Core ML | OpenL3 embeddings (~5ms/window) |
| **GPU** | Metal | Spectrograms, onset detection |
| **CPU** | Accelerate vDSP | FFT, key detection, beatgrid |
| **Media Engine** | AVFoundation | Audio decode (FLAC/AAC/MP3) |

Unified Memory Architecture means zero-copy data flow between engines.

---

## Privacy

- Audio files are **never uploaded** anywhere
- Analysis runs **100% locally** on your Mac
- No telemetry, no analytics, no cloud sync
- App works completely **offline**

---

## Development

### Project Structure

```
.
├── Algiers/                 # Xcode project (SwiftUI wrapper)
├── analyzer-swift/          # Swift audio analyzer
├── cmd/engine/              # Go HTTP server
├── web/                     # React frontend
├── scripts/                 # Build & notarization scripts
└── docs/                    # Documentation
```

### Build Commands

```bash
# Full build with notarization (requires Apple Developer ID)
bash scripts/build-and-notarize.sh

# Development mode (3 terminals)
cd analyzer-swift && swift build -c release && .build/release/analyzer-swift serve --port 50052 --proto grpc
go run ./cmd/engine --analyzer-addr localhost:50052
cd web && npm run dev
```

### Tech Stack

| Layer | Technology |
|-------|------------|
| App Shell | SwiftUI, WKWebView |
| Frontend | React 19, TypeScript, Vite, D3.js |
| Backend | Go 1.24, HTTP REST, gRPC |
| Analyzer | Swift 6, Accelerate, Core ML |
| Storage | SQLite (WAL mode) |

---

## Roadmap

- [x] Standalone macOS app with code signing
- [x] Apple notarization for Gatekeeper
- [x] Intro wizard for first-run onboarding
- [x] OpenL3 vibe matching
- [x] Rekordbox/Serato/Traktor export
- [x] Drag-and-drop folder import
- [x] Real-time analysis progress indicator
- [x] Modern UI refresh with gradient styling
- [x] Waveform-based section editing
- [x] Batch operations (select all, analyze all)
- [ ] Live audio preview with crossfade
- [ ] Custom ML model training
- [ ] Keyboard shortcuts for power users

---

## License

Blue Oak Model License 1.0.0. See [LICENSE](LICENSE).

---

<div align="center">

**Built for DJs who want to prep smarter, not harder.**

*Made with Apple Silicon, Metal, Core ML, and too much coffee.*

</div>
