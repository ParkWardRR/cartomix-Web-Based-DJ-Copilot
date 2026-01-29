<div align="center">

<img src="docs/assets/logo.svg" alt="Algiers - DJ Set Prep Copilot" width="100" />

# Algiers — DJ Set Prep Copilot

**Apple Silicon–native DJ prep tool: analyze, cue, transition, export — all local.**

<!-- Status & Version -->
[![Phase](https://img.shields.io/badge/alpha-blueviolet?style=for-the-badge)](#alpha-features)
[![Version](https://img.shields.io/badge/v0.1.0-blue?style=for-the-badge)](#changelog)
[![Local-First](https://img.shields.io/badge/local--first-success?style=for-the-badge)](#what-this-is)
[![Privacy](https://img.shields.io/badge/100%25%20Local-222222?style=for-the-badge&logo=lock&logoColor=white)](#what-this-is)
[![Offline](https://img.shields.io/badge/Offline%20Ready-00C853?style=for-the-badge)](#why-its-different)

<!-- Apple Silicon Platform -->
[![Apple Silicon](https://img.shields.io/badge/Apple%20Silicon-000000?style=for-the-badge&logo=apple&logoColor=white)](#apple-silicon-only)
[![M1–M5](https://img.shields.io/badge/M1--M5%20Native-000000?style=for-the-badge&logo=apple&logoColor=white)](#apple-silicon-only)
[![ANE](https://img.shields.io/badge/Neural%20Engine-FF9500?style=for-the-badge&logo=apple&logoColor=white)](#hardware-acceleration-m1-m5)
[![Metal](https://img.shields.io/badge/Metal%20GPU-147EFB?style=for-the-badge&logo=apple&logoColor=white)](#hardware-acceleration-m1-m5)
[![Accelerate](https://img.shields.io/badge/Accelerate%20vDSP-FF2D55?style=for-the-badge&logo=apple&logoColor=white)](#hardware-acceleration-m1-m5)
[![Core ML](https://img.shields.io/badge/Core%20ML-34C759?style=for-the-badge&logo=apple&logoColor=white)](#hardware-acceleration-m1-m5)

<!-- Backend Stack -->
[![Go](https://img.shields.io/badge/Go%201.24-00ADD8?style=for-the-badge&logo=go&logoColor=white)](#architecture)
[![Swift](https://img.shields.io/badge/Swift%206-F05138?style=for-the-badge&logo=swift&logoColor=white)](#architecture)
[![gRPC](https://img.shields.io/badge/gRPC-4285F4?style=for-the-badge&logo=google&logoColor=white)](#architecture)
[![SQLite](https://img.shields.io/badge/SQLite%20WAL-003B57?style=for-the-badge&logo=sqlite&logoColor=white)](#architecture)
[![Protobuf](https://img.shields.io/badge/Protobuf-4285F4?style=for-the-badge&logo=google&logoColor=white)](#architecture)

<!-- Frontend Stack -->
[![React](https://img.shields.io/badge/React%2019-61DAFB?style=for-the-badge&logo=react&logoColor=black)](#architecture)
[![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](#architecture)
[![Vite](https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white)](#architecture)
[![D3.js](https://img.shields.io/badge/D3.js%207-F9A03C?style=for-the-badge&logo=d3.js&logoColor=white)](#pro-visualizations)
[![Framer Motion](https://img.shields.io/badge/Framer%20Motion-0055FF?style=for-the-badge&logo=framer&logoColor=white)](#pro-visualizations)
[![Zustand](https://img.shields.io/badge/Zustand%205-443E38?style=for-the-badge)](#pro-visualizations)

<!-- Audio Formats -->
[![WAV](https://img.shields.io/badge/WAV-8B5CF6?style=for-the-badge)](#feature-highlights)
[![AIFF](https://img.shields.io/badge/AIFF-8B5CF6?style=for-the-badge)](#feature-highlights)
[![MP3](https://img.shields.io/badge/MP3-8B5CF6?style=for-the-badge)](#feature-highlights)
[![AAC](https://img.shields.io/badge/AAC-8B5CF6?style=for-the-badge)](#feature-highlights)
[![ALAC](https://img.shields.io/badge/ALAC-8B5CF6?style=for-the-badge)](#feature-highlights)
[![FLAC](https://img.shields.io/badge/FLAC-8B5CF6?style=for-the-badge)](#feature-highlights)

<!-- Quality & Docs -->
[![Tests](https://img.shields.io/badge/Go%20%7C%20Playwright-FF69B4?style=for-the-badge)](#developer-loop)
[![Docs](https://img.shields.io/badge/PLAN.md-8E43E7?style=for-the-badge)](docs/PLAN.md)
[![License](https://img.shields.io/badge/Blue%20Oak-lightgray?style=for-the-badge)](LICENSE)

<br/>

### See it in action

![Algiers Demo](docs/assets/screens/algiers-demo.webp)

**100% local. No cloud. ANE + Metal accelerated. Private by design.**

</div>

---

## Table of Contents
- [What This Is](#what-this-is)
- [Alpha Features](#alpha-features)
- [Pro Visualizations](#pro-visualizations)
- [Why It's Different](#why-its-different)
- [Apple Silicon Only](#apple-silicon-only)
- [Hardware Acceleration (M1–M5)](#hardware-acceleration-m1-m5)
- [Feature Highlights](#feature-highlights)
- [Architecture](#architecture)
- [Screenshots](#screenshots)
- [Quick Start](#quick-start)
- [Developer Loop](#developer-loop)
- [Project Layout](#project-layout)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

## What This Is

A local-first copilot for DJ set prep. It ingests your library, detects structure, surfaces mixable sections, proposes cue points and transition windows, and optimizes set order with explainable scoring—while keeping you in control of the mix.

**Key Principles:**
- End-to-end local: no cloud lock-in; audio never leaves your Mac
- Apple-accelerated DSP: tempo, key, and energy analysis run on Metal + Accelerate
- AI where it helps: ANE-powered models suggest cues/sections but every choice is editable
- Explainable transitions: every proposed segue cites tempo delta, key distance, energy window overlap, and beat-grid alignment

## Alpha Features

v0.1.0-alpha brings the first complete UI with professional-grade visualizations:

- **Live Dashboard** — Animated stats with real-time analysis progress, BPM/key distribution charts
- **Canvas Waveform Renderer** — High-performance waveform with section overlays, cue markers, playhead, and beat grid
- **Real-time Spectrum Analyzer** — WebGL-style frequency visualization with mirror mode
- **Set Energy Arc** — Animated SVG energy flow visualization across your set
- **Transition Graph** — D3.js force-directed graph showing track connections and flow
- **Three-View Layout** — Library, Set Builder, and Graph View modes
- **Dark Mode Default** — Pro-first dark theme with light mode toggle

## Pro Visualizations

Built with high-performance libraries for smooth 60fps rendering:

| Component | Technology | Description |
|-----------|-----------|-------------|
| **Waveform Canvas** | Canvas 2D API | GPU-accelerated waveform with sections, cues, and beat markers |
| **Spectrum Analyzer** | Canvas 2D + RAF | Real-time frequency bars with mirror/bars/circular modes |
| **Energy Arc** | SVG + Framer Motion | Animated bezier curve showing set energy journey |
| **Transition Graph** | D3.js Force | Interactive force-directed graph with zoom/pan/drag |
| **BPM/Key Charts** | D3.js | Animated bar charts with hover tooltips |
| **Live Stats** | Framer Motion | Spring-animated number counters and progress rings |

**Libraries Used:**
- **D3.js 7** — Data visualization and force simulation
- **Framer Motion 11** — Spring physics animations and gestures
- **React Virtuoso 4** — Virtualized list rendering (ready for large libraries)
- **Zustand 5** — Lightweight state management

## Why It's Different

| Feature | Algiers | Cloud-based Tools |
|---------|---------|-------------------|
| **Privacy** | Audio never leaves your Mac | Uploads to servers |
| **Speed** | ANE + Metal acceleration | Network latency |
| **Control** | Every edit is yours | AI decides for you |
| **Explainability** | Shows why transitions work | Black box |
| **Offline** | Works without internet | Requires connection |

## Apple Silicon Only

- Built, tested, and tuned for **M1–M5** (M1, M1 Pro/Max/Ultra, M2, M3, M4)
- Assumes unified memory and ANE availability
- Fallback code paths removed to keep performance consistent
- Core ML targets the ANE by default; GPU (Metal) is a secondary path

## Hardware Acceleration (M1–M5)

| Engine | Use Case | Implementation |
|--------|----------|----------------|
| **ANE (Neural Engine)** | Section classification, energy models, downbeat confidence | Core ML with ANE preference |
| **Metal (GPU)** | Real-time spectrograms, onset detection, waveform tiling | Metal Performance Shaders |
| **Accelerate (vDSP/BLAS)** | FFT for beatgrid, harmonic key detection | Vectorized SIMD operations |
| **Unified Memory** | Zero-copy data sharing between CPU/GPU/ANE | Chunked streaming pipelines |

## Feature Highlights

### Library Management
- Resumable library ingest (WAV/AIFF/MP3/AAC/ALAC/FLAC)
- Content-hash caching for instant re-scans
- "Needs review" flagging for low-confidence analysis
- Search by title, artist, or key

### Analysis
- Beatgrid detection (static + dynamic tempo maps)
- Key detection with Camelot wheel mapping
- Energy level (1-10) with per-section curves
- Section detection (Intro/Break/Build/Drop/Outro)
- Up to 8 cue suggestions per track with confidence scores

### Set Planning
- Weighted graph optimization with explainable scoring
- Set modes: Warm-up, Peak-time, Open-format
- Edge explanations: tempo delta, key relation, energy flow
- Transition window matching (intro→outro overlap)

### Export
- M3U8 playlist
- Analysis JSON with full track data
- Cues CSV with beat/time indices
- SHA256 checksum manifest
- Ready-to-share tar.gz bundles
- **Rekordbox XML** - Full DJ_PLAYLISTS schema with cues, tempo markers, and key
- **Serato crate** - Binary .crate format with supplementary cues CSV
- **Traktor NML** - Complete NML v19 export with CUE_V2 markers

## Architecture

Algiers uses a three-layer architecture optimized for Apple Silicon:

```mermaid
flowchart LR
    subgraph UI["Frontend"]
      direction TB
      REACT["React 19 + TypeScript"]
      VIZ["D3.js / Framer Motion"]
      REACT --> VIZ
    end

    subgraph ENGINE["Go Engine (1.24+)"]
      direction TB
      GRPC["gRPC Server"]
      SCHED["Job Scheduler"]
      PLAN["Set Planner"]
      EXPORT["Exporters"]
      GRPC --> SCHED --> PLAN --> EXPORT
    end

    subgraph ANALYZER["Swift Analyzer"]
      direction TB
      DSP["Accelerate vDSP"]
      METAL["Metal GPU"]
      ANE["Core ML / ANE"]
    end

    subgraph DATA["Storage"]
      direction TB
      DB["SQLite WAL"]
      BLOBS["Blob Store"]
    end

    UI <-->|"gRPC-web"| ENGINE
    ENGINE <-->|"gRPC"| ANALYZER
    ENGINE <--> DATA
    ANALYZER --> DATA
```

**How it works:**

1. **Frontend** — React UI sends commands (scan library, plan set, export) via gRPC-web / HTTP API
2. **Go Engine** — Coordinates jobs, runs the weighted-graph set planner, handles exports
3. **Swift Analyzer** — Performs DSP analysis using Apple frameworks (Accelerate, Metal, Core ML)
4. **Storage** — SQLite for metadata, blob store for waveform tiles and embeddings

| Layer | Tech | Role |
|-------|------|------|
| **Frontend** | React 19, TypeScript, Vite, D3.js, Framer Motion | Interactive UI with real-time visualizations |
| **Engine** | Go 1.24, gRPC, Protobuf | API server, job scheduling, set planning, exports |
| **Analyzer** | Swift 6, Accelerate, Metal, Core ML | Audio DSP and ML inference on ANE/GPU |
| **Storage** | SQLite (WAL mode) | Track metadata, analysis results, waveform tiles |

## Detailed Architecture

### Swift Analyzer Pipeline

The Swift analyzer (`analyzer-swift`) is the core audio analysis engine, built with Apple's Accelerate framework for high-performance DSP on Apple Silicon.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Swift Analyzer Pipeline                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐               │
│  │ AudioDecoder │───▶│ FFTProcessor │───▶│   Analysis   │               │
│  │ (AVFoundation)│    │ (Accelerate) │    │   Modules    │               │
│  └──────────────┘    └──────────────┘    └──────────────┘               │
│         │                   │                    │                       │
│         ▼                   ▼                    ▼                       │
│   ┌──────────┐       ┌──────────┐        ┌─────────────┐                │
│   │ 48kHz SR │       │  STFT    │        │ BeatgridDet │                │
│   │ Mono PCM │       │ Spectral │        │ KeyDetector │                │
│   │ Float32  │       │  Flux    │        │ EnergyAnalz │                │
│   └──────────┘       │ Chroma   │        │ SectionDet  │                │
│                      └──────────┘        │ CueGenerator│                │
│                                          └─────────────┘                │
│                                                 │                        │
│                                                 ▼                        │
│                                    ┌────────────────────┐               │
│                                    │ TrackAnalysisResult │               │
│                                    │  - beats, tempoMap  │               │
│                                    │  - key (Camelot)    │               │
│                                    │  - energy, sections │               │
│                                    │  - cues, waveform   │               │
│                                    └────────────────────┘               │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

#### DSP Components

| Component | Framework | Description |
|-----------|-----------|-------------|
| **AudioDecoder** | AVFoundation | Decodes WAV/MP3/AAC/FLAC to 48kHz mono PCM |
| **FFTProcessor** | Accelerate vDSP | FFT with Hann windowing, STFT, spectral flux, chroma features |
| **BeatgridDetector** | Accelerate | Onset detection via spectral flux + autocorrelation tempo estimation |
| **KeyDetector** | Accelerate | Krumhansl-Schmuckler key profiles with chroma correlation |
| **EnergyAnalyzer** | Accelerate | RMS/peak analysis, band energy (low/mid/high), energy curve |
| **SectionDetector** | — | Energy segmentation for intro/verse/build/drop/breakdown/outro |
| **CueGenerator** | — | Priority-based cue selection with beat alignment (max 8 cues) |

#### Key Detection Algorithm

```
1. Extract chroma features from STFT (12 pitch classes)
2. Average chroma vectors across all frames
3. Correlate with Krumhansl-Schmuckler major/minor profiles
4. For each pitch class (0-11):
   - Rotate major profile by pitch class
   - Rotate minor profile by pitch class
   - Compute Pearson correlation
5. Select key with highest correlation
6. Output: pitchClass, isMinor, confidence
7. Convert to Camelot (e.g., "8A") and Open Key (e.g., "1m") notation
```

#### Beatgrid Detection Algorithm

```
1. Compute STFT spectrogram (2048-point FFT, 512-sample hop)
2. Calculate spectral flux (half-wave rectified difference)
3. Autocorrelation-based tempo estimation:
   - Compute autocorrelation of onset strength signal
   - Find peak in 60-180 BPM range
4. Peak picking with adaptive threshold:
   - Mean + 0.5 * standard deviation
   - Minimum distance constraint (half beat interval)
5. Grid alignment from first strong onset
6. Output: beat markers with time, index, downbeat flag
```

### Go Engine Architecture

The Go engine handles coordination, storage, and exports.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          Go Engine (1.24+)                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────┐    ┌─────────────────┐    ┌──────────────────┐    │
│  │   HTTP API      │    │   gRPC Server   │    │  Library Scanner │    │
│  │ /api/tracks     │    │ EngineService   │    │ (recursive, hash) │    │
│  │ /api/export     │    │ AnalyzerWorker  │    └──────────────────┘    │
│  │ /api/propose-set│    │                 │                             │
│  └─────────────────┘    └─────────────────┘                             │
│           │                     │                                        │
│           ▼                     ▼                                        │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                       Set Planner                                │   │
│  │  - Weighted graph with edge scoring                             │   │
│  │  - Key compatibility (Camelot wheel distance)                    │   │
│  │  - Tempo delta penalty                                           │   │
│  │  - Energy flow scoring (climb/drop preferences)                  │   │
│  │  - Transition window overlap detection                           │   │
│  │  - Set modes: Warm-up, Peak-time, Open-format                   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                     │
│                                    ▼                                     │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                        Exporters                                 │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │   │
│  │  │ Generic  │ │ Rekordbox│ │  Serato  │ │ Traktor  │           │   │
│  │  │ M3U/JSON │ │   XML    │ │  .crate  │ │   NML    │           │   │
│  │  │ CSV/Tar  │ │DJ_PLAYLIST│ │ vrsn fmt │ │  v19     │           │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Web UI Architecture

React 19 + TypeScript with Zustand state management.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                      Web UI (React 19 + TypeScript)                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Zustand Store (store.ts)                      │   │
│  │  - tracks, trackMap, filteredTracks                             │   │
│  │  - selectedId, viewMode, chartMode                              │   │
│  │  - currentSetPlan with edges                                    │   │
│  │  - API actions: fetchTracks, proposeSet, exportSet             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                     │
│           ┌────────────────────────┼────────────────────────┐           │
│           ▼                        ▼                        ▼           │
│  ┌────────────────┐     ┌────────────────┐     ┌────────────────┐      │
│  │  Library View  │     │  Set Builder   │     │   Graph View   │      │
│  │  - LibraryGrid │     │  - SetBuilder  │     │ - TransitionGr │      │
│  │  - TrackDetail │     │  - EnergyArc   │     │ - D3 force sim │      │
│  │  - BPMKeyChart │     │  - ExportPanel │     │ - zoom/pan     │      │
│  └────────────────┘     └────────────────┘     └────────────────┘      │
│                                                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Visualization Components                      │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐             │   │
│  │  │WaveformCanvas│ │SpectrumAnalz│ │  EnergyArc   │             │   │
│  │  │ Canvas 2D    │ │ Canvas 2D   │ │ SVG + Framer │             │   │
│  │  │ sections     │ │ mirror/bars │ │ bezier curve │             │   │
│  │  │ cues, beats  │ │ 60fps RAF   │ │ spring anim  │             │   │
│  │  └──────────────┘ └──────────────┘ └──────────────┘             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Audio   │────▶│  Swift   │────▶│   Go     │────▶│   Web    │
│  Files   │     │ Analyzer │     │  Engine  │     │   UI     │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
     │                │                 │                │
     │                │                 │                │
     ▼                ▼                 ▼                ▼
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│WAV/MP3/  │     │TrackAnalz│     │ SQLite   │     │ Zustand  │
│AAC/FLAC  │     │Result    │     │ + Blobs  │     │  Store   │
└──────────┘     └──────────┘     └──────────┘     └──────────┘

1. Library Scanner finds audio files, computes SHA256 content hash
2. Swift Analyzer decodes audio, runs DSP pipeline
3. Analysis results stored in SQLite with WAL mode
4. Web UI fetches via HTTP API, stores in Zustand
5. User builds set, UI calls proposeSet API
6. Planner returns ordered tracks with edge explanations
7. Export generates Rekordbox/Serato/Traktor files
```

### Protocol Buffers

Shared contracts between Go engine and Swift analyzer:

```protobuf
// proto/common/types.proto
message TrackAnalysis {
  TrackId id = 1;
  double duration_seconds = 2;
  Beatgrid beatgrid = 3;
  MusicalKey key = 4;
  int32 energy_global = 5;
  repeated Section sections = 7;
  repeated CuePoint cue_points = 8;
  repeated TransitionWindow transition_windows = 9;
}

message MusicalKey {
  string value = 1;      // "8A" (Camelot)
  KeyFormat format = 2;  // CAMELOT or OPEN_KEY
  float confidence = 3;
}

message Beatgrid {
  repeated BeatMarker beats = 1;
  repeated TempoMapNode tempo_map = 2;
  float confidence = 3;
}
```

## Screenshots

| Library View | Set Builder | Graph View |
|:---:|:---:|:---:|
| ![Library](docs/assets/screens/algiers-library-view.png) | ![Set Builder](docs/assets/screens/algiers-set-builder.png) | ![Graph](docs/assets/screens/algiers-graph-view.png) |

<details>
<summary>More screenshots</summary>

| Active Waveform | Light Mode |
|:---:|:---:|
| ![Hero](docs/assets/screens/algiers-hero.png) | ![Light](docs/assets/screens/algiers-light-mode.png) |

</details>

## Quick Start

```bash
# Clone the repository
git clone https://github.com/cartomix/algiers.git
cd algiers

# Install Go dependencies
go mod download

# Install web dependencies
cd web && npm install && cd ..

# Generate proto stubs (if modified)
make proto

# Run tests
make test

# Start the engine (port 50051)
go run ./cmd/engine

# Start the web UI (separate terminal)
cd web && npm run dev
```

Open http://localhost:5173 to see the UI.

## Developer Loop

| Command | Description |
|---------|-------------|
| `make test` | Run Go + Swift unit tests |
| `make lint` | Run golangci-lint, swift-format, eslint |
| `make proto` | Regenerate protobuf stubs |
| `make fixturegen` | Generate test audio fixtures |
| `make build` | Build engine + analyzer binaries |
| `cd web && npm run dev` | Start Vite dev server |
| `cd web && npm run build` | Production build |
| `make screenshots-install` | Install Playwright browsers |
| `make screenshots` | Capture UI screenshots (headless) |
| `make screenshots-headed` | Capture UI screenshots (visible browser) |

### Test Fixtures

```bash
make fixturegen
```

Generates deterministic WAV files under `testdata/audio/`:
- BPM ladder (80–160 BPM click tracks)
- Swing click (52-65% swing ratios)
- Tempo ramp (128→100 BPM linear)
- Harmonic pad (chord in known keys)
- `manifest.json` for golden comparisons

### Export Verification

```bash
go run ./cmd/exportverify --manifest <path>/checksums.txt
```

Validates SHA256 checksums for exported bundles.

### Screenshot & GIF Capture

Automated screenshot and GIF capture using Playwright-Go for README assets.

```bash
# One-time: Install Playwright browsers + ffmpeg
make screenshots-install
brew install ffmpeg  # for GIF conversion

# Start the web UI (in a separate terminal)
make run-web

# Capture screenshots + animated GIF
make screenshots

# Or watch the browser (useful for debugging)
make screenshots-headed
```

**Output files:**
- `algiers-demo.gif` — Animated hero for README
- `algiers-library-view.png` — Library view
- `algiers-set-builder.png` — Set Builder
- `algiers-graph-view.png` — Transition graph
- `algiers-light-mode.png` — Light mode

Screenshots at 1280x720 @2x (retina). GIF at 640px width, 12fps.

## Project Layout

```
.
├── cmd/
│   ├── engine/          # Go gRPC server entrypoint
│   ├── exportverify/    # Checksum verification CLI
│   ├── fixturegen/      # Test audio generator
│   └── screenshots/     # Playwright-Go screenshot capture
├── analyzer-swift/      # Swift analyzer module (Metal + Core ML)
│   ├── Package.swift
│   └── Sources/
├── internal/
│   ├── analyzer/        # Analyzer interface + gRPC client
│   ├── auth/            # gRPC auth interceptors
│   ├── config/          # Server configuration
│   ├── exporter/        # M3U/JSON/CSV/Tar exports
│   ├── fixtures/        # WAV generator
│   ├── planner/         # Set ordering algorithm
│   ├── scanner/         # Library file scanner
│   ├── server/          # gRPC server implementation
│   └── storage/         # SQLite + blob store
├── web/
│   ├── src/
│   │   ├── components/  # React components
│   │   │   ├── WaveformCanvas.tsx
│   │   │   ├── SpectrumAnalyzer.tsx
│   │   │   ├── EnergyArc.tsx
│   │   │   ├── TransitionGraph.tsx
│   │   │   ├── LiveStats.tsx
│   │   │   ├── BPMKeyChart.tsx
│   │   │   └── ...
│   │   ├── App.tsx
│   │   └── types.ts
│   └── package.json
├── proto/               # gRPC/Protobuf contracts
├── gen/go/              # Generated Go stubs
├── docs/
│   ├── PLAN.md          # Milestones and tasks
│   └── ROADMAP.md       # Ops hardening checklist
├── testdata/audio/      # Generated test fixtures
└── Makefile
```

## Roadmap

### Alpha (Complete)
- [x] gRPC engine with health checks
- [x] Library scanner with content hashing
- [x] SQLite storage with migrations
- [x] Set planner with weighted graph
- [x] Generic exports (M3U/JSON/CSV/Tar)
- [x] Pro UI with visualizations
- [x] Dark mode default

### Beta (Current)
- [x] Swift analyzer with Accelerate DSP
- [x] Beatgrid detection (onset + autocorrelation)
- [x] Key detection with Camelot/Open Key mapping
- [x] Energy analysis with band breakdown
- [x] Section detection (intro/verse/build/drop/breakdown/outro)
- [x] Cue generation (priority-based, beat-aligned)
- [x] HTTP REST API bridge for web UI
- [x] Zustand state management with API integration
- [x] Export panel in Set Builder UI
- [x] Web Audio playback hook
- [ ] Core ML integration for ANE inference
- [ ] gRPC integration for Swift analyzer
- [ ] Embeddings/similarity for vibe continuity

### v1.0
- [x] Rekordbox XML export
- [x] Serato crate export
- [x] Traktor NML export
- [x] Playwright-Go E2E tests
- [ ] Alpha acceptance: 100 tracks → 30-track set → export

## Contributing

PRs welcome! Keep commits scoped and include:
- Repro steps for bugs
- Expected vs actual behavior
- Test coverage for new features

## Changelog

### v0.2.0-beta (2026-01-29)
- **Swift Analyzer with Accelerate DSP** (full implementation)
  - AudioDecoder using AVFoundation with sample rate conversion
  - FFTProcessor with vDSP: STFT, spectral flux, chroma features
  - BeatgridDetector: onset detection + autocorrelation tempo estimation
  - KeyDetector: Krumhansl-Schmuckler profiles with Camelot/Open Key output
  - EnergyAnalyzer: RMS/peak, band energy (low/mid/high), energy curve
  - SectionDetector: energy-based segmentation (intro/verse/build/drop/breakdown/outro)
  - CueGenerator: priority-based selection with beat alignment (max 8 cues)
  - Main Analyzer orchestrator with async progress callbacks
  - CLI tool (`analyzer-swift analyze/serve/healthcheck`)
  - HTTP API server for analyzer integration
- Web Audio playback hook (useAudioPlayer) with play/pause/seek
- Golden comparison tests for all export formats
- Detailed architecture documentation in README

### v0.1.1-alpha (2026-01-29)
- HTTP REST API bridge for web UI integration
- Zustand state management with API fallback to demo mode
- Export panel in Set Builder with vendor format selection
- Rekordbox XML export (DJ_PLAYLISTS schema)
- Serato crate export (.crate binary format)
- Traktor NML export (v19 schema with CUE_V2)
- Playwright-Go E2E test framework
- Property tests for planner algorithms

### v0.1.0-alpha (2026-01-29)
- Initial alpha release
- Pro UI with D3.js visualizations
- Canvas waveform renderer
- Real-time spectrum analyzer
- Energy arc and transition graph
- Live stats dashboard
- 16 demo tracks with realistic analysis
- Dark mode default

## License

Blue Oak Model License 1.0.0. See [LICENSE](LICENSE).

---

<div align="center">

**Built for DJs who want to prep smarter, not harder.**

Made with Metal, Core ML, and too much coffee.

</div>
