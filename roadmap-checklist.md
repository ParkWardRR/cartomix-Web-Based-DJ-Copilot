# Algiers Development Roadmap

## Proto & API plumbing

- [x] Generate Go stubs (protoc or buf), commit generated code; add make proto.
- [x] Stand up engine gRPC server: register services, health checks, config flags, logging.
- [x] Implement analyzer-worker client (gRPC) and a CPU fallback analyzer for early runs.
- [x] Add API auth toggle (local-only default open; stub for future tokens).

## Ingest & storage

- [x] File scanner (recursive, supported formats, hash cache) + resumable job queue.
- [x] SQLite schema migrations (tracks, analyses, cue edits, graph edges, blobs for tiles).
- [x] Content-addressed blob store for waveform tiles / embeddings.

## Analysis pipeline

- [x] Decode stage (AVFoundation) to fixed SR PCM; sample rate conversion.
- [x] Beatgrid: static tempo map, confidence score; monotonicity guarantees via onset detection + autocorrelation.
- [x] Section detection (intro/verse/build/drop/breakdown/outro) + transition windows.
- [x] Key detection (Krumhansl-Schmuckler chroma profiles) + Camelot/Open Key mapping; energy global + sectional curve.
- [x] Cue generator (max 8, beat-aligned, priority-based with safety bounds).
- [x] Embeddings/similarity vector for "vibe" continuity (128-dim MFCC-like features with cosine similarity).
- [x] Loudness (EBU R128) + true peak (K-weighting filters, integrated/momentary/short-term loudness).

## Swift accel worker

- [x] Implement Accelerate STFT + onset features (FFTProcessor with vDSP).
- [x] AudioDecoder using AVFoundation with sample rate conversion.
- [x] BeatgridDetector with spectral flux onset detection and autocorrelation tempo estimation.
- [x] KeyDetector with chroma features and Krumhansl-Schmuckler key profiles.
- [x] EnergyAnalyzer with band analysis (low/mid/high) and energy curve.
- [x] SectionDetector with energy-based segmentation.
- [x] CueGenerator with priority-based cue selection and beat alignment.
- [x] Main Analyzer orchestrator with progress callbacks.
- [x] CLI executable (analyzer-swift) with analyze/serve/healthcheck commands.
- [x] HTTP server for analyzer API (JSON-based).
- [x] Core ML inference path; Metal required with graceful error.
- [x] gRPC server with proto stubs; error/timeout handling; process-supervision hooks.
- [x] Metal/ANE requirement check - fails gracefully on unsupported hardware.

## Set planning

- [x] Edge scoring (key/tempo/energy/window overlap + constraints).
- [x] Path solver variants (best full ordering, longest good path); deterministic seeded runs.
- [x] Explanations per edge/node.

## Exports

- [x] Generic: M3U8, analysis JSON, cues CSV with beat/time indices.
- [x] Checksum manifest (SHA256) + tar.gz bundle for verified handoff.
- [x] Vendor: Rekordbox, Serato, Traktor writers; round-trip sanity with sample libraries.
- [x] Export CLI/API options and UI download links.

## ML Features (Apple-First)

- [x] OpenL3 512-dimensional embeddings (Core ML, ANE-accelerated).
- [x] Similarity search with weighted scoring (vibe/tempo/key/energy).
- [x] Human-readable explanation strings for all AI decisions.
- [x] Apple SoundAnalysis built-in classifier (Layer 1 - 300+ labels).
- [x] Sound context detection (music/speech/noise) with confidence.
- [x] QA flag generation (needs_review, mixed_content, speech_detected).
- [x] Custom DJ section model training (Layer 3 - opt-in).
- [x] Training data collection UI with waveform label editor.
- [x] Model versioning and rollback support.

## Web UI (TS/React)

- [x] Toolchain setup (Node 22, Vite/React 19, TS strict, eslint/prettier config).
- [x] Theme system (Dark default, Auto via prefers-color-scheme, manual toggle).
- [x] High-performance visualization libraries (D3.js 7, Framer Motion 11, React Virtuoso 4, Zustand 5).
- [x] Screens: Library grid + filters; Track detail (waveform, beat markers, sections, cues); Set builder (order list + reasons + energy arc); Transition rehearsal (dual deck preview).
- [x] Canvas waveform renderer with section overlays, cue markers, beat grid, playhead.
- [x] Real-time spectrum analyzer (mirror/bars/circular modes).
- [x] Energy arc visualization with animated SVG.
- [x] Transition graph (D3.js force-directed) with zoom/pan/drag.
- [x] Live stats dashboard with animated counters and progress rings.
- [x] BPM/key distribution charts.
- [x] Three-view layout: Library, Set Builder, Graph View.
- [x] Four-view layout: Library, Set Builder, Graph, Settings.
- [x] Audio preview: Web Audio + AudioWorklet player; streamed waveform tiles/PCM from engine.
- [x] gRPC-web or HTTP bridge; optimistic UI for cue edits; undo/redo.
- [x] ML Settings view with feature toggles and hardware status.
- [x] Similar Tracks panel with explainable scoring.
- [x] Analysis Panel with DSP and ML results display.

