<div align="center">

# ML Training Guide

**Train custom DJ section classifiers on your Mac — all local, ANE-accelerated**

[![Custom Training](https://img.shields.io/badge/Custom%20Training-F59E0B?style=for-the-badge)](#overview)
[![Create ML](https://img.shields.io/badge/Create%20ML-34C759?style=for-the-badge)](#how-it-works)
[![Neural Engine](https://img.shields.io/badge/Neural%20Engine-FF9500?style=for-the-badge)](#performance)
[![Local Only](https://img.shields.io/badge/100%25%20Local-222222?style=for-the-badge)](#privacy)

</div>

---

## Table of Contents

- [Overview](#overview)
- [3-Layer ML Architecture](#3-layer-ml-architecture)
- [DJ Section Labels](#dj-section-labels)
- [Training Workflow](#training-workflow)
- [Training UI Guide](#training-ui-guide)
- [REST API Reference](#rest-api-reference)
- [Model Management](#model-management)
- [Performance](#performance)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Future Enhancements](#future-enhancements)

---

## Overview

Algiers supports **opt-in custom model training** for DJ-specific section classification. Unlike generic audio classifiers, you can train a model that recognizes sections the way *you* label them.

### Key Features

| Feature | Description |
|---------|-------------|
| **Local Training** | All training happens on your Mac — audio never leaves |
| **ANE Acceleration** | Uses Apple Neural Engine for fast training and inference |
| **Create ML Backend** | Built on Apple's MLSoundClassifier for audio |
| **Model Versioning** | Keep multiple versions, rollback anytime |
| **Explainable** | Every classification includes confidence scores |

### When to Use Custom Training

| Use Case | Recommendation |
|----------|----------------|
| **Standard DJ prep** | Built-in section detection is sufficient |
| **Specific genre** | Train on your genre for better accuracy |
| **Unique labeling style** | Your "drop" might be someone else's "break" |
| **Production workflow** | Consistent labeling across large libraries |

---

## 3-Layer ML Architecture

Algiers uses a layered approach to ML, each layer building on the previous:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Algiers ML Architecture                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │                         Layer 1                                     │   │
│   │              Apple SoundAnalysis (Built-in)                         │   │
│   │                                                                     │   │
│   │  • 300+ pre-trained audio labels                                   │   │
│   │  • Zero configuration required                                      │   │
│   │  • Music/speech/noise detection                                     │   │
│   │  • QA flag generation                                               │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │                         Layer 2                                     │   │
│   │                  OpenL3 Embeddings (512-dim)                        │   │
│   │                                                                     │   │
│   │  • Pre-trained on millions of audio-video pairs                    │   │
│   │  • Captures timbre, texture, mood                                  │   │
│   │  • Powers similarity search                                         │   │
│   │  • "Vibe matching" for transitions                                  │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                                    ▼                                        │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │                         Layer 3                                     │   │
│   │            Custom DJ Section Classifier (Opt-in)                    │   │
│   │                                                                     │   │
│   │  • Train on YOUR labeled data                                       │   │
│   │  • 7 DJ-specific section labels                                     │   │
│   │  • Uses Create ML MLSoundClassifier                                 │   │
│   │  • Model versioning with rollback                                   │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│   ┌────────────────────────────────────────────────────────────────────┐   │
│   │                    All Layers Run on ANE                            │   │
│   │              Apple Neural Engine — 100% Local                       │   │
│   └────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Layer Comparison

| Layer | Model | Training | Use Case |
|-------|-------|----------|----------|
| **1 - SoundAnalysis** | Apple built-in | None | Context detection, QA |
| **2 - OpenL3** | Pre-trained | None | Similarity search |
| **3 - Custom** | Your data | Required | Section classification |

---

## DJ Section Labels

Algiers uses 7 DJ-specific section labels:

| Label | Color | Description | Example |
|-------|-------|-------------|---------|
| **intro** | 🟢 `#22c55e` | Track opening, minimal elements | First 16-32 bars, drums only |
| **build** | 🟡 `#eab308` | Energy rising, tension building | Riser, filter sweep, snare roll |
| **drop** | 🔴 `#ef4444` | Main energy peak, full arrangement | Bass drops, full drums, synths |
| **break** | 🟣 `#a855f7` | Breakdown, stripped back | No drums, melodic interlude |
| **outro** | 🔵 `#3b82f6` | Track ending, elements fading | Last 16-32 bars, drums only |
| **verse** | ⬛ `#4b5563` | Vocal or melodic verse | Verse lyrics, melodic phrase |
| **chorus** | 🩷 `#ec4899` | Hook or main phrase | Catchy hook, main vocal |

### Label Guidelines

**Good labeling:**
- Consistent start/end on beat boundaries
- Clear section transitions
- At least 8 beats per section

**Poor labeling:**
- Sections that span multiple types
- Off-beat boundaries
- Very short sections (<4 beats)

---

## Training Workflow

### Step 1: Label Tracks

Add section labels to your tracks through the Training UI:

```
┌─────────────────────────────────────────────────────────────────┐
│  Label Editor                                    Track: Drop.wav│
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐│
│  │Intro │ │Build │ │ Drop │ │Break │ │Outro │ │Verse │ │Chorus││
│  └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘ └──────┘│
│                        ▲                                        │
│                    selected                                      │
│                                                                  │
│  Start Beat: [  64  ]     End Beat: [ 128  ]                   │
│                                                                  │
│  [        Add Label        ]                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Step 2: Check Statistics

Review your label distribution in the stats grid:

```
┌─────────────────────────────────────────────────────────────────┐
│  Training Dataset                                    142 labels │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐            │
│  │ Intro   │  │ Build   │  │  Drop   │  │ Break   │            │
│  │ 25/10 ✓ │  │ 22/10 ✓ │  │ 28/10 ✓ │  │ 18/10 ✓ │            │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘            │
│                                                                  │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐                         │
│  │ Outro   │  │ Verse   │  │ Chorus  │                         │
│  │ 21/10 ✓ │  │ 15/10 ✓ │  │ 13/10 ✓ │                         │
│  └─────────┘  └─────────┘  └─────────┘                         │
│                                                                  │
│  ✓ Ready for training (all classes have 10+ samples)           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**Minimum requirements:**
- At least 10 samples per class
- At least 2 different label types
- Recommended: 20+ samples per class for good accuracy

### Step 3: Start Training

Click "Start Training" to begin:

```
┌─────────────────────────────────────────────────────────────────┐
│  Model Training                                      training   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ████████████████████████░░░░░░░░░░░░░░  60%                   │
│                                                                  │
│  Epoch 6 / 10                                                   │
│  Loss: 0.2847                                                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

Training stages:
1. **Pending** — Job queued
2. **Preparing** — Extracting audio features
3. **Training** — Create ML model training
4. **Evaluating** — Computing accuracy metrics
5. **Completed** — Model ready to activate

### Step 4: Evaluate Results

Review training results:

```
┌─────────────────────────────────────────────────────────────────┐
│  Model Training                                     completed   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Accuracy:      87.5%                                           │
│  F1 Score:      85.2%                                           │
│  Model Version: v3                                               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Step 5: Activate Model

Choose which model version to use for inference:

```
┌─────────────────────────────────────────────────────────────────┐
│  Model Versions                                                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  v3                                           Active    │   │
│  │  Accuracy: 87.5%    F1: 85.2%                          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  v2                                                     │   │
│  │  Accuracy: 82.1%    F1: 80.4%                          │   │
│  │  [ Activate ]  [ Delete ]                               │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  v1                                                     │   │
│  │  Accuracy: 75.3%    F1: 73.8%                          │   │
│  │  [ Activate ]  [ Delete ]                               │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Training UI Guide

Access the Training screen from the main navigation:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  ◈ Algiers                                                          alpha   │
├─────────────────────────────────────────────────────────────────────────────┤
│  [ Library ] [ Set Builder ] [ Graph ] [ Settings ] [ Training ]            │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Training Screen Layout

```
┌─────────────────────────────────────────┬──────────────────────────────────┐
│                                         │                                   │
│  Training Dataset            142 labels │  Add Label                       │
│                                         │  ──────────────────────          │
│  ┌───────────────────────────────────┐ │  Track: Summer Nights.wav        │
│  │  Stats Grid (label counts)        │ │                                   │
│  └───────────────────────────────────┘ │  [Intro] [Build] [Drop] ...      │
│                                         │                                   │
│  ┌───────────────────────────────────┐ │  Start: [64]  End: [128]         │
│  │  Track      Label    Beats Source │ │                                   │
│  ├───────────────────────────────────┤ │  [ Add Label ]                   │
│  │  Drop.wav   drop    64-128  user  │ │                                   │
│  │  Build.wav  build   32-64   user  │ ├──────────────────────────────────┤
│  │  Intro.wav  intro   0-32    user  │ │                                   │
│  │  ...                              │ │  Model Training        training  │
│  └───────────────────────────────────┘ │  ──────────────────────          │
│                                         │  ████████░░░░░░░░  60%          │
│                                         │  Epoch 6/10  Loss: 0.28         │
│                                         │                                   │
│                                         │  [ Start Training ]              │
│                                         │                                   │
│                                         ├──────────────────────────────────┤
│                                         │                                   │
│                                         │  Model Versions                  │
│                                         │  ──────────────────────          │
│                                         │  v3  87.5%  [Active]             │
│                                         │  v2  82.1%  [Activate] [Delete]  │
│                                         │                                   │
└─────────────────────────────────────────┴──────────────────────────────────┘
```

---

## REST API Reference

### Training Labels

#### Add Label
```http
POST /api/training/labels
Content-Type: application/json

{
  "track_id": 123,
  "label_value": "drop",
  "start_beat": 64,
  "end_beat": 128,
  "start_time_seconds": 32.0,
  "end_time_seconds": 64.0,
  "source": "user"
}
```

#### List Labels
```http
GET /api/training/labels
GET /api/training/labels?track_id=123
GET /api/training/labels?label=drop
```

Response:
```json
[
  {
    "id": 1,
    "track_id": 123,
    "content_hash": "abc123",
    "track_path": "/music/track.wav",
    "label_value": "drop",
    "start_beat": 64,
    "end_beat": 128,
    "start_time_seconds": 32.0,
    "end_time_seconds": 64.0,
    "source": "user",
    "created_at": "2026-01-29T12:00:00Z"
  }
]
```

#### Delete Label
```http
DELETE /api/training/labels/{id}
```

#### Get Statistics
```http
GET /api/training/labels/stats
```

Response:
```json
{
  "total_labels": 142,
  "label_counts": {
    "intro": 25,
    "build": 22,
    "drop": 28,
    "break": 18,
    "outro": 21,
    "verse": 15,
    "chorus": 13
  },
  "tracks_covered": 47,
  "avg_per_track": 3.02,
  "ready_for_training": true,
  "min_samples_required": 10
}
```

### Training Jobs

#### Start Training
```http
POST /api/training/start
```

Response:
```json
{
  "job_id": "job_1706550000000000",
  "message": "training job started"
}
```

#### List Jobs
```http
GET /api/training/jobs
GET /api/training/jobs?limit=10
```

#### Get Job Status
```http
GET /api/training/jobs/{job_id}
```

Response:
```json
{
  "id": 1,
  "job_id": "job_1706550000000000",
  "status": "training",
  "progress": 0.6,
  "current_epoch": 6,
  "total_epochs": 10,
  "current_loss": 0.2847,
  "accuracy": null,
  "f1_score": null,
  "model_path": null,
  "model_version": null,
  "error_message": null,
  "label_counts": {
    "intro": 25,
    "build": 22,
    "drop": 28
  },
  "started_at": "2026-01-29T12:00:00Z",
  "completed_at": null,
  "created_at": "2026-01-29T12:00:00Z"
}
```

### Model Versions

#### List Models
```http
GET /api/training/models
GET /api/training/models?type=dj_section
```

Response:
```json
[
  {
    "id": 3,
    "model_type": "dj_section",
    "version": 3,
    "model_path": "/models/dj_section_v3.mlmodelc",
    "accuracy": 0.875,
    "f1_score": 0.852,
    "is_active": true,
    "label_counts": {"drop": 28, "build": 22},
    "training_job_id": "job_1706550000",
    "created_at": "2026-01-29T12:30:00Z"
  }
]
```

#### Activate Model
```http
POST /api/training/models/{version}/activate
```

#### Delete Model
```http
DELETE /api/training/models/{version}
```

---

## Model Management

### Version Strategy

Each training run creates a new model version:

```
/models/
├── dj_section_v1.mlmodelc    # First training
├── dj_section_v1.json        # Metadata
├── dj_section_v2.mlmodelc    # Second training
├── dj_section_v2.json
└── dj_section_v3.mlmodelc    # Current active
```

### Rollback

To rollback to a previous version:

1. Open Training screen
2. Find the version you want
3. Click "Activate"
4. Previous version becomes active immediately

### Delete

To free disk space:

1. Open Training screen
2. Find the version to delete
3. Click "Delete"
4. **Note:** Cannot delete the active version

---

## Performance

### Training Time

| Samples | M1 | M1 Pro | M3 |
|---------|-----|--------|-----|
| 100 | ~30s | ~20s | ~15s |
| 500 | ~2min | ~1.5min | ~1min |
| 1000 | ~4min | ~3min | ~2min |

### Inference Time

| Operation | Latency |
|-----------|---------|
| Single section | ~5ms |
| Full track (5 min) | ~2.5s |
| Batch (10 tracks) | ~20s |

### Memory Usage

| Component | Memory |
|-----------|--------|
| Model (mlmodelc) | ~2-5 MB |
| Training session | ~500 MB |
| Inference | ~100 MB |

---

## Best Practices

### Data Quality

1. **Consistent labeling** — Same person should label all data
2. **Clear boundaries** — Start/end on beats
3. **Representative samples** — Include variety within each class
4. **Balanced classes** — Similar sample count per label

### Training Tips

1. **Start small** — 10-20 samples per class to test
2. **Iterate** — Train, evaluate, add more data
3. **Review errors** — Check which sections are misclassified
4. **Genre-specific** — Train separate models for different genres

### Model Selection

1. **Accuracy > 80%** — Good for production use
2. **F1 > 75%** — Balanced precision/recall
3. **Compare versions** — Test on new tracks before activating

---

## Troubleshooting

### Training Fails

| Error | Solution |
|-------|----------|
| "Need more labels" | Add at least 10 samples per class |
| "Need 2+ classes" | Label at least 2 different section types |
| "Model export failed" | Check disk space, restart |
| "Training timeout" | Reduce dataset size or check memory |

### Poor Accuracy

| Symptom | Solution |
|---------|----------|
| <60% accuracy | More samples, cleaner labels |
| Confuses drop/build | Label more distinct examples |
| Good train, bad test | Overfitting — more diverse samples |

### Model Not Loading

| Issue | Solution |
|-------|----------|
| "Model not found" | Check model path in settings |
| "Incompatible model" | Retrain with current Swift version |
| "ANE unavailable" | Falls back to CPU (slower) |

---

## Future Enhancements

### Planned Features

| Feature | Description | ETA |
|---------|-------------|-----|
| **Waveform painting** | Drag to select sections on waveform | v1.1 |
| **Auto-suggestions** | Model suggests labels, you confirm | v1.1 |
| **Transfer learning** | Pre-trained base for faster training | v1.2 |
| **Cross-validation** | K-fold CV for better accuracy estimates | v1.2 |
| **Export models** | Share models with other DJs | v1.3 |

### Research Directions

- **Self-supervised pre-training** — Train on unlabeled data first
- **Multi-task learning** — Energy + sections in one model
- **Attention visualization** — Show what the model "listens" to
- **Temporal consistency** — Smooth predictions across time

---

<div align="center">

**Train smarter. Mix better.**

*Your labels, your model, your Mac — 100% local.*

</div>
