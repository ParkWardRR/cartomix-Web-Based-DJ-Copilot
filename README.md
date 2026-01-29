# DJ Set Prep Copilot

![status](https://img.shields.io/badge/status-planning-blue)
![license](https://img.shields.io/badge/license-Blue_Oak_1.0.0-lightgrey)
![stack](https://img.shields.io/badge/stack-Go_1.22%2B_%7C_Swift_6_%7C_TS%2FReact-3e7ea8)
![tests](https://img.shields.io/badge/tests-Go_%7C_Swift_XCTest_%7C_Playwright--Go-ff69b4)

A local‑first copilot for DJ set prep: it analyzes your library, surfaces mixable sections, proposes cue points and transition windows, and optimizes set order with explainable scoring—while keeping you in control of the mix.

## What we are building
- Library ingest with resumable background analysis (WAV/AIFF/MP3/AAC/ALAC/FLAC).
- Beatgrid + key + energy + section detection (static or dynamic tempo), with confidence and “needs review” filters.
- Up to 8 cue suggestions per track, fully editable and exportable.
- Transition rehearsal: dual‑deck preview, beat‑synced scrubbing, AB loop per candidate.
- Set ordering via weighted graph with human‑readable reasons (key, tempo delta, energy, window overlap).
- Exports: M3U + JSON + cues CSV in v1, plus Rekordbox/Serato/Traktor formats.

## Architecture snapshot
```mermaid
graph TD
  UI[Web UI (SPA + Web Audio)] -->|gRPC-web / HTTP| API[Local API bridge]
  API --> Engine[Go engine daemon]
  Engine --> DB[(SQLite + blob store)]
  Engine --> Worker[Swift 6 analyzer]
  Worker --> Accel[Accelerate / Core ML / Metal]
  Engine --> Export[Exporters: M3U, JSON, cues, RB/SER/Traktor]
```

## Status & next steps
- Phase: planning & scaffold (as of 2026-01-29).
- See `docs/PLAN.md` for milestones, risks, and task breakdown.
- Source spec: `spec.md` (kept in-repo for traceability).

## Running (scaffold)
- Toolchains: Go 1.22, Swift 6 (see `go.mod`, `analyzer-swift/.swift-version`).
- Dev loop: `make test` runs Go + Swift stubs; `make fixturegen` writes a manifest placeholder under `testdata/audio`.
- CI: `.github/workflows/ci.yml` runs Go tests on Ubuntu and Swift tests on macOS (lint placeholder for now).

## License
This project uses the Blue Oak Model License 1.0.0 (see `LICENSE`).