## Fixtures & testing

- [x] Flesh fixturegen to render WAVs per spec (click ladder, swing, tempo ramp, harmonic pad) + manifest JSON.
- [x] Flesh fixturegen to render WAVs per spec (phrase track, harmonic set, club noise) + golden JSON.
- [x] Unit tests: beatgrid math, scoring, DB migrations (Go side); Swift XCTest pending for DSP kernels.
- [x] Property tests: monotonic beats, cue bounds, export/import round-trip.
- [x] Integration: golden comparisons on fixture corpus.
- [x] E2E: Playwright-Go flows (import→analyze→cues→set→rehearse→export) across Chromium/WebKit; theme toggle tests.
- [ ] CI: run go/swift/unit/property + E2E (WebKit on macOS runner).
  - In progress: go + Swift + React build workflows running on CI (macOS + Ubuntu); WebKit E2E still pending.

## DevX & quality gates

- [x] Linting: golangci-lint config; swift-format/swiftlint; eslint/prettier.
- [x] Makefile targets: proto, lint, test, e2e, fixturegen.
- [ ] Pre-commit hooks or CI checks for generated artifacts in sync.
- [x] Logging/metrics stubs; crash/retry for analyzer worker.

## Packaging & alpha readiness

- [x] Provide local run scripts: engine + analyzer-swift + ui dev server.
- [x] Pro UI with visualizations ready for alpha demo.
- [x] Updated README with alpha features, architecture, and changelog.
- [x] Versioning/analysis cache migration strategy; backup/export of DB.
- [x] Minimal docs: quickstart, API surface, test corpus instructions.
- [ ] Alpha acceptance checklist: import 100 tracks, analyze, build 30-track set, rehearse transitions, export to Rekordbox/Serato/Traktor without manual metadata fixes.

---

## Alpha Release Summary (v0.1.0-alpha - 2026-01-29)

### Completed
- gRPC engine with health checks and graceful shutdown
- Library scanner with SHA256 content hashing
- SQLite storage with migrations and WAL mode
- Set planner with weighted graph and explainable scoring
- Generic exports (M3U8, JSON, CSV, checksums, tar.gz)
- Vendor exports (Rekordbox XML, Serato crate, Traktor NML)
- HTTP REST API bridge for web UI integration
- Property tests for planner algorithms
- E2E test framework with Playwright-Go
- Pro UI with D3.js visualizations
- Canvas waveform renderer
- Real-time spectrum analyzer
- Energy arc and transition graph
- Live stats dashboard
- Dark mode default

### Next (Beta)
- Core ML integration for ANE inference
- gRPC server integration for Swift analyzer (HTTP server complete)
- CI pipeline for macOS (go/swift/unit/property + E2E)

### Recently Completed (Post-Alpha)
- Connected React UI to HTTP API with Zustand state management
- Vite dev proxy for API requests
- Graceful fallback to demo mode when API unavailable
- Export panel in Set Builder UI with Rekordbox/Serato/Traktor format selection
- Full vendor export integration (UI + API)
- Golden comparison tests for all vendor export formats
- Web Audio useAudioPlayer hook with Web Audio API

### Swift Analyzer (v0.2.0-beta - Completed)
- Accelerate-based FFT processor with STFT, spectral flux, chroma features
- AVFoundation audio decoder with sample rate conversion
- Beatgrid detector using onset detection and autocorrelation
- Key detector with Krumhansl-Schmuckler profiles and Camelot mapping
- Energy analyzer with band analysis (low/mid/high frequencies)
- Section detector (intro/verse/build/drop/breakdown/outro)
- Cue generator with priority-based selection and beat alignment
- CLI tool (analyzer-swift analyze/serve/healthcheck)
- HTTP API server for analyzer integration

### Testing & Quality (v0.2.0-beta)
- Golden comparison tests for Rekordbox XML export
- Golden comparison tests for Serato crate export
- Golden comparison tests for Traktor NML export
- Export round-trip validation with checksums
- Web Audio useAudioPlayer hook for browser playback

