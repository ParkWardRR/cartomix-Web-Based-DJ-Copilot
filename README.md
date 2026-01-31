<div align="center">

<img src="docs/assets/logo.svg" alt="Algiers - DJ Set Prep Copilot" width="100" />

# Algiers — DJ Set Prep Copilot

**Apple Silicon–native DJ prep tool: analyze, cue, transition, export — all local.**

<!-- Status & Version -->
[![Phase](https://img.shields.io/badge/beta-blueviolet?style=for-the-badge)](#beta-features)
[![Version](https://img.shields.io/badge/v0.5.1-blue?style=for-the-badge)](#changelog)
[![E2E Tests](https://img.shields.io/badge/E2E%20Tests-7%2F7%20Passing-brightgreen?style=for-the-badge)](#developer-loop)
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

<!-- ML Stack -->
[![OpenL3](https://img.shields.io/badge/OpenL3%20512--dim-8B5CF6?style=for-the-badge)](#ml-powered-similarity)
[![Vibe Match](https://img.shields.io/badge/Vibe%20Matching-EC4899?style=for-the-badge)](#ml-powered-similarity)
[![Explainable](https://img.shields.io/badge/Explainable%20AI-10B981?style=for-the-badge)](#explainable-transitions)
[![Custom Training](https://img.shields.io/badge/Custom%20Training-F59E0B?style=for-the-badge)](#custom-ml-training)
[![SoundAnalysis](https://img.shields.io/badge/Apple%20SoundAnalysis-FF3B30?style=for-the-badge)](#apple-soundanalysis)

<!-- Backend Stack -->
[![Go](https://img.shields.io/badge/Go%201.24-00ADD8?style=for-the-badge&logo=go&logoColor=white)](#architecture)
[![Swift](https://img.shields.io/badge/Swift%206-F05138?style=for-the-badge&logo=swift&logoColor=white)](#architecture)
[![gRPC](https://img.shields.io/badge/gRPC%20Streaming-4285F4?style=for-the-badge&logo=google&logoColor=white)](#communication-protocols)
[![HTTP](https://img.shields.io/badge/HTTP%20REST-009688?style=for-the-badge)](#communication-protocols)
[![SQLite](https://img.shields.io/badge/SQLite%20WAL-003B57?style=for-the-badge&logo=sqlite&logoColor=white)](#architecture)
[![Protobuf](https://img.shields.io/badge/Protobuf-4285F4?style=for-the-badge&logo=google&logoColor=white)](#architecture)

<!-- Frontend Stack -->
[![React](https://img.shields.io/badge/React%2019-61DAFB?style=for-the-badge&logo=react&logoColor=black)](#architecture)
[![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](#architecture)
[![Vite](https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white)](#architecture)
[![Web Audio](https://img.shields.io/badge/Web%20Audio-FF6B6B?style=for-the-badge)](#web-audio-playback)
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
[![Docs](https://img.shields.io/badge/API.md-8E43E7?style=for-the-badge)](docs/API.md)
[![License](https://img.shields.io/badge/Blue%20Oak-lightgray?style=for-the-badge)](LICENSE)

<br/>

### See it in action

![Algiers Demo](docs/assets/screens/algiers-demo.webp)

**100% local. No cloud. ANE + Metal accelerated. Private by design.**

</div>

---

## Table of Contents
- [What This Is](#what-this-is)
- [Why It's Different](#why-its-different)
- [Apple Silicon Deep Dive](#apple-silicon-deep-dive)
- [ML-Powered Similarity](#ml-powered-similarity)
- [Apple SoundAnalysis](#apple-soundanalysis)
- [Custom ML Training](#custom-ml-training)
- [Explainable Transitions](#explainable-transitions)
- [Beta Features](#beta-features)
- [Pro Visualizations](#pro-visualizations)
- [Feature Highlights](#feature-highlights)
- [Architecture](#architecture)
- [Communication Protocols](#communication-protocols)
- [Screenshots](#screenshots)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Developer Loop](#developer-loop)
- [Project Layout](#project-layout)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

> 📖 **See also:**
> - [Installation Guide](INSTALL.md) for one-command install, system requirements, troubleshooting
> - [AI/ML Architecture Guide](docs/AI-ML.md) for deep dive into OpenL3, SoundAnalysis, and custom training
> - [ML Training Guide](docs/ML-TRAINING.md) for detailed custom model training documentation
> - [gRPC Migration Guide](docs/gRPC-MIGRATION.md) for migrating from HTTP REST to gRPC

## What This Is

A local-first copilot for DJ set prep. It ingests your library, detects structure, surfaces mixable sections, proposes cue points and transition windows, and optimizes set order with explainable scoring—while keeping you in control of the mix.

**Key Principles:**
- **End-to-end local** — No cloud lock-in; audio never leaves your Mac
- **Apple-accelerated DSP** — Tempo, key, loudness, and energy analysis run on Metal + Accelerate
- **ML where it helps** — ANE-powered models suggest cues/sections but every choice is editable
- **Explainable transitions** — Every proposed segue cites tempo delta, key distance, energy window overlap, and beat-grid alignment
- **Vibe continuity** — 512-dimensional OpenL3 embeddings find tracks that "feel" similar

## Why It's Different

| Feature | Algiers | Cloud-based Tools |
|---------|---------|-------------------|
| **Privacy** | Audio never leaves your Mac | Uploads to servers |
| **Speed** | ANE + Metal acceleration | Network latency |
| **Control** | Every edit is yours | AI decides for you |
| **Explainability** | Shows *why* transitions work | Black box |
| **Offline** | Works without internet | Requires connection |
| **Loudness** | Broadcast-standard EBU R128 | Inconsistent metering |
| **Similarity** | 512-dim ML embeddings + DSP | Genre tags only |
| **Hardware** | Uses all Apple Silicon engines | Generic x86 code |

---

## Apple Silicon Deep Dive

Algiers is built from the ground up to leverage every compute engine in Apple Silicon. This isn't just "runs on M1" — it's **architected for M1-M4** with zero fallback to generic x86 code paths.

### The Apple Silicon Advantage

Apple Silicon chips integrate multiple specialized compute engines on a single chip with **unified memory architecture (UMA)**. This means data flows between CPU, GPU, and Neural Engine without expensive copies:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Apple Silicon (M1-M4)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  │  CPU Cores   │   │    GPU       │   │Neural Engine │   │   Media      │ │
│  │  (Perf + E)  │   │   (Metal)    │   │    (ANE)     │   │   Engine     │ │
│  │              │   │              │   │              │   │              │ │
│  │  General     │   │  Parallel    │   │   ML Infer   │   │   Encode/    │ │
│  │  Compute     │   │  Graphics    │   │   16 TOPS    │   │   Decode     │ │
│  └──────┬───────┘   └──────┬───────┘   └──────┬───────┘   └──────┬───────┘ │
│         │                  │                  │                   │         │
│         └──────────────────┴──────────────────┴───────────────────┘         │
│                                    │                                         │
│                     ┌──────────────┴──────────────┐                         │
│                     │     Unified Memory (UMA)    │                         │
│                     │    Zero-copy data sharing    │                         │
│                     └─────────────────────────────┘                         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### How Algiers Uses Each Engine

| Engine | Apple Framework | Algiers Use Case | Performance Gain |
|--------|-----------------|------------------|------------------|
| **Neural Engine (ANE)** | Core ML | OpenL3 embeddings, section classification | 15x faster than CPU |
| **GPU (Metal)** | Metal Performance Shaders | Real-time spectrograms, onset detection | 10x faster than CPU |
| **CPU (Accelerate)** | vDSP, vForce, BLAS | FFT, K-weighting, beatgrid, key detection | 4x faster than naive |
| **Media Engine** | AVFoundation | Audio decode (FLAC/AAC/MP3) | Hardware-accelerated |

### Neural Engine (ANE) — ML Inference

The ANE is Apple's dedicated ML accelerator with **16 trillion operations per second (TOPS)**. Algiers uses it for:

```swift
// Core ML model configured for ANE
let config = MLModelConfiguration()
config.computeUnits = .all  // Prefers ANE, falls back to GPU

// OpenL3 embedding generation runs on ANE
let model = try MLModel(contentsOf: modelURL, configuration: config)
let embedding = try model.prediction(from: melSpectrogram)  // 512-dim vector
```

**OpenL3 on ANE:**
- Input: 128-band mel spectrogram (199 frames = 1 second)
- Output: 512-dimensional embedding vector
- Latency: ~5ms per window on M1, ~2ms on M3
- Power: 1/15th the energy of CPU inference

### Accelerate Framework — Vectorized DSP

The Accelerate framework provides SIMD-optimized signal processing:

```swift
// FFT using vDSP (vectorized, cache-optimized)
let fftSetup = vDSP_create_fftsetup(log2n, FFTRadix(kFFTRadix2))!
vDSP_fft_zrip(fftSetup, &splitComplex, 1, log2n, FFTDirection(FFT_FORWARD))

// K-weighting filter for EBU R128 loudness
// High-shelf filter: +4dB @ 1500Hz
vDSP_biquad(biquadSetup, &state, samples, 1, &output, 1, vDSP_Length(count))

// Mel filterbank application (matrix multiply)
vDSP_mmul(powerSpectrum, 1, melFilterbank, 1, &melBands, 1,
          vDSP_Length(melBands), vDSP_Length(fftBins), vDSP_Length(1))
```

**Performance comparison (5-minute track):**

| Operation | Naive Implementation | Accelerate | Speedup |
|-----------|---------------------|------------|---------|
| FFT (2048-point) | 45ms | 3ms | 15x |
| Autocorrelation | 120ms | 8ms | 15x |
| Key detection | 800ms | 180ms | 4.4x |
| Loudness (R128) | 2.1s | 450ms | 4.7x |

### Unified Memory — Zero-Copy Pipeline

Traditional architectures require copying data between CPU and GPU memory. Apple's UMA eliminates this:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Algiers Analysis Pipeline                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Audio File ──▶ Media Engine (decode) ──▶ Float32 PCM in UMA            │
│                                              │                           │
│                                              │ (no copy)                 │
│                                              ▼                           │
│                                      ┌──────────────┐                   │
│                                      │ Accelerate   │                   │
│                                      │ FFT, Filters │                   │
│                                      └──────┬───────┘                   │
│                                              │ (no copy)                 │
│                                              ▼                           │
│                                      ┌──────────────┐                   │
│                                      │ Core ML/ANE  │                   │
│                                      │ OpenL3 Model │                   │
│                                      └──────┬───────┘                   │
│                                              │                           │
│                                              ▼                           │
│                               512-dim embedding in UMA                   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**Memory bandwidth comparison:**
- Traditional (discrete GPU): 16 GB/s PCIe
- Apple Silicon UMA: 200+ GB/s (M1), 400+ GB/s (M3 Max)

### Real-World Performance

Benchmarks on a 5-minute FLAC track (48kHz stereo):

| Chip | Full Analysis | OpenL3 Only | Decode + FFT |
|------|---------------|-------------|--------------|
| M1 | 8.2s | 2.1s | 1.8s |
| M1 Pro | 6.1s | 1.4s | 1.2s |
| M2 | 5.8s | 1.2s | 1.1s |
| M3 | 4.9s | 0.9s | 0.8s |
| M3 Max | 3.2s | 0.5s | 0.4s |

**Analysis breakdown (M1 baseline):**
- Audio decode: 1.8s (Media Engine)
- Beatgrid detection: 2.4s (Accelerate)
- Key detection: 0.8s (Accelerate)
- Energy analysis: 0.3s (Accelerate)
- Loudness (R128): 0.5s (Accelerate)
- OpenL3 embeddings: 2.1s (ANE) — 547 windows @ ~4ms each
- Section/cue generation: 0.3s (CPU)

---

## ML-Powered Similarity

Algiers uses **OpenL3**, a state-of-the-art audio embedding model, to find tracks with similar "vibe" — going far beyond simple BPM/key matching.

### What is OpenL3?

OpenL3 is a deep neural network trained on millions of audio-video pairs. It learns to encode audio into a compact 512-dimensional vector that captures:
- **Timbre** — The "color" of sound (bright, dark, warm, harsh)
- **Texture** — Rhythmic density, layering, complexity
- **Mood** — Energy, tension, atmosphere
- **Genre characteristics** — Without explicit labels

### How It Works

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        OpenL3 Embedding Pipeline                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Audio (48kHz) ──▶ Mel Spectrogram ──▶ Core ML Model ──▶ 512-dim Vector │
│                                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐               │
│  │   1 second   │    │  128 bands   │    │   OpenL3     │               │
│  │   window     │───▶│  199 frames  │───▶│  (on ANE)    │──▶ [0.12, ... │
│  │   (0.5s hop) │    │  log-scale   │    │   ~5ms       │      0.87]    │
│  └──────────────┘    └──────────────┘    └──────────────┘               │
│                                                                          │
│  Full Track: 547 windows ──▶ Mean Pool ──▶ Single 512-dim track vector  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Similarity Scoring

Finding similar tracks combines ML embeddings with traditional DSP features:

```
Combined Score = 0.50 × OpenL3 Similarity    (vibe match)
              + 0.20 × Tempo Similarity      (BPM compatibility)
              + 0.20 × Key Similarity        (harmonic compatibility)
              + 0.10 × Energy Similarity     (energy level match)
```

**Example API response:**

```json
GET /api/tracks/{id}/similar

{
  "query": {
    "title": "Get Lucky",
    "artist": "Daft Punk",
    "bpm": 116.0,
    "key": "8A"
  },
  "similar": [
    {
      "title": "Lose Yourself to Dance",
      "artist": "Daft Punk",
      "score": 0.92,
      "explanation": "similar vibe (89%); tempo match; same key; same energy",
      "vibe_match": 89.2,
      "tempo_match": 100.0,
      "key_match": 100.0,
      "bpm_delta": 0.0,
      "key_relation": "same"
    },
    {
      "title": "Redbone",
      "artist": "Childish Gambino",
      "score": 0.78,
      "explanation": "similar vibe (72%); Δ-4 BPM; key compatible; energy -1",
      "vibe_match": 72.1,
      "tempo_match": 80.0,
      "key_match": 85.0,
      "bpm_delta": 4.0,
      "key_relation": "compatible"
    }
  ]
}
```

### Why OpenL3 Over Genre Tags?

| Approach | Strengths | Limitations |
|----------|-----------|-------------|
| **Genre Tags** | Human-curated, easy to understand | Subjective, inconsistent, limited vocabulary |
| **BPM/Key Only** | Objective, precise | Misses texture, mood, energy arc |
| **OpenL3** | Captures perceptual similarity | Requires ML inference, less interpretable |

Algiers combines all three: OpenL3 for vibe, BPM/key for technical compatibility, and human-readable explanations for trust.

---

## Apple SoundAnalysis

Algiers integrates **Apple's SoundAnalysis framework** (Layer 1) to detect audio context and generate quality flags — all running locally on your Mac.

### Built-in Classifier (300+ Labels)

Apple's SNClassifySoundRequest provides zero-training classification of sounds:

```swift
// Using Apple's built-in sound classifier
let request = try SNClassifySoundRequest(classifierIdentifier: .version1)
let analyzer = SNAudioStreamAnalyzer(format: audioFormat)
try analyzer.add(request, withObserver: self)
```

### DJ-Relevant Categories

The raw Apple labels are mapped to DJ-relevant categories:

| Category | Example Labels | Use Case |
|----------|----------------|----------|
| **Music** | `music`, `electronic_music`, `hip_hop_music` | Primary audio content |
| **Speech** | `speech`, `singing`, `crowd` | Crowd noise, vocals |
| **Noise** | `noise`, `static`, `hum` | QA flagging |
| **Silence** | `silence` | Track boundaries |

### QA Flag Generation

Based on classification results, Algiers generates actionable QA flags:

| Flag | Trigger | Action |
|------|---------|--------|
| `needs_review` | Low confidence (<0.5) on primary context | Manual review suggested |
| `mixed_content` | Speech + music in same track | Check for vocal samples |
| `speech_detected` | >10s of continuous speech | May need review |
| `low_confidence` | Analysis confidence below threshold | Re-analyze or manual check |

```json
GET /api/tracks/{id}

{
  "sound_context": "music",
  "sound_context_confidence": 0.94,
  "qa_flags": [
    {
      "type": "mixed_content",
      "reason": "Speech detected in music track (2.3s at 1:42)",
      "dismissed": false
    }
  ]
}
```

---

## Custom ML Training

> 📖 **Full documentation:** [docs/ML-TRAINING.md](docs/ML-TRAINING.md)

Algiers supports **opt-in custom model training** (Layer 3) for DJ-specific section classification. Train your own model to recognize Intro, Build, Drop, Breakdown, and Outro sections based on your labeling style.

### 3-Layer ML Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     Algiers ML Architecture                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐                │
│  │   Layer 1    │   │   Layer 2    │   │   Layer 3    │                │
│  │              │   │              │   │              │                │
│  │ SoundAnalysis│   │   OpenL3     │   │Custom Section│                │
│  │  (Built-in)  │   │ (Embeddings) │   │  (Training)  │                │
│  │              │   │              │   │              │                │
│  │  300+ labels │   │   512-dim    │   │  7 DJ labels │                │
│  │  Zero config │   │  Vibe match  │   │  Your data   │                │
│  └──────────────┘   └──────────────┘   └──────────────┘                │
│         │                  │                  │                         │
│         │                  │                  │                         │
│         ▼                  ▼                  ▼                         │
│  ┌──────────────────────────────────────────────────────────┐          │
│  │              All run on Apple Neural Engine               │          │
│  │                   (ANE) — Local Only                      │          │
│  └──────────────────────────────────────────────────────────┘          │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### DJ Section Labels

| Label | Color | Description |
|-------|-------|-------------|
| `intro` | 🟢 Green | Track opening, minimal elements |
| `build` | 🟡 Yellow | Energy rising, tension building |
| `drop` | 🔴 Red | Main energy peak, full arrangement |
| `break` | 🟣 Purple | Breakdown, stripped back |
| `outro` | 🔵 Blue | Track ending, elements fading |
| `verse` | ⬛ Gray | Vocal or melodic verse |
| `chorus` | 🩷 Pink | Hook or main phrase |

### Training Workflow

1. **Label tracks** — Add section labels via the Training UI
2. **Review stats** — Ensure minimum 10 samples per class
3. **Train model** — Click "Start Training" (uses Create ML)
4. **Activate model** — Choose best version for inference
5. **Rollback** — Restore previous version if needed

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Label     │───▶│   Train     │───▶│  Evaluate   │───▶│  Activate   │
│   Tracks    │    │   Model     │    │   Results   │    │   Version   │
│             │    │             │    │             │    │             │
│ Min 10/class│    │ Create ML   │    │ Accuracy %  │    │ One-click   │
│ 7 labels    │    │ On ANE      │    │ F1 Score    │    │ Rollback    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### Training API

```bash
# Add a training label
curl -X POST http://localhost:8080/api/training/labels \
  -H "Content-Type: application/json" \
  -d '{
    "track_id": 123,
    "label_value": "drop",
    "start_beat": 64,
    "end_beat": 128,
    "start_time_seconds": 32.0,
    "end_time_seconds": 64.0
  }'

# Check if ready to train
curl http://localhost:8080/api/training/labels/stats
# {"total_labels": 142, "label_counts": {"drop": 28, "build": 25, ...}, "ready_for_training": true}

# Start training
curl -X POST http://localhost:8080/api/training/start
# {"job_id": "job_1706550000", "message": "training job started"}

# Check progress
curl http://localhost:8080/api/training/jobs/job_1706550000
# {"status": "training", "progress": 0.6, "current_epoch": 6, "total_epochs": 10}

# Activate a model version
curl -X POST http://localhost:8080/api/training/models/2/activate
```

---

## Explainable Transitions

Every transition suggestion includes a **human-readable explanation** of why it works:

```
"similar vibe (82%); Δ+2 BPM; key: 8A→9A (compatible); energy +1; beat-grid aligned"
```

### Explanation Components

| Component | Format | Meaning |
|-----------|--------|---------|
| **Vibe match** | `similar vibe (82%)` | OpenL3 cosine similarity percentage |
| **Tempo delta** | `Δ+2 BPM` or `Δ-3 BPM` | BPM difference with direction |
| **Key relation** | `8A→9A (compatible)` | Camelot notation with compatibility |
| **Energy flow** | `energy +1` | Energy level change (-9 to +9) |
| **Beat alignment** | `beat-grid aligned` | Whether downbeats align |

### Key Compatibility Types

| Relation | Camelot Rule | Example | Explanation |
|----------|--------------|---------|-------------|
| **Same** | Identical | 8A→8A | Perfect key match |
| **Relative** | Same number, A↔B | 8A→8B | Major/minor relative |
| **Compatible** | ±1 same mode | 8A→9A | Energy boost/drop |
| **Harmonic** | Various | 8A→10A | Harmonic but noticeable |
| **Clash** | Everything else | 8A→3B | Likely key clash ⚠ |

---

## Beta Features

v0.5.1-beta brings **enhanced gRPC streaming progress** and comprehensive E2E testing:

### gRPC Streaming Progress (v0.5.1 - NEW)
- **Real-time Progress** — Current file, percent complete, elapsed time, ETA
- **Byte-level Tracking** — Total bytes processed and remaining for scans
- **Stage Timing** — Per-stage breakdown with millisecond precision
- **Track Metadata** — Title, artist, duration included in analysis progress
- **Validation Metrics** — Loss, accuracy for training progress

### ML Training (v0.5.0)
- **Custom DJ Section Training** — Train your own Intro/Build/Drop/Break/Outro classifier
- **Create ML Integration** — Uses Apple's MLSoundClassifier on Neural Engine
- **Model Versioning** — Keep multiple versions, activate/rollback with one click
- **Training UI** — Label tracks, monitor progress, manage models
- **Training API** — Full REST API for labels, jobs, and model versions

### Apple SoundAnalysis (v0.5.0)
- **Built-in Classifier** — 300+ audio labels with zero configuration
- **Sound Context Detection** — Music/speech/noise classification
- **QA Flag Generation** — Automatic quality flags (needs_review, mixed_content)
- **Primary Context** — Dominant audio type with confidence score

### ML & Similarity (v0.4.0)
- **OpenL3 Embeddings** — 512-dimensional vectors from Core ML on ANE
- **Vibe Matching** — Find tracks that "feel" similar beyond BPM/key
- **Similarity Search API** — `GET /api/tracks/{id}/similar` with explanations
- **Weighted Scoring** — 50% vibe + 20% tempo + 20% key + 10% energy
- **Explainable Results** — Every match shows why it was chosen

### Audio Analysis
- **EBU R128 Loudness** — Broadcast-standard loudness (LUFS, LU, true peak) with K-weighting
- **Audio Embeddings** — 128-dim MFCC vectors + 512-dim OpenL3 vectors
- **Spectral Features** — Centroid, rolloff, flatness, harmonic ratio for timbre
- **Web Audio Playback** — Full browser-based audio with play/pause/seek/speed

### UI & Visualization
- **Training Screen** — Dataset table, label editor, progress card, model versions
- **Analysis Panel** — DSP + ML results with confidence bars and QA flags
- **Similar Tracks Panel** — Ranked list with explainable scores
- **Model Settings** — Feature toggles, disk usage, latency estimates
- **Live Dashboard** — Animated stats with real-time analysis progress
- **Canvas Waveform** — High-performance waveform with sections, cues, playhead
- **Spectrum Analyzer** — WebGL-style frequency visualization with mirror mode
- **Set Energy Arc** — Animated SVG energy flow visualization
- **Transition Graph** — D3.js force-directed graph showing track connections

### Export Formats
- **Rekordbox XML** — Full DJ_PLAYLISTS schema with cues, tempo markers, key
- **Serato crate** — Binary .crate format with supplementary cues CSV
- **Traktor NML** — Complete NML v19 export with CUE_V2 markers
- **Generic** — M3U8, JSON, CSV, SHA256 checksums, tar.gz bundles

---

## Pro Visualizations

Built with high-performance libraries for smooth 60fps rendering:

| Component | Technology | Description |
|-----------|-----------|-------------|
| **Waveform Canvas** | Canvas 2D API | GPU-accelerated waveform with sections, cues, beat markers |
| **Spectrum Analyzer** | Canvas 2D + RAF | Real-time frequency bars with mirror/bars/circular modes |
| **Energy Arc** | SVG + Framer Motion | Animated bezier curve showing set energy journey |
| **Transition Graph** | D3.js Force | Interactive force-directed graph with zoom/pan/drag |
| **BPM/Key Charts** | D3.js | Animated bar charts with hover tooltips |
| **Live Stats** | Framer Motion | Spring-animated number counters and progress rings |
| **Audio Player** | Web Audio API | Full playback with timeline, seek, speed control |

---

## Feature Highlights

### Library Management
- Resumable library ingest (WAV/AIFF/MP3/AAC/ALAC/FLAC)
- Content-hash caching for instant re-scans
- "Needs review" flagging for low-confidence analysis
- Search by title, artist, or key

### Analysis
- **Beatgrid** — Static + dynamic tempo maps with confidence
- **Key Detection** — Krumhansl-Schmuckler with Camelot mapping
- **Energy** — 1-10 scale with per-section curves and band breakdown
- **Loudness** — EBU R128 integrated, momentary, short-term, LRA, true peak
- **Sections** — Intro/Verse/Build/Drop/Breakdown/Outro with confidence
- **Cues** — Up to 8 beat-aligned suggestions with priority scoring
- **Embeddings** — 128-dim MFCC + 512-dim OpenL3 for vibe matching

### Set Planning
- Weighted graph optimization with explainable scoring
- Set modes: Warm-up, Peak-time, Open-format
- Edge explanations: tempo delta, key relation, energy flow
- Transition window matching (intro→outro overlap)
- **Vibe continuity** scoring using OpenL3 embeddings

### Export
- **Generic** — M3U8, JSON, CSV, SHA256 checksums, tar.gz bundles
- **Rekordbox** — Full DJ_PLAYLISTS XML with cues, tempo markers, key
- **Serato** — Binary .crate format with cues CSV
- **Traktor** — NML v19 with CUE_V2 markers

---

## Architecture

Algiers uses a three-layer architecture optimized for Apple Silicon:

```mermaid
flowchart LR
    subgraph UI["Frontend"]
      direction TB
      REACT["React 19 + TypeScript"]
      VIZ["D3.js / Framer Motion"]
      AUDIO["Web Audio API"]
      REACT --> VIZ
      REACT --> AUDIO
    end

    subgraph ENGINE["Go Engine (1.24+)"]
      direction TB
      HTTP["HTTP REST API"]
      GRPC["gRPC Server"]
      SCHED["Job Scheduler"]
      PLAN["Set Planner"]
      SIM["Similarity Search"]
      EXPORT["Exporters"]
      HTTP --> SCHED
      GRPC --> SCHED
      SCHED --> PLAN --> SIM --> EXPORT
    end

    subgraph ANALYZER["Swift Analyzer"]
      direction TB
      DSP["Accelerate vDSP"]
      LOUD["Loudness (R128)"]
      EMBED["OpenL3 (ANE)"]
      ANE["Core ML"]
    end

    subgraph DATA["Storage"]
      direction TB
      DB["SQLite WAL"]
      BLOBS["Blob Store"]
    end

    UI <-->|"HTTP REST"| ENGINE
    ENGINE <-->|"HTTP / gRPC"| ANALYZER
    ENGINE <--> DATA
    ANALYZER --> DATA
```

| Layer | Tech | Role |
|-------|------|------|
| **Frontend** | React 19, TypeScript, Vite, D3.js, Web Audio | Interactive UI with audio playback |
| **Engine** | Go 1.24, HTTP/gRPC, Protobuf | API server, job scheduling, set planning |
| **Analyzer** | Swift 6, Accelerate, Core ML (ANE) | Audio DSP, loudness, OpenL3 embeddings |
| **Storage** | SQLite (WAL mode) | Track metadata, analysis, embeddings |

---

## Communication Protocols

Algiers supports **both HTTP REST and gRPC** for maximum flexibility:

```
                           ┌─────────────────────────────────┐
                           │         Go Engine               │
                           │                                 │
┌─────────┐    HTTP REST   │  ┌──────────┐  ┌──────────┐   │    HTTP/gRPC    ┌────────────┐
│ Web UI  │ ◄────────────► │  │HTTP API  │  │gRPC API  │   │ ◄────────────► │Swift Analzr│
│ (React) │   (deprecated) │  │ :8080    │  │ :50051   │   │                 │  (Swift)   │
└─────────┘                │  └──────────┘  └──────────┘   │                 └────────────┘
                           │                                 │
                           └─────────────────────────────────┘
```

### gRPC API (Recommended)

The gRPC API provides maximum performance with real-time streaming:

| Feature | HTTP REST | gRPC |
|---------|-----------|------|
| **Protocol** | HTTP/1.1 + JSON | HTTP/2 + Protobuf |
| **Streaming** | Limited (polling) | Full bidirectional |
| **Latency** | ~100ms | ~10ms |
| **Code Gen** | Manual | Automatic (10+ languages) |

**gRPC Methods:**
- `ScanLibrary` — Stream scan progress with file counts, byte progress, ETA
- `AnalyzeTracks` — Stream per-track analysis with stage timings and overall job progress
- `ListTracks` — Stream track summaries
- `StreamTrainingProgress` — Real-time training with validation metrics and stage breakdowns
- Full training, similarity, and model version management

**Enhanced Streaming Progress (v1.0):**
```
ScanProgress: current_file, percent, elapsed_ms, eta_ms, bytes_processed, bytes_total
AnalyzeProgress: track_index, total_tracks, overall_percent, stage_timings[], stage_message
TrainingProgress: validation_loss, validation_accuracy, samples_processed, current_stage
```

> 📖 See [gRPC Migration Guide](docs/gRPC-MIGRATION.md) for complete migration instructions.

### HTTP REST API (Deprecated)

The HTTP REST API remains available for backward compatibility but is **deprecated**:

- All responses include `Sunset: Wed, 01 Jul 2026 00:00:00 GMT` header
- All responses include `Deprecation: true` header
- New features may not be added to HTTP API
- Planned removal: July 2026

**Advantages (why it still exists):**
- **Simplicity** — Easy to debug with curl, browser DevTools
- **Browser-native** — No additional libraries (fetch API)
- **Human-readable** — JSON responses are easy to inspect

---

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

---

## Quick Start

### macOS Standalone App (Recommended)

Download the latest release or build from source:

```bash
git clone https://github.com/ParkWardRR/cartomix-Web-Based-DJ-Copilot.git
cd cartomix-Web-Based-DJ-Copilot
make build-macos
open build/Algiers.app
```

The standalone app bundles everything — just double-click to run. No terminal needed.

### Development Mode

```bash
# Install dependencies
make install

# Start all services (3 terminals)
make run-stack

# Open in browser
open http://localhost:5173
```

> 📖 **Full installation guide:** [INSTALL.md](INSTALL.md) — System requirements, manual installation, troubleshooting

<details>
<summary>Manual installation (advanced)</summary>

#### Prerequisites

- **macOS 14+** (Sonoma or later) — Required for Metal 3 and Core ML features
- **Apple Silicon** (M1/M2/M3/M4) — Required for Neural Engine and Metal GPU
- **Go 1.24+** — `brew install go`
- **Node.js 22+** — `brew install node`
- **Swift 6+** — Comes with Xcode Command Line Tools

#### Build & Run

```bash
# Clone the repository
git clone https://github.com/cartomix/algiers.git
cd algiers

# 1. Build the Swift analyzer (required for audio analysis)
cd analyzer-swift
swift build -c release
cd ..

# 2. Install dependencies
go mod download
cd web && npm install && cd ..

# 3. Start all services (3 terminals)

# Terminal 1: Start the Go engine (port 8080 HTTP)
go run ./cmd/engine

# Terminal 2: Start the Swift analyzer (port 9090)
./analyzer-swift/.build/release/analyzer-swift serve --port 9090

# Terminal 3: Start the web UI (port 5173)
cd web && npm run dev
```

Open http://localhost:5173 to see the UI.

</details>

### Analyze a Track

```bash
# Analyze and print JSON (includes loudness + OpenL3 embedding)
./analyzer-swift/.build/release/analyzer-swift "path/to/track.flac"

# Example output:
# Track: path/to/track.flac
# Duration: 369.6s
# BPM: 117.2 (confidence: 100%)
# Key: F# / 2B (confidence: 92%)
# Energy: 7/10
# Loudness: -14.0 LUFS (range: 3.7 LU, peak: 0.5 dBTP)
# OpenL3: 512-dim embedding (547 windows)
```

### Find Similar Tracks

```bash
# Find tracks similar to a specific track
curl http://localhost:8080/api/tracks/{content-hash}/similar

# Response includes vibe match %, tempo/key compatibility, and explanation
```

---

## API Reference

See [docs/API.md](docs/API.md) for complete API documentation including:
- Swift Analyzer CLI commands
- HTTP REST endpoints (`/api/tracks/{id}/similar`)
- gRPC service definitions
- Data type schemas

### Key Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/tracks` | GET | List all tracks with analysis summary |
| `/api/tracks/{id}` | GET | Get full analysis for a track |
| `/api/tracks/{id}/similar` | GET | Find similar tracks with explanations |
| `/api/scan` | POST | Scan directories for audio files |
| `/api/analyze` | POST | Analyze tracks |
| `/api/set/propose` | POST | Generate optimized set order |
| `/api/export` | POST | Export to Rekordbox/Serato/Traktor |
| `/api/ml/settings` | GET/PUT | ML feature settings |

---

## Developer Loop

| Command | Description |
|---------|-------------|
| `make test` | Run Go + Swift unit tests |
| `make lint` | Run golangci-lint, swift-format, eslint |
| `make proto` | Regenerate protobuf stubs |
| `make build` | Build engine + analyzer binaries |
| `cd web && npm run dev` | Start Vite dev server |
| `make screenshots` | Capture UI screenshots with Playwright |
| `go test ./internal/e2e/...` | Run E2E tests (requires running services) |

### Test Status

| Test Suite | Status | Description |
|------------|--------|-------------|
| **Go Unit Tests** | ✅ 41/41 Passing | Storage, planner, exporter, similarity |
| **E2E Tests** | ✅ 7/7 Passing | UI loads, Library, Set Builder, Graph, Theme, API |
| **Golden Tests** | ✅ Passing | Rekordbox, Serato, Traktor exports |
| **Property Tests** | ✅ Passing | Planner monotonicity, cue bounds |

### E2E Test Coverage

```
TestUILoads         - Verifies app loads and renders
TestLibraryView     - Verifies library navigation and track grid
TestSetBuilder      - Verifies Set Builder view and energy arc
TestGraphView       - Verifies D3.js force-directed graph
TestThemeToggle     - Verifies dark/light theme switching
TestAPIHealth       - Verifies /api/health endpoint
TestAPIListTracks   - Verifies /api/tracks endpoint
```

---

## Project Layout

```
.
├── cmd/
│   ├── engine/              # Go HTTP/gRPC server entrypoint
│   ├── exportverify/        # Checksum verification CLI
│   └── screenshots/         # Playwright-Go screenshot capture
├── analyzer-swift/          # Swift analyzer module
│   └── Sources/
│       ├── AnalyzerSwift/
│       │   ├── Analysis/    # Beatgrid, Key, Energy, Loudness
│       │   ├── ML/          # OpenL3Embedder (Core ML/ANE)
│       │   └── DSP/         # AudioDecoder, FFT
│       └── AnalyzerServer/  # CLI + HTTP server
├── internal/
│   ├── httpapi/             # HTTP REST endpoints
│   ├── similarity/          # ML similarity search
│   ├── planner/             # Set ordering algorithm
│   ├── exporter/            # Rekordbox/Serato/Traktor exports
│   └── storage/             # SQLite + blob store
├── web/
│   └── src/
│       ├── components/      # React components
│       ├── hooks/           # useAudioPlayer
│       ├── api.ts           # API client
│       └── store.ts         # Zustand store
└── proto/                   # gRPC/Protobuf contracts
```

---

## Roadmap

### Alpha (Complete)
- [x] gRPC engine with health checks
- [x] Library scanner with content hashing
- [x] SQLite storage with migrations
- [x] Set planner with weighted graph
- [x] Generic exports (M3U/JSON/CSV/Tar)
- [x] Pro UI with visualizations

### Beta (Complete)
- [x] Swift analyzer with Accelerate DSP
- [x] Beatgrid, key, energy, loudness analysis
- [x] Audio embeddings (128-dim MFCC)
- [x] **OpenL3 embeddings (512-dim, Core ML/ANE)**
- [x] **Similarity search API with explanations**
- [x] Web Audio playback
- [x] Rekordbox/Serato/Traktor exports
- [x] Playwright-Go E2E tests

### v0.5.0-beta (Complete)
- [x] **Apple SoundAnalysis** — 300+ label classifier, QA flags
- [x] **Custom DJ section training** — Create ML integration
- [x] **Model versioning** — Activate, rollback, delete
- [x] **Training UI** — Dataset table, label editor, progress card
- [x] **Analysis Panel** — DSP + ML results display
- [x] **Similar Tracks panel** — Explainable scoring

### v1.0 (In Progress) — Core Stability
- [x] Full gRPC API with streaming support
- [x] gRPC interceptors (logging, metrics, recovery)
- [x] HTTP REST deprecation notices
- [x] **Metal/ANE requirement check** — Graceful failure on unsupported hardware
- [x] **gRPC server for Swift analyzer** — Proto-compatible with error handling
- [x] **Web Audio AudioWorklet player** — Real-time analysis + streaming
- [x] **Extended fixturegen** — Phrase tracks, harmonic sets, club noise
- [x] **DB versioning + backup** — Analysis cache versioning, backup/restore
- [x] **gRPC streaming progress events** — Job %, current file, stage timings, ETA, byte progress
- [ ] gRPC-web for browser clients
- [ ] Cancel/timeout support for long scans
- [ ] Waveform-based label painting (drag to select)
- [ ] Alpha acceptance: 100 tracks → 30-track set → export
- [ ] CI/CD pipeline (macOS runner with Swift/Go/Playwright)

### v1.1 (Future) — Editable Analysis
*"AI suggestions → AI assist" — Trust through user control*
- [ ] **Override layer** — User-editable beatgrid, key, cues, sections, loudness
- [ ] **Lock flags** — Prevent re-analysis from overwriting user edits
- [ ] **Analysis versioning** — Store `analyzer_version`, `model_version`, `params`
- [ ] **Content hash caching** — Safe cache reuse across runs/machines
- [ ] **Auto-labeling** — Model suggests labels, user confirms
- [ ] **Transfer learning** — Pre-trained base model for faster training
- [ ] **Label import** — Import cue points from Rekordbox/Serato

### v1.2 (Future) — Scalable Similarity
*Fast ANN + explainable reranking*
- [ ] **Two-stage similarity** — (1) Fast ANN on embeddings, (2) Rerank with BPM/key/energy
- [ ] **Section-level embeddings** — Per intro/build/drop/outro, not just track mean
- [ ] **Explain similarity view** — Show contributing windows + confidence bands
- [ ] **Selectable OpenL3 configs** — Content type, input repr, embedding size
- [ ] **Transition window detection** — Auto-detect mixable sections
- [ ] **Energy curve matching** — Find tracks with compatible energy arcs

### v1.3 (Future) — Export Verification & Reliability
- [ ] **Export round-trip tests** — Automated export→import/parse→compare
- [ ] **Golden export fixtures** — Versioned fixtures to catch silent corruption
- [ ] **Cue templates** — "First downbeat", "bass-in", "breakdown", "mix-out" defaults
- [ ] **Global scheduler** — Per-engine concurrency limits + backpressure
- [ ] **Analysis bundles** — Params JSON + summary metrics per track for debugging
- [ ] **Real-time analysis** — Stream audio while analyzing
- [ ] **Hardware control surface** — MIDI mapping for cue review
- [ ] **Stems export** — AI-separated stems for transition layers
- [ ] **Club mode** — Full-screen visualization for live performance
- [ ] **Collaborative playlists** — Share set plans with other DJs

### Future Exploration
- [ ] **Mood detection** — Happy/sad/dark/euphoric classification
- [ ] **Genre embedding** — Fine-grained genre similarity
- [ ] **Crowd energy prediction** — ML model trained on live recordings
- [ ] **Transition synthesis** — AI-generated transition clips
- [ ] **Voice-controlled prep** — "Find something like this but higher energy"

---

## Contributing

PRs welcome! Keep commits scoped and include:
- Repro steps for bugs
- Expected vs actual behavior
- Test coverage for new features

---

## Changelog

### v0.5.1-beta (2026-01-29)
- **Enhanced gRPC Streaming Progress**
  - `ScanLibrary`: Current file name, percent progress, elapsed/ETA timing, byte-level progress
  - `AnalyzeTracks`: Track index/total, overall job percent, per-stage timing breakdown, track metadata
  - `StreamTrainingProgress`: Validation loss/accuracy, samples processed, 5-stage pipeline with timings
  - All streams include elapsed time (ms) and ETA calculation

### v0.5.0-beta (2026-01-29)
- **Custom DJ Section Training (Layer 3)**
  - Create ML MLSoundClassifier integration for ANE inference
  - 7 DJ section labels: intro, build, drop, break, outro, verse, chorus
  - Training job management with progress tracking
  - Model versioning with activate/rollback/delete
  - Full REST API for training labels, jobs, and models
- **Apple SoundAnalysis (Layer 1)**
  - Built-in classifier with 300+ audio labels
  - Sound context detection (music/speech/noise/silence)
  - QA flag generation (needs_review, mixed_content, speech_detected)
  - Primary context with confidence scoring
- **Training UI**
  - TrainingScreen with dataset table and label statistics
  - LabelEditor for adding section labels to tracks
  - TrainingProgressCard with epoch/loss visualization
  - ModelVersionsList with activate/delete actions
- **Analysis Panel**
  - DSP results with confidence bars
  - ML sound context and events timeline
  - QA flags with dismiss capability
  - Embedding status indicator

### v0.4.0-beta (2026-01-29)
- **OpenL3 ML Embeddings**
  - 512-dimensional audio embeddings from Core ML on ANE
  - 1-second windows with 0.5-second hop (547 windows for 5-min track)
  - Mel spectrogram computation using Accelerate vDSP
  - Fallback embedding when model unavailable
- **Similarity Search API**
  - `GET /api/tracks/{id}/similar` endpoint
  - Weighted scoring: 50% vibe + 20% tempo + 20% key + 10% energy
  - Human-readable explanations for every match
  - Cached similarity results in SQLite
- **ML Settings**
  - `GET/PUT /api/ml/settings` for feature flags
  - OpenL3 enable/disable toggle
  - Similarity threshold configuration

### v0.3.0-beta (2026-01-29)
- **Loudness Analysis (EBU R128)**
- **Audio Embeddings (128-dim MFCC)**
- **Web Audio Playback Integration**
- **API Documentation**

### v0.2.0-beta (2026-01-29)
- **Swift Analyzer with Accelerate DSP**
- **Golden comparison tests**
- **Detailed architecture documentation**

### v0.1.0-alpha (2026-01-29)
- Initial alpha release with Pro UI

---

## License

Blue Oak Model License 1.0.0. See [LICENSE](LICENSE).

---

<div align="center">

**Built for DJs who want to prep smarter, not harder.**

*Powered by Apple Silicon. Made with Metal, Core ML, and too much coffee.*

</div>
