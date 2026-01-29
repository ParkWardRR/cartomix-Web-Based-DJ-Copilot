Commit and merege to main branch frequintly and often. Start with plan, diagrams etc then go ot the max right the readme first with lots of badges and commit and push to main.  
Blue Oak Model License
Version 1.0.0



Build a “DJ set prep copilot” that (1) scans your library, (2) detects *mixable* sections (intro/outro, breakdowns, drops), (3) proposes cue points + transition windows, and (4) auto-sorts tracks into an optimized set order—while keeping the human fully in control of the actual mix. [mixedinkey](https://mixedinkey.com/wiki/how-cue-points-can-improve-your-djing/)

## Product spec (what DJs need)
### Core workflows
| Workflow | User story | Must-have UX details |
|---|---|---|
| Library ingest | “Point it at a folder (or Rekordbox/Serato export) and it just analyzes everything.” | Background analysis queue, resumable jobs, per-track progress, “re-analyze with different settings.” |
| Set planning | “I pick 30–80 tracks; it suggests a playable order with reasons.” | Graph view (track nodes + edge scores), sortable list w/ explanations (“key compatible”, “energy ramp”, “tempo bridge”), manual override always wins. |
| Cue point authoring | “Mark me good cue points, but let me move them fast.” | Up to 8 suggested cues per track, editable and exportable (hotkeys, snap-to-beat, undo history).  [mixedinkey](https://mixedinkey.com/wiki/how-cue-points-can-improve-your-djing/) |
| Transition rehearsal | “Preview the transitions quickly without loading my DJ app.” | Two-deck preview player, beat-synced scrubbing, AB loop per transition window, instant “try next candidate.” |
| Export | “Send results to my DJ ecosystem.” | Export cue points + beatgrid hints + playlist order (start with M3U + JSON; later Rekordbox/Serato formats). |

### Analysis outputs (per track)
Borrow the DJ mental model used by tools like Mixed In Key: key detection, energy rating, and automatic cue points/markers. [dj](https://dj.studio/blog/mixed-in-key-integration)
Also support variable-tempo tracks using static vs dynamic grid modes (dynamic when tempo varies). [lexicondj](https://www.lexicondj.com/blog/understanding-rekordbox-beatgrid-analysis)

## Architecture (maximum performance)
### High-level system
| Component | Tech | Responsibilities |
|---|---|---|
| Desktop “engine” service | Go (primary) | File scanning, waveform decode, feature extraction, similarity graph build, job scheduling, local DB, export. |
| Apple-optimized DSP/ML module | Swift 6 (+ C/Metal where needed) | Hot path: FFT/STFT, embeddings inference, Core ML integration, Metal/MPS acceleration paths.  [developer.apple](https://developer.apple.com/videos/play/wwdc2024/10218/) |
| Web app UI | Latest web app stack (SPA) + Web Audio | Library UI, set builder, transition preview, cue editing; low-latency audio preview via Web Audio (AudioWorklet for real-time processing).  [developer.mozilla](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet) |
| Local API | gRPC over Unix socket + HTTP fallback | UI ↔ engine communication, streaming analysis results, waveform tiles, audio preview endpoints. |
| DB | SQLite (WAL) | Track metadata, analysis artifacts, graph edges, user edits, versioned analysis parameters. |

### Performance principles
- Do *not* do DSP in JavaScript; use Web Audio for playback/preview, but do analysis in Go + Swift modules, and stream down compact results (waveform peaks, beat markers, cue candidates). [developer.mozilla](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet)
- Use Apple acceleration where it matters: Core ML can use Metal Performance Shaders under the hood for best Apple Silicon performance, and Metal/MPSGraph is explicitly positioned for accelerating ML workloads. [createwithswift](https://www.createwithswift.com/core-ml-explained-apples-machine-learning-framework/)
- FFT/STFT: use Accelerate/vDSP on macOS/iOS-family platforms for highly optimized transforms. [docs.huihoo](https://docs.huihoo.com/apple/wwdc/2012/session_708__the_accelerate_framework.pdf)

## Detailed functional requirements
### 1) Ingest + decode
| Requirement | Spec |
|---|---|
| Supported formats | WAV/AIFF/MP3/AAC/ALAC/FLAC (decode via platform libs where possible; fallback decoder via ffmpeg CLI optional). |
| Loudness normalization | Compute EBU R128-style integrated loudness + true peak; store replay gain (for preview only, never rewrite original). |
| Caching | Content-hash audio fingerprints; if file unchanged, skip full re-analysis; store analysis version keyed by parameters. |

### 2) Beat/tempo + beatgrid
| Feature | Spec |
|---|---|
| BPM + downbeat | Estimate tempo, detect downbeats, output beat times array. |
| Static vs dynamic grid | Default static; enable dynamic mode for variable-tempo material (funk/disco/live) similar to “dynamic analysis” guidance.  [lexicondj](https://www.lexicondj.com/blog/understanding-rekordbox-beatgrid-analysis) |
| Confidence + QC | Provide “grid confidence” score and a quick “needs review” filter for DJs. |

### 3) Phrase/section detection (mix opportunities)
Detect and label regions:
- Intro/outro (low spectral flux + stable rhythm)
- Breakdowns (energy dips)
- Drops/peaks (energy spikes)
- “Transition windows” (8/16/32-bar aligned candidate zones)

Store: section start/end (in beats + seconds), label, confidence, recommended cue placement.

### 4) Key + harmonic compatibility
| Feature | Spec |
|---|---|
| Key detection | Compute musical key + Camelot-style mapping (for UI). |
| Compatibility scoring | Same key, relative minor/major, +/-1 Camelot step, etc. (configurable ruleset). |
| Explainability | UI must show *why* the next track is suggested (“8A → 9A, energy +1, tempo -2 BPM”). |

### 5) Energy model (track + within-track)
Implement an “energy level” concept and also detect energy changes within a track, because DJs use both for programming and skipping to higher-energy sections. [mixedinkey](https://mixedinkey.com/learn-more/)
Outputs:
- TrackEnergy: 1–10
- EnergySegments: list of (t0,t1,level)

### 6) Automatic cue points
Generate up to 8 cue points per track (configurable), aligned to beats and meaningful sections; allow user editing and export. [mixedinkey](https://mixedinkey.com/wiki/how-cue-points-can-improve-your-djing/)
Cue types: Load, IntroStart, FirstDownbeat, Breakdown, Build, Drop, OutroStart, SafetyLoop.

### 7) Set ordering (playlist optimization)
Model as a weighted directed graph where edges are “how well Track A transitions to Track B”.
- Edge score = weighted sum of: harmonic compatibility, tempo delta (penalize big jumps unless a “bridge” track), energy delta smoothness, section-window availability overlap, user constraints (must-play, ban, max BPM step).
- Solve: find best path through chosen tracks (variants: longest good path, or full ordering with minimal penalty).  
Provide modes: “Warm-up”, “Peak-time”, “Open-format (allows big jumps)”.

## Web app UX spec (DJ-first)
### Key screens
| Screen | UX requirements |
|---|---|
| Library | Fast search, key/energy chips, “needs grid review” filter, bulk actions. |
| Track detail | Waveform w/ beat markers + section overlays + cue flags; drag cues; audition from cue; snap to beat. |
| Set builder | Two panes: proposed order + “reasons”; timeline energy curve; click any transition to rehearse. |
| Transition rehearsal | Dual deck view; show recommended in/out windows; AB compare multiple candidates. |

### Audio preview tech
Use Web Audio for preview and timing-accurate scheduling (musical apps need high rhythmic precision), and use AudioWorklet when custom low-latency processing is needed. [w3](https://www.w3.org/TR/webaudio-1.1/)

## Apple Silicon / M4 optimization plan (Go + Swift 6)
### Where Swift 6 is “needed”
- A Swift package (or XCFramework) exposing:
  - Accelerate/vDSP FFT/STFT utilities for feature extraction. [developer.apple](https://developer.apple.com/documentation/accelerate/vdsp/fft)
  - Core ML model runner for embeddings/section classifiers (Core ML can leverage Metal Performance Shaders). [developer.apple](https://developer.apple.com/videos/play/wwdc2024/10218/)
  - Optional Metal/MPSGraph fast path for transformer-ish models where appropriate (Apple documents MPSGraph acceleration). [developer.apple](https://developer.apple.com/videos/play/wwdc2024/10218/)

### How Go calls into Swift
- Preferred: run Swift module as a separate local process (“analyzer-worker”) for crash isolation + easy profiling, communicate via gRPC/flatbuffers shared memory.  
- Alternate: cgo bridge to C ABI wrapper around Swift (more complex build + harder crash isolation).

## Testing (Playwright-Go + max rigor)
### Test layers
| Layer | Tooling | What to test |
|---|---|---|
| Unit | Go test, Swift XCTest | Beat detection determinism, cue placement invariants, DB migrations, scoring math. |
| Property-based | Go (rapid/gopter) | “Cue points always within track duration”, “beat times monotonic”, “export/import round-trips.” |
| Integration | Golden test corpus | Compare analysis outputs against locked reference results; tolerate epsilon drift. |
| E2E UI | playwright-go | Full flows: import → analyze → build set → edit cues → rehearse → export; run on Chromium + WebKit.  [pkg.go](https://pkg.go.dev/github.com/playwright-community/playwright-go) |

Playwright-Go install is via `go get github.com/playwright-community/playwright-go`, and it drives Chromium/Firefox/WebKit through Playwright’s automation stack. [browserstack](https://www.browserstack.com/guide/playwright-go)

## Test audio generation (for brutal validation)
### Generated fixtures (commit to repo)
| Fixture | Purpose | How to generate (spec) |
|---|---|---|
| Perfect 4/4 click + kick | Baseline BPM accuracy | Render at multiple BPMs (80–180), fixed phase, no swing. |
| Swing + shuffle hats | Groove robustness | Add swing ratios 52–65%, vary hats. |
| Tempo ramp track | Dynamic grid validation | Linear BPM change (e.g., 128→100) to force dynamic analysis. |
| Phrase structure | Cue correctness | 16-bar intro → 32-bar verse → breakdown → build → drop → outro. |
| Harmonic set | Key detection accuracy | Sine/chord progressions in known keys + controlled noise. |
| “Club mix” noise | Realism | Add sidechain compression envelope + pink noise bed. |

Implementation: generate WAV via Go (pure synthesis) + optional SoX/ffmpeg for convenience; store expected beat times and section labels as JSON golden files.

## README assets (GIFs)
Create deterministic UI GIFs from Playwright traces:
- Script Playwright-Go to run flows and record video; convert to GIF in CI for README badges/demos. (This is an implementation requirement; exact conversion tool can be `ffmpeg` in CI.)

## Open questions (to finalize scope)
What DJ ecosystem is the top export target first: Rekordbox, Serato, Traktor, or “generic (M3U + CSV/JSON)”?


You’re building a local-first “Set Prep Copilot” (Go + Swift 6 + modern web UI) that analyzes tracks to propose beatgrids, cue points, and *mixable* transition windows, then optimizes set order with explainable scoring—wrapped in an Apple-level UI (light by default, auto-switch via system preference + manual toggle). [developer.apple](https://developer.apple.com/design/human-interface-guidelines/dark-mode)

### Research anchors (so the spec matches reality)
| Topic | What we’re aligning to | Official link |
|---|---|---|
| Light/dark behavior | Apple’s guidance for Dark Mode and appearance | https://developer.apple.com/design/human-interface-guidelines/dark-mode  [developer.apple](https://developer.apple.com/design/human-interface-guidelines/dark-mode) |
| Auto theme switching | Use `prefers-color-scheme` to detect user OS/browser preference | https://mdn.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme  [mdn2.netlify](https://mdn2.netlify.app/en-us/docs/web/css/@media/prefers-color-scheme/) |
| Low-latency web audio | AudioWorklet runs audio processing on a separate thread for very low latency | https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet  [developer.mozilla](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet) |
| Web audio perf | MDN notes AudioWorklet processors can benefit from WebAssembly for near-native perf | https://developer.mozilla.org/en-US/docs/Web/API/Web_Audio_API/Using_AudioWorklet  [developer.mozilla](https://developer.mozilla.org/en-US/docs/Web/API/Web_Audio_API/Using_AudioWorklet) |
| Apple ML acceleration | Apple states MPSGraph is used under-the-hood in frameworks like Core ML on Apple Silicon | https://developer.apple.com/videos/play/wwdc2024/10218/  [developer.apple](https://developer.apple.com/videos/play/wwdc2024/10218/) |
| Metal 4 ML | Metal 4 adds “Shader ML” + GPU-timeline ML capabilities (useful for specialized kernels) | https://developer.apple.com/videos/play/wwdc2025/262/  [developer.apple](https://developer.apple.com/videos/play/wwdc2025/262/) |
| Web UI E2E tests | Playwright-Go automates Chromium/Firefox/WebKit via a single API | https://pkg.go.dev/github.com/playwright-community/playwright-go  [pkg.go](https://pkg.go.dev/github.com/playwright-community/playwright-go) |

## 1) Deliverable-grade product spec
| Area | Requirement (dev-team testable) | Acceptance criteria |
|---|---|---|
| Primary user | DJ prepping sets (house/techno/open format), wants fewer trial transitions and faster cueing | A user can import 100 tracks, select 30, and get a proposed order + transition rehearsal UI without manual tagging. |
| Core outputs | Beatgrid (static/dynamic), musical key, energy curve, section markers, cue candidates, transition windows, set ordering with reasons | Every transition in the proposed set has a “why this works” panel with computed scores and human-readable reasons. |
| Human control | No destructive edits; user edits override the model; allow “lock” per-track/per-transition | Re-running analysis never deletes user-authored cues; it creates a new analysis version and re-applies user diffs. |
| Local-first | Runs offline; files stay on disk; analysis DB local | With network disabled, import/analyze/build/rehearse/export still works. |
| Pro export | Export playlist order + cues + beat markers (v1 generic), then Rekordbox/Serato targets | v1 export produces: `set.m3u8`, `analysis.json`, and `cues.csv` with time + beat indices. |

## 2) Architecture (Go + Swift 6 + “latest web app”)
| Component | Tech choice | Why / constraints | Interfaces |
|---|---|---|---|
| Engine daemon | Go 1.22+ | High-throughput pipeline, strong concurrency, easy local service model | gRPC over Unix domain socket; HTTP fallback for browser UI if needed |
| Apple accel module | Swift 6 + Accelerate/vDSP + Core ML | Put FFT/STFT + ML inference where Apple tooling is strongest; Core ML can leverage MPSGraph under the hood on Apple Silicon.  [developer.apple](https://developer.apple.com/videos/play/wwdc2024/10218/) | Separate worker process (`analyzer-swift`) with protobuf contracts; shared memory for large tensors/waveforms |
| UI | SPA (React 19 / Next.js or Vite-based), TypeScript, Web Audio | UI must be instant; audio preview in browser via Web Audio; heavier DSP stays native | gRPC-web proxy or local HTTP bridge |
| Audio preview | Web Audio + AudioWorklet | AudioWorklet executes on a separate audio thread for very low latency processing.  [developer.mozilla](https://developer.mozilla.org/en-US/docs/Web/API/AudioWorklet) | UI sends transport state; engine streams decoded PCM chunks or pre-rendered previews |
| Optional DSP-in-browser | WASM inside AudioWorklet | MDN notes AudioWorklet processors may benefit greatly from WebAssembly performance.  [developer.mozilla](https://developer.mozilla.org/en-US/docs/Web/API/Web_Audio_API/Using_AudioWorklet) | Only for lightweight preview FX (filters, gain staging), not full analysis |
| Storage | SQLite (WAL) + blob store | WAL for concurrency; store large arrays (beat times, embeddings) as compressed blobs | Migration-managed schema with semantic versioning |

### Process model (performance + stability)
| Process | Responsibility | Crash containment | Perf notes |
|---|---|---|---|
| `app-ui` | Renders UI, calls local API, plays previews | N/A | Keep main thread < 10ms tasks; use workers for decoding UI assets |
| `engine-go` | Scans library, schedules jobs, persists results, serves APIs | If crash, restart and resume from DB | Work-stealing worker pool; CPU pinning optional |
| `analyzer-swift` | Hot path DSP/ML | If crash, engine restarts worker and retries job | Batch inference; reuse FFT plans; reuse Core ML model instances |

## 3) Analysis pipeline (DJ-grade, tech-driven)
### Data model (persisted per track)
| Artifact | Type | Notes |
|---|---|---|
| Decode profile | sampleRate, channels, duration, codec | Keep original; generate internal analysis PCM at fixed SR (e.g., 48k mono) |
| Waveform tiles | multi-resolution peak/RMS tiles | For instant zooming; precompute for UI |
| Beatgrid | `beats[]` (seconds), `downbeats[]`, `bpm(t)` model | Support static + dynamic grid; dynamic represented as piecewise linear tempo map |
| Sections | list of (startBeat, endBeat, label, confidence) | Labels: intro, verse, breakdown, build, drop, outro |
| Energy | global 1–10 + per-section curve | Store both scalar and curve for set planning |
| Key | key + confidence + Camelot mapping | Camelot mapping only for display |
| Cue candidates | list of (time, beatIndex, type, confidence) | Types: load, first-downbeat, intro-in, breakdown-in, build-in, drop, outro-in, safety-loop |
| Track embedding | vector<float16> | Used for similarity and “vibe” continuity |

### Core algorithms (implementation-level requirements)
| Stage | Implementation requirement | Optimization requirement |
|---|---|---|
| Decode | Stream decode; never load full track into RAM unless explicitly requested | Cache decoded analysis PCM chunks in temp store; memory-map where possible |
| STFT/FFT | Use Accelerate/vDSP in Swift module for FFT-heavy work | Reuse FFT setup/plans; batch frames; SIMD-friendly memory layout |
| Beat tracking | Multi-band onset envelope + tempo hypothesis scoring + downbeat model | Use vectorized ops; keep determinism (same input ⇒ same beat list) |
| Section detection | Classify bars/phrases using rhythmic + spectral features; optional ML classifier | Batch inference; prefer Core ML pipeline so it can utilize Apple acceleration.  [developer.apple](https://developer.apple.com/videos/play/wwdc2024/10218/) |
| Cue generation | Rule+model hybrid: snap to beats/downbeats; prefer boundaries of sections | Must never place cues within first/last 0.5s unless type is load/outro |
| Transition windows | For each track, identify N candidate “mix in” and “mix out” windows aligned to 8/16/32 bars | Store windows with compatibility tags (low-energy, peak, breakdown-to-build) |
| Set ordering | Weighted directed graph; solve best path under constraints | Must produce explanations per edge: key match, BPM delta, energy delta, window overlap |

### Hardware acceleration policy (Apple M4 class targets)
| Workload | Preferred engine | Fallback | Notes |
|---|---|---|---|
| FFT/STFT feature extraction | Accelerate/vDSP | Go SIMD where feasible | vDSP is the default fast path on Apple platforms |
| ML inference (sections/embeddings) | Core ML (Metal/MPSGraph/ANE selection) | CPU inference | Apple describes Metal acceleration and notes MPSGraph is used under-the-hood in frameworks like Core ML for performance on Apple Silicon.  [developer.apple](https://developer.apple.com/videos/play/wwdc2024/10218/) |
| Custom ML kernels (future) | Metal 4 Shader ML | Core ML | Metal 4 adds Shader ML and high-performance ML capabilities on the GPU timeline for smaller embedded networks.  [developer.apple](https://developer.apple.com/videos/play/wwdc2025/262/) |

## 4) UI spec: “maximum beautiful like Apple”
### Theme + appearance (light default, auto + manual override)
| Requirement | Spec | Acceptance criteria |
|---|---|---|
| Default | Light mode first-run | First-run renders light theme regardless of system; then auto mode can be enabled |
| Auto mode | Detect OS/browser preference via `prefers-color-scheme` and live-update on change | Switching OS appearance flips UI without reload when Auto is enabled.  [mdn2.netlify](https://mdn2.netlify.app/en-us/docs/web/css/@media/prefers-color-scheme/) |
| Manual toggle | UI has “Light / Dark / Auto” selector (persisted) | Persists per device; override beats auto |
| Dark mode quality | Follow Apple dark mode guidance: keep contrast readable, avoid “inverted” artifacts | All charts/waveforms remain legible and colorblind-safe in both modes.  [developer.apple](https://developer.apple.com/design/human-interface-guidelines/dark-mode) |

### Visual system (Apple-grade constraints)
| System | Spec |
|---|---|
| Layout | 8pt grid, generous whitespace, large typography hierarchy, motion restrained (200–300ms) |
| Materials | Use subtle translucency sparingly (only for overlays/modals), strong depth separation for waveforms and timelines |
| Typography | Use system UI fonts (SF on macOS via system stack), do not bundle SF files (licensing risk) |
| Icons | SF Symbols-like weight and geometry; consistent stroke widths; crisp at 1x/2x |
| Micro-interactions | “Snap to beat” while dragging cues; spring easing; haptic-style soundless feedback (visual only) |

### Core screens (pixel-level behavior, but implementable)
| Screen | Components | Interaction requirements |
|---|---|---|
| Library | Data grid + fast filters + background job HUD | Scroll is virtualized; filters update <50ms; analysis runs without blocking UI |
| Track detail | Waveform, beat markers, section overlays, cue flags, mini “mix windows” lane | Drag cue snaps to nearest beat; hold modifier for fine adjust; undo/redo always available |
| Set builder | Left: ordered list; Right: transition inspector; Top: energy arc timeline | Drag reorder recalculates only affected transitions; every transition shows “recommended in/out bars” |
| Transition rehearsal | 2-deck preview with sync + AB compare | One keypress to jump to recommended in-point; instant audition of next candidate transition |

## 5) Testing, test audio, README GIFs (pro pipeline)
### E2E web testing (Playwright-Go)
Playwright-Go must be the primary browser automation harness, and it automates Chromium, Firefox, and WebKit through one API for cross-browser coverage. [pkg.go](https://pkg.go.dev/github.com/playwright-community/playwright-go)
| Suite | What it validates | How it runs |
|---|---|---|
| Import + analyze | Job queue correctness, progress UI, resumability | Deterministic test corpus + golden outputs |
| Cue editing | Snap-to-beat, undo/redo, persistence | Playwright asserts DOM + canvas snapshots |
| Set ordering | Explanations, constraint locks, recompute locality | Seeded graph; verify stable ordering with fixed seed |
| Theme | Light default, Auto follows system preference, manual override | Simulate `prefers-color-scheme` and verify CSS variables |

### Generated test audio (repo-owned fixtures)
| Fixture | Goal | Determinism rule |
|---|---|---|
| BPM ladder clicks | Exact BPM accuracy check | Beat times must match to ±1ms |
| Swing hats | Robustness to groove | Onset detector must still find beats |
| Dynamic tempo ramp | Force dynamic grid path | Must output non-constant BPM model |
| Structured “EDM phrase” | Cue + section correctness | Must find intro/build/drop boundaries within ±1 bar |
| Harmonic chords | Key detection sanity | Must match declared key with confidence threshold |

#### One-line generator entrypoint (spec)
```bash
go run ./cmd/fixturegen --out ./testdata/audio --seed 1337 --render-wav --bpm-ladder 80,100,120,128,140,160 --include-swing --include-tempo-ramp --include-harmonic-keys "8A,9A,10A,11A"
```

### README GIFs (from Playwright videos)
| Asset | Source | Command (CI-friendly) |
|---|---|---|
| `import-analyze.gif` | Playwright video recording | ```bash
ffmpeg -y -i ./artifacts/playwright/import-analyze.webm -vf "fps=24,scale=960:-1:flags=lanczos" ./README/import-analyze.gif
``` |
| `set-builder.gif` | Playwright video recording | ```bash
ffmpeg -y -i ./artifacts/playwright/set-builder.webm -vf "fps=24,scale=960:-1:flags=lanczos" ./README/set-builder.gif
``` |

### Budget notes (explicitly a guess)
If you want this to feel “stupid good,” plan for 2–4 senior engineers for ~4–8 months plus design, which is roughly $250k–$1.2M USD depending on team/location and whether you hire contractors (guess).  

If you tell me your target DJ platform to export to first (Rekordbox vs Serato vs Traktor) and whether this is macOS-only or cross-platform, I’ll lock the file-format/export spec and the build/release pipeline details.
Export targets (v1): Rekordbox, Serato, and Traktor.

We will support the three most common DJ ecosystems out of the box—Rekordbox, Serato, and Traktor—so DJs can take the proposed set order, cue points, and analysis into the software they already perform with. This is explicitly designed to avoid “generic-only” exports that lose critical DJ prep data like hot cues/beatgrid metadata (e.g., an M3U alone won’t carry cues/beatgrids in many workflows). [web:78][web:81][web:85]