### Audio Analysis (v0.3.0-beta)
- EBU R128 loudness analyzer with K-weighting filters
- Integrated, momentary, short-term loudness + true peak measurement
- 128-dimensional audio embeddings with MFCC-like features
- Vibe similarity scoring for DJ set continuity
- Spectral centroid, rolloff, flatness, and harmonic ratio features
- Web Audio integration into TrackDetail UI component

### ML Workflow (v0.4.0-beta)
- OpenL3 512-dimensional embeddings with Core ML (Layer 2)
- Similarity search API with weighted scoring (vibe 50%, tempo 20%, key 20%, energy 10%)
- Human-readable explanation strings for AI decisions
- UI: SimilarTracks component with animated score bars
- UI: ModelSettings component with ML feature toggles
- UI: AnalysisPanel component for DSP and ML results
- Apple SoundAnalysis integration (Layer 1) - Swift classifier
- Sound event detection (music/speech/noise categories)
- QA flag generation (needs_review, mixed_content, low_confidence, speech_detected)
- Primary context detection with confidence scoring

### Custom Training (v0.5.0-beta) - Layer 3
- DJSectionTrainer with Create ML MLSoundClassifier integration
- DJSectionClassifier for inference with custom models
- 7 DJ section labels: intro, build, drop, break, outro, verse, chorus
- Training job management with progress tracking
- Model versioning with activate/rollback support
- Training data REST API (CRUD for labels, jobs, models)
- UI: TrainingScreen with dataset table and label statistics
- UI: LabelEditor for adding section labels to tracks
- UI: TrainingProgressCard with epoch/loss visualization
- UI: ModelVersionsList with activate/delete actions
- SQLite schema for training_labels, training_jobs, model_versions

### Aurora UI & Distribution (v1.6-beta - 2026-02-01)
- Glassmorphic "Aurora" skin with hero HUD, neon grid, and elevated panels.
- Quick filter chips for analyzed/high-energy/review crates plus energy hero lane.
- Local-only validation (UI build, Go tests, Swift tests); GitHub CI disabled per cost constraints.
- Notarized DMG v1.6, README/screenshots refreshed for release.

### Command Palette & HUD Polish (v1.7-beta - 2026-02-01)
- Global command palette (⌘K) for navigation, filters, batch actions, charts, and theme toggle.
- Refined aurora HUD: journey progress meter, crisp stats, tightened chip filters.
- Local pipeline only; notarized DMG v1.7 released.

---

## Future Roadmap (Recommended)

### v1.0 Release - Core Stability
- [x] Full gRPC API implementation (ML, Training, Similarity, Model services)
- [x] gRPC interceptors (logging, metrics, recovery)
- [x] HTTP REST deprecation headers (Sunset: July 2026)
- [x] gRPC migration documentation (docs/gRPC-MIGRATION.md)
- [ ] gRPC-web for browser clients
- [x] gRPC streaming progress events (job %, current file, stage timings, ETA, byte progress)
- [ ] Cancel/timeout support end-to-end for long scans
- [ ] Waveform-based label painting (drag-to-select sections)
- [ ] Alpha acceptance: 100 tracks → 30-track set → export
- [ ] CI/CD pipeline with macOS runner (Go/Swift/Playwright)
- [ ] Pre-commit hooks for generated artifact validation

### v1.1 - Editable Analysis (AI Assist)
*"AI suggestions → AI assist" - Trust through user control*

- [ ] **Override layer**: User-editable beatgrid, key, cues, sections, loudness gain
- [ ] **Lock flags**: Prevent re-analysis from overwriting user edits
- [ ] **Analysis versioning**: Store `analyzer_version`, `model_version`, `params` (hop size, thresholds)
- [ ] **Content hash of decoded PCM**: Enable cache reuse across runs/machines
- [ ] **Partial-stage caching**: Cache decode→features→embedding→planner outputs separately
- [ ] Auto-labeling with active learning (model suggests, user confirms)
- [ ] Transfer learning with pre-trained base model
- [ ] K-fold cross-validation for accuracy estimates
- [ ] Confusion matrix visualization for training results
- [ ] Label import from Rekordbox/Serato cue points

### v1.2 - Scalable Similarity
*Fast ANN retrieval + explainable reranking*

- [ ] **Two-stage similarity**: (1) Fast ANN on embeddings, (2) Rerank with BPM/key/energy + explanation
- [ ] **Section-level embeddings**: Compute embeddings per intro/build/drop/outro, not just track mean
- [ ] **Transition window matching**: Find mixable moments via section-level similarity
- [ ] **Explain similarity view**: Show top contributing windows/sections + confidence bands
- [ ] **Selectable OpenL3 configs**: `content_type` (music/env), `input_repr` (mel128/mel256), embedding size
- [ ] Energy curve matching (find tracks with compatible arcs)
- [ ] Mood detection (happy/sad/dark/euphoric)
- [ ] Fine-grained genre embedding
- [ ] ANN indexing for >1000 track libraries (HNSW or similar)

