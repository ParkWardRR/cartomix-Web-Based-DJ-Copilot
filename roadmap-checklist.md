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
- [ ] Embeddings/similarity vector for "vibe" continuity (stub model first).
- [ ] Loudness (EBU R128) + true peak.

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
- [ ] Core ML inference path; Metal/MPS optional.
- [ ] gRPC server with proto stubs; error/timeout handling; process-supervision hooks.
- [ ] Fallback CPU path verified on non-Metal hosts.

## Set planning

- [x] Edge scoring (key/tempo/energy/window overlap + constraints).
- [x] Path solver variants (best full ordering, longest good path); deterministic seeded runs.
- [x] Explanations per edge/node.

## Exports

- [x] Generic: M3U8, analysis JSON, cues CSV with beat/time indices.
- [x] Checksum manifest (SHA256) + tar.gz bundle for verified handoff.
- [x] Vendor: Rekordbox, Serato, Traktor writers; round-trip sanity with sample libraries.
- [x] Export CLI/API options and UI download links.

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
- [ ] Audio preview: Web Audio + AudioWorklet player; streamed waveform tiles/PCM from engine.
- [x] gRPC-web or HTTP bridge; optimistic UI for cue edits; undo/redo.

## Fixtures & testing

- [x] Flesh fixturegen to render WAVs per spec (click ladder, swing, tempo ramp, harmonic pad) + manifest JSON.
- [ ] Flesh fixturegen to render WAVs per spec (phrase track, harmonic set, club noise) + golden JSON.
- [x] Unit tests: beatgrid math, scoring, DB migrations (Go side); Swift XCTest pending for DSP kernels.
- [x] Property tests: monotonic beats, cue bounds, export/import round-trip.
- [x] Integration: golden comparisons on fixture corpus.
- [x] E2E: Playwright-Go flows (import→analyze→cues→set→rehearse→export) across Chromium/WebKit; theme toggle tests.
- [ ] CI: run go/swift/unit/property + E2E (WebKit on macOS runner).

## DevX & quality gates

- [x] Linting: golangci-lint config; swift-format/swiftlint; eslint/prettier.
- [x] Makefile targets: proto, lint, test, e2e, fixturegen.
- [ ] Pre-commit hooks or CI checks for generated artifacts in sync.
- [x] Logging/metrics stubs; crash/retry for analyzer worker.

## Packaging & alpha readiness

- [x] Provide local run scripts: engine + analyzer-swift + ui dev server.
- [x] Pro UI with visualizations ready for alpha demo.
- [x] Updated README with alpha features, architecture, and changelog.
- [ ] Versioning/analysis cache migration strategy; backup/export of DB.
- [ ] Minimal docs: quickstart, API surface, test corpus instructions.
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
- gRPC server integration for Swift analyzer
- Web Audio playback UI integration (hook complete, UI pending)
- Embeddings/similarity vector for vibe continuity
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
