<div align="center">

# AI/ML Architecture Guide

**Deep dive into Algiers' Apple-first ML stack — OpenL3, SoundAnalysis, and custom training**

[![OpenL3](https://img.shields.io/badge/OpenL3%20512--dim-8B5CF6?style=for-the-badge)](#layer-2-openl3-embeddings)
[![SoundAnalysis](https://img.shields.io/badge/Apple%20SoundAnalysis-FF3B30?style=for-the-badge)](#layer-1-apple-soundanalysis)
[![Custom Training](https://img.shields.io/badge/Custom%20Training-F59E0B?style=for-the-badge)](#layer-3-custom-dj-section-classifier)
[![Neural Engine](https://img.shields.io/badge/Neural%20Engine-FF9500?style=for-the-badge)](#apple-neural-engine-ane)
[![Local Only](https://img.shields.io/badge/100%25%20Local-222222?style=for-the-badge)](#privacy--local-first)

</div>

---

## Table of Contents

- [Overview](#overview)
- [3-Layer Architecture](#3-layer-architecture)
- [Layer 1: Apple SoundAnalysis](#layer-1-apple-soundanalysis)
- [Layer 2: OpenL3 Embeddings](#layer-2-openl3-embeddings)
- [Layer 3: Custom DJ Section Classifier](#layer-3-custom-dj-section-classifier)
- [Similarity Search Algorithm](#similarity-search-algorithm)
- [Explainability System](#explainability-system)
- [Apple Neural Engine (ANE)](#apple-neural-engine-ane)
- [Hardware Requirements](#hardware-requirements)
- [Privacy & Local-First](#privacy--local-first)
- [Performance Benchmarks](#performance-benchmarks)
- [File Locations](#file-locations)
- [API Reference](#api-reference)
- [Research Background](#research-background)

---

## Overview

Algiers uses a **3-layer ML architecture** designed specifically for Apple Silicon. Every layer runs on the Apple Neural Engine (ANE) for maximum performance while keeping all audio processing 100% local.

### Design Principles

| Principle | Implementation |
|-----------|----------------|
| **Local-first** | No cloud dependencies; audio never leaves your Mac |
| **Explainable** | Every AI decision includes human-readable rationale |
| **Incremental** | Layers build on each other; all are optional |
| **Hardware-optimized** | Native ANE/Metal acceleration on M1-M5 |
| **User-controlled** | Override any AI suggestion; lock edits from re-analysis |

### What Each Layer Does

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Algiers ML Architecture                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ LAYER 1: Apple SoundAnalysis (Built-in)                              │  │
│  │ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━                               │  │
│  │ • 300+ pre-trained audio labels (music, speech, noise, etc.)         │  │
│  │ • Zero configuration — works out of the box                          │  │
│  │ • Generates QA flags for tracks needing review                       │  │
│  │ • Use case: Content detection, quality assurance                     │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ LAYER 2: OpenL3 Embeddings (Pre-trained)                             │  │
│  │ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━                               │  │
│  │ • 512-dimensional audio embeddings                                   │  │
│  │ • Trained on millions of audio-video pairs                           │  │
│  │ • Captures timbre, texture, mood, genre characteristics              │  │
│  │ • Use case: Similarity search, vibe matching, transitions            │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                    │                                        │
│                                    ▼                                        │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │ LAYER 3: Custom DJ Section Classifier (Opt-in Training)              │  │
│  │ ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━                               │  │
│  │ • 7 DJ-specific labels: intro, build, drop, break, outro, verse, chorus │
│  │ • Train on YOUR labeled data using Create ML                         │  │
│  │ • Model versioning with instant rollback                             │  │
│  │ • Use case: Automated section detection matching your style          │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │                    ⚡ All Layers Run on ANE                           │  │
│  │              Apple Neural Engine — 16 TOPS — 100% Local              │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3-Layer Architecture

### Layer Comparison

| Aspect | Layer 1 | Layer 2 | Layer 3 |
|--------|---------|---------|---------|
| **Name** | Apple SoundAnalysis | OpenL3 Embeddings | Custom Classifier |
| **Model Source** | Apple built-in | Pre-trained (Google Research) | Your training data |
| **Training Required** | None | None | Yes (opt-in) |
| **Output** | Labels + confidence | 512-dim vector | Section labels |
| **Primary Use** | QA, content type | Similarity search | Section detection |
| **Latency (5min track)** | ~500ms | ~2.5s | ~1s |
| **Disk Space** | 0 (system) | ~50MB | ~2-5MB per model |

### When to Use Each Layer

```
┌───────────────────────────────────────────────────────────────────────────┐
│                          Decision Flow                                     │
├───────────────────────────────────────────────────────────────────────────┤
│                                                                            │
│  "Is this audio music or speech?"                                         │
│       │                                                                    │
│       ▼                                                                    │
│  ┌─────────────────┐                                                      │
│  │  LAYER 1        │  ← Quick categorization, QA flagging                 │
│  │  SoundAnalysis  │                                                      │
│  └────────┬────────┘                                                      │
│           │                                                                │
│  "Find tracks with similar vibe"                                          │
│       │                                                                    │
│       ▼                                                                    │
│  ┌─────────────────┐                                                      │
│  │  LAYER 2        │  ← Semantic similarity beyond BPM/key                │
│  │  OpenL3         │                                                      │
│  └────────┬────────┘                                                      │
│           │                                                                │
│  "Label sections my way"                                                   │
│       │                                                                    │
│       ▼                                                                    │
│  ┌─────────────────┐                                                      │
│  │  LAYER 3        │  ← Custom section detection                          │
│  │  Custom Model   │                                                      │
│  └─────────────────┘                                                      │
│                                                                            │
└───────────────────────────────────────────────────────────────────────────┘
```

---

## Layer 1: Apple SoundAnalysis

### Overview

Apple's SoundAnalysis framework provides a **built-in classifier** with 300+ audio labels. It runs with zero configuration and requires no training data.

### How It Works

```swift
// Swift implementation (analyzer-swift/Sources/AnalyzerSwift/Analysis/SoundClassifier.swift)
import SoundAnalysis

let request = try SNClassifySoundRequest(classifierIdentifier: .version1)
let analyzer = SNAudioStreamAnalyzer(format: audioFormat)
try analyzer.add(request, withObserver: observer)

// Process audio in 0.5-second windows
analyzer.analyze(buffer, atAudioFramePosition: framePosition)
```

### Label Categories

Algiers maps Apple's 300+ labels to DJ-relevant categories:

| Category | Apple Labels (Examples) | DJ Relevance |
|----------|------------------------|--------------|
| **Music** | `music`, `electronic_music`, `hip_hop_music`, `rock_music` | Primary content |
| **Speech** | `speech`, `singing`, `crowd`, `applause` | Vocal detection |
| **Noise** | `noise`, `static`, `hum`, `buzz` | Quality issues |
| **Silence** | `silence` | Track boundaries |
| **Percussion** | `drum`, `snare`, `hi_hat`, `kick` | Beat-heavy sections |

### QA Flag Generation

Based on classification, Algiers generates actionable flags:

| Flag | Condition | Meaning |
|------|-----------|---------|
| `needs_review` | Primary context confidence < 50% | Ambiguous audio |
| `mixed_content` | Speech + music both >20% | Vocal samples present |
| `speech_detected` | >10s continuous speech | May need attention |
| `low_confidence` | Average confidence < 60% | Unusual audio |

### API Response

```json
GET /api/tracks/{id}

{
  "sound_context": "music",
  "sound_context_confidence": 0.94,
  "sound_events": [
    {
      "label": "electronic_music",
      "category": "music",
      "confidence": 0.89,
      "start_time": 0.0,
      "end_time": 180.0
    },
    {
      "label": "singing",
      "category": "speech",
      "confidence": 0.72,
      "start_time": 45.2,
      "end_time": 52.8
    }
  ],
  "qa_flags": [
    {
      "type": "mixed_content",
      "reason": "Singing detected in music track (7.6s at 0:45)",
      "dismissed": false
    }
  ]
}
```

---

## Layer 2: OpenL3 Embeddings

### What is OpenL3?

OpenL3 is a deep neural network trained on **millions of audio-video pairs** from YouTube. It learned to encode audio in a way that captures perceptual similarity — sounds that "feel" alike end up close together in the embedding space.

### Technical Details

| Specification | Value |
|---------------|-------|
| **Architecture** | CNN-based encoder |
| **Input** | 128-band mel spectrogram (199 frames = 1 second) |
| **Output** | 512-dimensional float vector |
| **Training Data** | 60M+ audio-video pairs |
| **Model Size** | ~50MB (FP16 quantized) |
| **Inference** | ~5ms/window on ANE |

### Embedding Generation Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        OpenL3 Embedding Pipeline                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Audio (48kHz) ──▶ Mel Spectrogram ──▶ Core ML Model ──▶ 512-dim Vector    │
│                                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐                  │
│  │   1 second   │    │  128 bands   │    │   OpenL3     │                  │
│  │   window     │───▶│  199 frames  │───▶│  (on ANE)    │──▶ [0.12, 0.45, │
│  │   (0.5s hop) │    │  log-scale   │    │   ~5ms       │      ..., 0.87] │
│  └──────────────┘    └──────────────┘    └──────────────┘                  │
│                                                                              │
│  Full Track: 547 windows ──▶ Mean Pool ──▶ Single 512-dim track embedding  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Swift Implementation

```swift
// analyzer-swift/Sources/AnalyzerSwift/ML/OpenL3Embedder.swift

public final class OpenL3Embedder {
    private let model: MLModel

    public init() throws {
        let config = MLModelConfiguration()
        config.computeUnits = .all  // Prefers ANE, falls back to GPU/CPU

        let modelURL = Bundle.module.url(forResource: "OpenL3Music", withExtension: "mlmodelc")!
        model = try MLModel(contentsOf: modelURL, configuration: config)
    }

    /// Generate embeddings for all 1-second windows
    public func generateEmbeddings(_ samples: [Float], sampleRate: Int) -> [OpenL3Embedding] {
        let windowSamples = sampleRate  // 1 second
        let hopSamples = sampleRate / 2  // 0.5 second hop

        var embeddings: [OpenL3Embedding] = []
        var position = 0

        while position + windowSamples <= samples.count {
            let window = Array(samples[position..<position + windowSamples])
            let melSpec = computeMelSpectrogram(window, sampleRate: sampleRate)
            let embedding = try! model.prediction(from: melSpec)

            embeddings.append(OpenL3Embedding(
                timestamp: Double(position) / Double(sampleRate),
                vector: embedding.featureValue(for: "embedding")!.multiArrayValue!
            ))

            position += hopSamples
        }

        return embeddings
    }

    /// Pool window embeddings into single track embedding
    public func poolEmbeddings(_ windows: [OpenL3Embedding]) -> [Float] {
        guard !windows.isEmpty else { return Array(repeating: 0, count: 512) }

        var pooled = [Float](repeating: 0, count: 512)
        for window in windows {
            for i in 0..<512 {
                pooled[i] += window.vector[i]
            }
        }
        for i in 0..<512 {
            pooled[i] /= Float(windows.count)
        }
        return pooled
    }
}
```

### Why OpenL3 Over Genre Tags?

| Approach | Captures | Misses |
|----------|----------|--------|
| **Genre tags** | Human categories | Nuance, mood, texture |
| **BPM/Key only** | Technical compatibility | Feel, vibe, energy arc |
| **OpenL3** | Perceptual similarity | Human-readable categories |

**Algiers combines all three** — OpenL3 for vibe, BPM/key for mixing compatibility, and human-readable explanations for trust.

### What OpenL3 Captures

Through its training on millions of audio-video pairs, OpenL3 learned to encode:

- **Timbre** — The "color" of sound (bright vs dark, warm vs harsh)
- **Texture** — Rhythmic density, layering complexity, sparseness
- **Mood** — Energy level, tension, euphoria, melancholy
- **Production style** — Compression, reverb characteristics, mix quality
- **Instrumentation** — Synthesizers vs acoustic, electronic vs organic

---

## Layer 3: Custom DJ Section Classifier

> **Full documentation:** [docs/ML-TRAINING.md](ML-TRAINING.md)

### Overview

Layer 3 enables **opt-in custom model training** for DJ section classification. Train a model that labels sections the way *you* do.

### Labels

| Label | Description | Color |
|-------|-------------|-------|
| `intro` | Track opening, minimal elements | 🟢 Green |
| `build` | Energy rising, tension building | 🟡 Yellow |
| `drop` | Peak energy, full arrangement | 🔴 Red |
| `break` | Breakdown, stripped back | 🟣 Purple |
| `outro` | Track ending, elements fading | 🔵 Blue |
| `verse` | Vocal or melodic verse | ⬛ Gray |
| `chorus` | Main hook or phrase | 🩷 Pink |

### Training Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       Custom Model Training Pipeline                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌────────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐     │
│  │   Label    │    │  Extract   │    │  Train     │    │  Export    │     │
│  │   Tracks   │───▶│  Features  │───▶│  Model     │───▶│  mlmodelc  │     │
│  │            │    │            │    │            │    │            │     │
│  │ Min 10/cls │    │ Mel specs  │    │ Create ML  │    │ Core ML    │     │
│  │ 7 classes  │    │ MFCCs      │    │ ~2 min     │    │ ANE-ready  │     │
│  └────────────┘    └────────────┘    └────────────┘    └────────────┘     │
│                                                                              │
│  Model Versioning: v1 → v2 → v3 (active) → v4                              │
│                    One-click rollback to any previous version               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Swift Implementation

```swift
// analyzer-swift/Sources/AnalyzerSwift/ML/DJSectionTrainer.swift

public final class DJSectionTrainer {

    /// Train a new DJ section model
    public func train(data: [TrainingData], outputPath: URL) async throws -> TrainingResult {
        // 1. Export audio segments to temp directory
        let dataSource = try await prepareDataSource(data)

        // 2. Create ML model using MLSoundClassifier
        let params = MLSoundClassifier.ModelParameters(
            validationData: .automatic(splitPercentage: 0.2),
            maxIterations: 10
        )

        let job = try MLSoundClassifier.train(
            trainingData: dataSource,
            parameters: params
        )

        // 3. Wait for training to complete
        let model = try await job.result

        // 4. Export to Core ML format
        try model.write(to: outputPath)

        // 5. Compute metrics
        let metrics = try model.evaluation(on: dataSource)

        return TrainingResult(
            accuracy: metrics.classificationError,
            f1Score: computeF1(metrics),
            modelPath: outputPath
        )
    }
}
```

### Model Versioning

```
/models/
├── dj_section_v1.mlmodelc    # First training
├── dj_section_v1.json        # Metadata + metrics
├── dj_section_v2.mlmodelc    # More data, better accuracy
├── dj_section_v2.json
└── dj_section_v3.mlmodelc    # Current active ←
```

**Rollback:** One-click restore to any previous version via API or UI.

---

## Similarity Search Algorithm

### Weighted Scoring Formula

```
Combined Score = 0.50 × OpenL3 Similarity    (vibe match)
              + 0.20 × Tempo Similarity      (BPM compatibility)
              + 0.20 × Key Similarity        (harmonic compatibility)
              + 0.10 × Energy Similarity     (energy level match)
```

### Component Calculations

#### 1. OpenL3 Similarity (50%)

Cosine similarity between 512-dimensional embedding vectors:

```go
// internal/similarity/similarity.go

func CosineSimilarity(a, b []float32) float64 {
    var dot, normA, normB float64
    for i := range a {
        dot += float64(a[i]) * float64(b[i])
        normA += float64(a[i]) * float64(a[i])
        normB += float64(b[i]) * float64(b[i])
    }
    if normA == 0 || normB == 0 {
        return 0
    }
    return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

#### 2. Tempo Similarity (20%)

BPM compatibility with half/double tempo matching:

```go
func ComputeTempoSimilarity(bpmA, bpmB float64) float64 {
    if bpmA <= 0 || bpmB <= 0 {
        return 0
    }

    // Direct match
    diff := math.Abs(bpmA - bpmB)

    // Half tempo match (e.g., 70 BPM ≈ 140 BPM)
    halfDiff := math.Abs(bpmA - bpmB/2)
    doubleDiff := math.Abs(bpmA - bpmB*2)

    minDiff := math.Min(diff, math.Min(halfDiff, doubleDiff))

    // Score: 1.0 at 0 BPM diff, 0.0 at 10+ BPM diff
    return math.Max(0, 1.0 - minDiff/10.0)
}
```

#### 3. Key Similarity (20%)

Camelot wheel compatibility:

| Relation | Score | Example |
|----------|-------|---------|
| Same | 1.0 | 8A → 8A |
| Relative (A↔B) | 0.9 | 8A → 8B |
| Adjacent (±1) | 0.8 | 8A → 9A |
| Two steps | 0.6 | 8A → 10A |
| Distant | 0.2 | 8A → 3B |

#### 4. Energy Similarity (10%)

Energy level difference (1-10 scale):

```go
func ComputeEnergySimilarity(energyA, energyB int32) float64 {
    diff := math.Abs(float64(energyA) - float64(energyB))

    // Score: 1.0 at same energy, 0.0 at 5+ difference
    return math.Max(0, 1.0 - diff/5.0)
}
```

### Full Algorithm

```go
// internal/similarity/similarity.go

func FindSimilarTracks(db *storage.DB, queryID int64, limit int) ([]SimilarityResult, error) {
    query, err := db.GetTrackAnalysis(queryID)
    if err != nil {
        return nil, err
    }

    candidates, err := db.GetAllAnalyses()
    if err != nil {
        return nil, err
    }

    var results []SimilarityResult

    for _, candidate := range candidates {
        if candidate.TrackID == queryID {
            continue
        }

        // Compute component similarities
        vibeSim := CosineSimilarity(query.OpenL3Embedding, candidate.OpenL3Embedding)
        tempoSim := ComputeTempoSimilarity(query.BPM, candidate.BPM)
        keySim := ComputeKeySimilarity(query.Key, candidate.Key)
        energySim := ComputeEnergySimilarity(query.Energy, candidate.Energy)

        // Weighted combination
        score := 0.50*vibeSim + 0.20*tempoSim + 0.20*keySim + 0.10*energySim

        // Generate explanation
        explanation := BuildExplanation(vibeSim, tempoSim, keySim, energySim, query, candidate)

        results = append(results, SimilarityResult{
            TrackID:     candidate.TrackID,
            Score:       score,
            Explanation: explanation,
            VibeSim:     vibeSim,
            TempoSim:    tempoSim,
            KeySim:      keySim,
            EnergySim:   energySim,
        })
    }

    // Sort by score descending
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    if len(results) > limit {
        results = results[:limit]
    }

    return results, nil
}
```

---

## Explainability System

### Design Philosophy

Every AI decision in Algiers includes a **human-readable explanation**. Users should never wonder "why did it suggest this?"

### Explanation Format

```
"similar vibe (82%); Δ+2 BPM; key: 8A→9A (compatible); energy +1; beat-grid aligned"
```

### Explanation Components

| Component | Format | Meaning |
|-----------|--------|---------|
| **Vibe match** | `similar vibe (82%)` | OpenL3 cosine similarity as percentage |
| **Tempo delta** | `Δ+2 BPM` or `Δ-3 BPM` | BPM difference with direction |
| **Key relation** | `8A→9A (compatible)` | Camelot notation with compatibility type |
| **Energy flow** | `energy +1` or `energy -2` | Energy level change |
| **Beat alignment** | `beat-grid aligned` or `beat offset` | Downbeat alignment status |
| **ML confidence** | `(87% confidence)` | When custom model provides prediction |

### Key Compatibility Types

| Type | Camelot Rule | Explanation Text |
|------|--------------|------------------|
| **Same** | Identical | `same key` |
| **Relative** | A↔B same number | `relative key` |
| **Compatible** | ±1 same mode | `compatible` |
| **Harmonic** | Various | `harmonic` |
| **Clash** | Distant | `key clash ⚠` |

### Explanation Builder

```go
// internal/similarity/similarity.go

func BuildExplanation(vibeSim, tempoSim, keySim, energySim float64,
                      query, candidate *AnalysisRecord) string {
    var parts []string

    // Vibe match
    vibePercent := int(vibeSim * 100)
    parts = append(parts, fmt.Sprintf("similar vibe (%d%%)", vibePercent))

    // Tempo delta
    bpmDelta := candidate.BPM - query.BPM
    if bpmDelta > 0 {
        parts = append(parts, fmt.Sprintf("Δ+%.0f BPM", bpmDelta))
    } else if bpmDelta < 0 {
        parts = append(parts, fmt.Sprintf("Δ%.0f BPM", bpmDelta))
    } else {
        parts = append(parts, "tempo match")
    }

    // Key relation
    keyRelation := ComputeKeyRelation(query.Key, candidate.Key)
    parts = append(parts, fmt.Sprintf("key: %s→%s (%s)",
        query.Key, candidate.Key, keyRelation))

    // Energy flow
    energyDelta := int(candidate.Energy) - int(query.Energy)
    if energyDelta != 0 {
        parts = append(parts, fmt.Sprintf("energy %+d", energyDelta))
    } else {
        parts = append(parts, "same energy")
    }

    return strings.Join(parts, "; ")
}
```

---

## Apple Neural Engine (ANE)

### What is the ANE?

The Apple Neural Engine is a **dedicated ML accelerator** integrated into Apple Silicon chips. It provides:

- **16 TOPS** (trillion operations per second) on M1
- **15x faster** than CPU for ML inference
- **1/15th power consumption** compared to CPU
- **Hardware matrix multiply** optimized for neural networks

### How Algiers Uses ANE

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Apple Silicon (M1-M5)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐ │
│  │  CPU Cores   │   │    GPU       │   │Neural Engine │   │   Media      │ │
│  │  (Perf + E)  │   │   (Metal)    │   │    (ANE)     │   │   Engine     │ │
│  │              │   │              │   │              │   │              │ │
│  │  Planner     │   │  Waveform    │   │  OpenL3      │   │  Audio       │ │
│  │  Algorithm   │   │  Render      │   │  Inference   │   │  Decode      │ │
│  └──────────────┘   └──────────────┘   └──────────────┘   └──────────────┘ │
│                                                                              │
│                     ┌─────────────────────────────────┐                     │
│                     │     Unified Memory (UMA)        │                     │
│                     │    Zero-copy data sharing        │                     │
│                     └─────────────────────────────────┘                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Core ML Configuration

```swift
// Force ANE usage for ML inference
let config = MLModelConfiguration()
config.computeUnits = .all  // Prefers ANE, falls back to GPU, then CPU

// Load model with ANE optimization
let model = try MLModel(contentsOf: modelURL, configuration: config)
```

### ANE vs CPU Performance

| Operation | CPU | ANE | Speedup |
|-----------|-----|-----|---------|
| OpenL3 (1 window) | 75ms | 5ms | 15x |
| OpenL3 (full track) | 40s | 2.5s | 16x |
| SoundAnalysis | 8s | 500ms | 16x |
| Custom classifier | 15s | 1s | 15x |

---

## Hardware Requirements

### Minimum Requirements

| Component | Requirement |
|-----------|-------------|
| **Chip** | Apple Silicon (M1 or later) |
| **macOS** | 13.0 Ventura or later |
| **RAM** | 8 GB (16 GB recommended) |
| **Storage** | 500 MB for models + library |

### Hardware Check

Algiers performs a hardware check at startup:

```swift
// analyzer-swift/Sources/AnalyzerSwift/HardwareCheck.swift

public struct HardwareCheck {

    public enum HardwareError: Error, LocalizedError {
        case metalNotSupported

        public var errorDescription: String? {
            switch self {
            case .metalNotSupported:
                return """
                Metal GPU is not available on this system.

                Algiers requires Apple Silicon (M1 or later) for:
                • OpenL3 embedding generation (ANE)
                • Real-time audio processing (Metal)
                • Custom model training (Create ML)

                Intel Macs are not supported.
                """
            }
        }
    }

    public static func requireMetal() throws {
        guard let device = MTLCreateSystemDefaultDevice() else {
            throw HardwareError.metalNotSupported
        }

        // Log device info
        print("Metal device: \(device.name)")
        print("Unified memory: \(device.hasUnifiedMemory)")
        print("Recommended max working set: \(device.recommendedMaxWorkingSetSize / 1_000_000) MB")
    }
}
```

### Graceful Degradation

If ANE is unavailable (unlikely on supported hardware):

1. **SoundAnalysis** — Falls back to CPU (slower but functional)
2. **OpenL3** — Falls back to GPU, then CPU
3. **Custom training** — Requires ANE, will fail gracefully

---

## Privacy & Local-First

### Zero Cloud Dependencies

| Component | Cloud? | Notes |
|-----------|--------|-------|
| Audio decoding | ❌ Local | AVFoundation |
| DSP analysis | ❌ Local | Accelerate vDSP |
| OpenL3 inference | ❌ Local | Core ML on ANE |
| SoundAnalysis | ❌ Local | Apple framework |
| Custom training | ❌ Local | Create ML |
| Similarity search | ❌ Local | SQLite + Go |
| UI | ❌ Local | React + Vite |

### Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Your Mac (100% Local)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Audio Files ──▶ Swift Analyzer ──▶ SQLite DB ──▶ Go Engine ──▶ React UI   │
│       │              │                   │            │            │         │
│       │              │                   │            │            │         │
│       ▼              ▼                   ▼            ▼            ▼         │
│   /music/       Accelerate          /data/        HTTP/gRPC     Browser      │
│                 Core ML              algiers.db                              │
│                                                                              │
│  ═══════════════════════════════════════════════════════════════════════   │
│                         NO NETWORK TRAFFIC                                   │
│  ═══════════════════════════════════════════════════════════════════════   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### What Never Leaves Your Mac

- ✅ Audio files
- ✅ Embeddings
- ✅ Analysis results
- ✅ Training data
- ✅ Custom models
- ✅ Set plans
- ✅ User preferences

---

## Performance Benchmarks

### Analysis Time (5-minute track, 48kHz stereo)

| Operation | M1 | M1 Pro | M3 | M3 Max |
|-----------|-----|--------|-----|--------|
| **Audio decode** | 1.8s | 1.2s | 1.1s | 0.4s |
| **Beatgrid** | 2.4s | 1.8s | 1.5s | 0.9s |
| **Key detection** | 0.8s | 0.6s | 0.5s | 0.3s |
| **Energy** | 0.3s | 0.2s | 0.2s | 0.1s |
| **Loudness (R128)** | 0.5s | 0.4s | 0.3s | 0.2s |
| **OpenL3** | 2.1s | 1.4s | 0.9s | 0.5s |
| **SoundAnalysis** | 0.5s | 0.4s | 0.3s | 0.2s |
| **Sections/cues** | 0.3s | 0.2s | 0.2s | 0.1s |
| **Total** | **8.2s** | **6.1s** | **4.9s** | **3.2s** |

### Training Time

| Samples | M1 | M1 Pro | M3 |
|---------|-----|--------|-----|
| 100 | ~30s | ~20s | ~15s |
| 500 | ~2min | ~1.5min | ~1min |
| 1000 | ~4min | ~3min | ~2min |

### Memory Usage

| Component | Memory |
|-----------|--------|
| OpenL3 model | ~50 MB |
| Custom model | ~2-5 MB |
| Training session | ~500 MB |
| Inference | ~100 MB |
| Full analysis | ~200 MB |

---

## File Locations

### Swift ML Components

| Component | Path |
|-----------|------|
| OpenL3 Embedder | `analyzer-swift/Sources/AnalyzerSwift/ML/OpenL3Embedder.swift` |
| DJ Section Classifier | `analyzer-swift/Sources/AnalyzerSwift/ML/DJSectionClassifier.swift` |
| DJ Section Trainer | `analyzer-swift/Sources/AnalyzerSwift/ML/DJSectionTrainer.swift` |
| Sound Classifier | `analyzer-swift/Sources/AnalyzerSwift/Analysis/SoundClassifier.swift` |
| Hardware Check | `analyzer-swift/Sources/AnalyzerSwift/HardwareCheck.swift` |
| OpenL3 Model | `analyzer-swift/Sources/AnalyzerSwift/Resources/OpenL3Music.mlpackage` |

### Go Backend

| Component | Path |
|-----------|------|
| Similarity Algorithm | `internal/similarity/similarity.go` |
| Analysis Storage | `internal/storage/analysis.go` |
| Training Storage | `internal/storage/training.go` |
| HTTP API | `internal/httpapi/httpapi.go` |
| gRPC Server | `internal/server/grpc.go` |

### Database Schema

| Table | Purpose |
|-------|---------|
| `analyses` | Track analysis results + embeddings |
| `sound_events` | SoundAnalysis detected events |
| `qa_flags` | Quality assurance flags |
| `openl3_embeddings` | Window-level embeddings (optional) |
| `training_labels` | User-provided section labels |
| `training_jobs` | Training job status and metrics |
| `model_versions` | Custom model version history |

---

## API Reference

### Similarity Search

```http
GET /api/tracks/{id}/similar
GET /api/tracks/{id}/similar?limit=10
GET /api/tracks/{id}/similar?min_score=0.7
```

Response:
```json
{
  "query": {
    "track_id": 123,
    "title": "Get Lucky",
    "artist": "Daft Punk",
    "bpm": 116.0,
    "key": "8A",
    "energy": 7
  },
  "similar": [
    {
      "track_id": 456,
      "title": "Lose Yourself to Dance",
      "artist": "Daft Punk",
      "score": 0.92,
      "explanation": "similar vibe (89%); tempo match; same key; same energy",
      "vibe_match": 89.2,
      "tempo_match": 100.0,
      "key_match": 100.0,
      "energy_match": 100.0,
      "bpm_delta": 0.0,
      "key_relation": "same"
    }
  ]
}
```

### ML Settings

```http
GET /api/ml/settings
PUT /api/ml/settings

{
  "sound_analysis_enabled": true,
  "openl3_enabled": true,
  "dj_section_model_enabled": false,
  "active_model_version": 3,
  "show_explanations": true
}
```

### Training API

See [ML-TRAINING.md](ML-TRAINING.md) for complete training API reference.

---

## Research Background

### OpenL3

OpenL3 is based on the Look, Listen and Learn (L³) research:

> **"Look, Listen and Learn"**
> Arandjelović & Zisserman, ICCV 2017
> https://arxiv.org/abs/1705.08168

The model learns audio representations by predicting whether audio and video are temporally aligned. This self-supervised approach creates embeddings that capture perceptual similarity.

### Audio Embeddings for Music

Additional research informing Algiers' approach:

- **VGGish** — Audio feature extraction for AudioSet (Google)
- **musicnn** — Music-specific CNN for tagging (MTG Barcelona)
- **CLMR** — Contrastive learning for music representations

### DJ-Specific Analysis

DJ analysis techniques draw from:

- **Beat tracking** — Ellis, 2007; Böck et al., 2015
- **Key detection** — Krumhansl-Schmuckler profiles; Essentia
- **Section detection** — Structural segmentation (MIREX)
- **Loudness** — EBU R128 broadcast standard

---

<div align="center">

**All ML processing happens on your Mac.**

*Apple Neural Engine powered. Private by design. No cloud required.*

</div>