### v1.3 - Export Verification & Reliability
*Bulletproof exports + production-grade reliability*

- [ ] **Export round-trip tests**: Automated export→import/parse→compare for Rekordbox/Traktor/Serato
- [ ] **Golden export fixtures**: Versioned test fixtures to catch silent corruption
- [ ] **Schema validation**: Validate XML/NML/crate against official schemas
- [ ] **Cue templates**: "First downbeat", "bass-in", "breakdown", "mix-out" + per-genre defaults
- [ ] **Global scheduler**: Per-engine concurrency limits (decode N, DSP M, ML K) + backpressure
- [ ] **Analysis bundles**: Save params JSON, summary metrics, downsampled features per track
- [ ] Real-time streaming analysis
- [ ] Hardware control surface (MIDI mapping)
- [ ] Cross-device model sync
- [ ] Model export/import for sharing
- [ ] Versioning/backup strategy for SQLite

### v2.0 - Advanced AI
- [ ] AI-generated transition clips
- [ ] Stem separation for transition layers
- [ ] Crowd energy prediction from live recordings
- [ ] Voice-controlled set prep ("find something higher energy")
- [ ] Collaborative playlist sharing

### Research Directions
- [ ] Self-supervised pre-training on unlabeled audio
- [ ] Multi-task learning (energy + sections + mood)
- [ ] Attention visualization (what does the model hear?)
- [ ] Temporal smoothing for section predictions
- [ ] Beat-grid alignment in similarity scoring
- [ ] Window-level similarity attribution (explain which time ranges match)

---

## Highest-Leverage Upgrades Summary

| Area | Upgrade | Why | Effort |
|------|---------|-----|--------|
| AI Assist | Override layer + lock flags | Trust through user control on weird tracks | Med |
| Determinism | Version artifacts + content hash | Debuggable, safe cache reuse | Low-Med |
| Similarity | Two-stage ANN + rerank | Scales to large libraries while keeping explanations | Med-High |
| Progress | Stream progress + cancel | Long scans feel broken without feedback | Med |
| Exports | Round-trip tests + golden fixtures | Export bugs are the most painful class | Med |

## Audio/ML Improvements (Practical)

| Problem | Improvement | Notes |
|---------|-------------|-------|
| One embedding per track loses transitions | Section-level embeddings per transition window | OpenL3 is window-based; extending to sections yields better mix-point matches |
| Vibe match feels opaque | Explain similarity view with top contributing windows | Can attribute similarity back to time ranges |
| Model tradeoffs | Selectable OpenL3 configs | Power users can tune for their genre/library |
| Cue usefulness | Cue templates with genre defaults | Common DJ prep cue points (first beat, bass-in, breakdown, mix-out) |

## Reliability/Performance

| Concern | Improvement | Benefit |
|---------|-------------|---------|
| Resource contention | Global scheduler with backpressure | Prevents collapse on 500+ tracks |
| Cache safety | Cache by (content_hash, analyzer_version, params_hash) | Resume + avoid recomputing unchanged steps |
| Debuggability | Analysis bundles per track | Actionable bug reports without original audio |

---

## Screenshot Placeholders

The following screenshots should be captured for documentation:

| Screenshot | Location | Description |
|------------|----------|-------------|
| `training-screen.png` | docs/assets/screens/ | Full training screen with all panels |
| `training-stats.png` | docs/assets/screens/ | Label statistics grid showing counts |
| `training-progress.png` | docs/assets/screens/ | Training in progress with progress bar |
| `training-complete.png` | docs/assets/screens/ | Completed training with accuracy/F1 |
| `model-versions.png` | docs/assets/screens/ | Model version list with active badge |
| `label-editor.png` | docs/assets/screens/ | Label editor with selected track |
| `analysis-panel.png` | docs/assets/screens/ | Analysis panel with DSP/ML results |
| `similar-tracks.png` | docs/assets/screens/ | Similar tracks panel with scores |
| `model-settings.png` | docs/assets/screens/ | ML settings with feature toggles |

### Animated GIFs (Recommended)

| GIF | Description |
|-----|-------------|
| `training-workflow.gif` | Full workflow: label → train → evaluate → activate |
| `similarity-demo.gif` | Select track → see similar tracks with explanations |
| `section-detection.gif` | Real-time section detection on waveform |
| `model-rollback.gif` | Activate old version, rollback demonstration |

To capture screenshots:
```bash
make screenshots
# or
go run ./cmd/screenshots
```
