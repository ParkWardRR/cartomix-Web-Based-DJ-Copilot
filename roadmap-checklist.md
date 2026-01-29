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

- [ ] Decode stage (ffmpeg fallback) to fixed SR PCM; loudness (EBU R128) + true peak.
- [ ] Beatgrid: static + dynamic tempo map, confidence score; monotonicity guarantees.
- [ ] Section detection (intro/break/build/drop/outro) + transition windows.
- [ ] Key detection + Camelot mapping; energy global + sectional curve.
- [ ] Cue generator (max 8, beat-aligned, safety bounds).
- [ ] Embeddings/similarity vector for "vibe" continuity (stub model first).

## Swift accel worker

- [ ] Implement Accelerate STFT + onset features; Core ML inference path; Metal/MPS optional.
- [ ] gRPC server inside analyzer-swift; error/timeout handling; process-supervision hooks.
- [ ] Fallback CPU path verified on non-Metal hosts.

## Set planning

- [x] Edge scoring (key/tempo/energy/window overlap + constraints).
- [x] Path solver variants (best full ordering, longest good path); deterministic seeded runs.
- [x] Explanations per edge/node.

## Exports

- [x] Generic: M3U8, analysis JSON, cues CSV with beat/time indices.
- [x] Checksum manifest (SHA256) + tar.gz bundle for verified handoff.
- [ ] Vendor: Rekordbox, Serato, Traktor writers; round-trip sanity with sample libraries.
- [ ] Export CLI/API options and UI download links.

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
- [ ] gRPC-web or HTTP bridge; optimistic UI for cue edits; undo/redo.

## Fixtures & testing

- [x] Flesh fixturegen to render WAVs per spec (click ladder, swing, tempo ramp, harmonic pad) + manifest JSON.
- [ ] Flesh fixturegen to render WAVs per spec (phrase track, harmonic set, club noise) + golden JSON.
- [x] Unit tests: beatgrid math, scoring, DB migrations (Go side); Swift XCTest pending for DSP kernels.
- [ ] Property tests: monotonic beats, cue bounds, export/import round-trip.
- [ ] Integration: golden comparisons on fixture corpus.
- [ ] E2E: Playwright-Go flows (import→analyze→cues→set→rehearse→export) across Chromium/WebKit; theme toggle tests.
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
- Pro UI with D3.js visualizations
- Canvas waveform renderer
- Real-time spectrum analyzer
- Energy arc and transition graph
- Live stats dashboard
- Dark mode default

### Next (Beta)
- Swift analyzer with Accelerate DSP
- Core ML integration for ANE inference
- Beatgrid and key detection algorithms
- Web Audio playback integration
- gRPC-web bridge for real data
