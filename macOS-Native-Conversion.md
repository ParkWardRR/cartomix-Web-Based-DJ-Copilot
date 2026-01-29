# macOS-Native Conversion Plan

[![SwiftUI-first](https://img.shields.io/badge/SwiftUI-first-F05138?style=for-the-badge&logo=swift&logoColor=white)](#target-architecture)
[![XPC-multiprocess](https://img.shields.io/badge/XPC-multiprocess-222222?style=for-the-badge&logo=apple&logoColor=white)](#why-xpc-over-grpc)
[![Structured-Concurrency](https://img.shields.io/badge/TaskGroups-actors-0A84FF?style=for-the-badge&logo=swift&logoColor=white)](#why-swift-structured-concurrency)
[![SQLite-WAL](https://img.shields.io/badge/SQLite-WAL-003B57?style=for-the-badge&logo=sqlite&logoColor=white)](#why-grdb-over-swiftdata)
[![Core-ML](https://img.shields.io/badge/Core%20ML-34C759?style=for-the-badge&logo=apple&logoColor=white)](#ml-pipeline-evidence)
[![Metal-Accelerate](https://img.shields.io/badge/Metal%20%7C%20Accelerate-147EFB?style=for-the-badge&logo=apple&logoColor=white)](#why-keep-accelerate-vdsp)
[![OpenL3-embeddings](https://img.shields.io/badge/OpenL3-512d_embeddings-FF6F61?style=for-the-badge)](#ml-pipeline-evidence)
[![Offline-first](https://img.shields.io/badge/Offline-first-00C853?style=for-the-badge&logo=lock&logoColor=white)](#impact)
[![Export-safe](https://img.shields.io/badge/Export-safe-8E43E7?style=for-the-badge)](#export-format-evidence)

## Overview
We are migrating **Damascus** (DJ Set Prep Copilot) from a Go/React-centric stack to a **bleeding-edge macOS-native app**. The new experience is SwiftUI-first, isolates heavy DSP/ML in XPC helper processes, and relies on Swift structured concurrency for predictable cancellation, throughput, and crash containment.

This document captures scope, impacts, architecture, sequencing, and acceptance bars for the conversion **with evidence-based justifications** from the existing codebase.

---

## Impact (executive in 6 bullets)
- Faster perceived latency: UI runs in-process; analysis is pipelined via XPC with GPU/ANE acceleration, cutting "first waveform" time by >30% target.
- Stability & safety: multi-process design contains crashes/overloads to helper services; the UI stays responsive and stateful.
- Offline-private by default: all compute local; no cloud calls introduced in the migration.
- Better UX fit for macOS: SwiftUI + AppKit interop unlocks native menus, keyboard focus rules, drag/drop, and media shortcuts.
- Deterministic exports: versioned analysis store + immutable job logs make Rekordbox/Serato exports auditable.
- Future-proof stack: leans into Apple's forward guidance (Swift 6, structured concurrency, XPC, SwiftData/SQLite migrations).

## Quick Goals & Non-Goals
- **Goals:** Native UI shell; XPC analyzer pipeline; versioned analysis store; deterministic exports; preserve existing protobuf contracts; retain OpenL3 ML embeddings + similarity search.
- **Non-Goals (v1):** Windows/Linux parity; Catalyst/iPad build; UI theme overhaul; cloud sync.

---

## Architectural Decisions with Evidence

### Why XPC over gRPC?

**Current State:** The `analyzer-swift` server uses raw HTTP sockets (not even full gRPC):

```swift
// analyzer-swift/Sources/AnalyzerServer/main.swift
struct Serve: AsyncParsableCommand {
    func run() async throws {
        if proto == .http {
            try await startHTTPServer(port: port, logger: logger)
        } else {
            logger.error("gRPC server not yet implemented - use --proto http")
            throw AnalyzerCLIError.notImplemented("gRPC server")
        }
    }

    private func startHTTPServer(port: Int, logger: Logger) async throws {
        let serverSocket = try createServerSocket(port: port)
        while true {
            let clientSocket = accept(serverSocket, nil, nil)
            Task {
                await handleHTTPRequest(socket: clientSocket, logger: logger)
            }
        }
    }
}
```

**Problem:** Raw socket handling, no streaming, no crash isolation, single-threaded request dispatch.

**XPC Advantages:**
1. **Crash containment:** XPC helper crashes don't bring down the UI
2. **Launchd integration:** Automatic restart with exponential backoff
3. **Sandbox compatibility:** Clean file handle passing for audio paths
4. **Lower latency:** No TCP overhead for local IPC
5. **Progress streaming:** Native support for async sequences

---

### Why Swift Structured Concurrency?

**Current State:** Analysis modules already use `@unchecked Sendable` and async patterns:

```swift
// analyzer-swift/Sources/AnalyzerSwift/Analyzer.swift
public final class Analyzer: @unchecked Sendable {
    public func analyze(
        path: String,
        progress: AnalysisProgressHandler? = nil
    ) async throws -> TrackAnalysisResult {
        // 1. Decode
        progress?(.decoding)
        let audio = try decoder.decode(path: path)

        // 2. Beatgrid (sequential - no parallelism today)
        progress?(.beatgrid(progress: 0))
        let beatgridResult = beatgridDetector.detect(samples)

        // 3-10. Other stages run sequentially...
    }
}

// Progress callback type
public typealias AnalysisProgressHandler = @Sendable (AnalysisProgress) -> Void
```

**Evidence for TaskGroups:** Current pipeline is sequential but stages 3-5 (key, energy, loudness) are independent. With TaskGroups:

```swift
// Proposed parallel execution
await withTaskGroup(of: AnalysisStageResult.self) { group in
    group.addTask { await keyDetector.detect(samples) }
    group.addTask { await energyAnalyzer.analyze(samples) }
    group.addTask { await loudnessAnalyzer.analyze(samples) }
    // Collect results...
}
```

**Cancellation propagation** is automatic - cancelling the parent Task cancels all children.

---

### Why Keep Accelerate/vDSP?

**Evidence:** FFT is already highly optimized with vDSP:

```swift
// analyzer-swift/Sources/AnalyzerSwift/DSP/FFT.swift
public final class FFTProcessor: @unchecked Sendable {
    private let fftSetup: FFTSetup

    public init(fftSize: Int = 2048) {
        self.fftSetup = vDSP_create_fftsetup(log2n, FFTRadix(kFFTRadix2))!
        vDSP_hann_window(&window, vDSP_Length(fftSize), Int32(vDSP_HANN_NORM))
    }

    public func magnitudeSpectrum(_ samples: [Float]) -> [Float] {
        // Window application
        vDSP_vmul(samples, 1, window, 1, &windowed, 1, vDSP_Length(fftSize))

        // FFT via split complex representation
        vDSP_fft_zrip(fftSetup, &splitComplex, 1, log2n, FFTDirection(FFT_FORWARD))

        // Magnitude + dB conversion
        vDSP_zvmags(&splitComplex, 1, magPtr.baseAddress!, 1, vDSP_Length(halfN))
        vDSP_vdbcon(magPtr.baseAddress!, 1, &one, magPtr.baseAddress!, 1,
                    vDSP_Length(halfN), 1)
        return magnitudes
    }

    deinit {
        vDSP_destroy_fftsetup(fftSetup)
    }
}
```

**Key patterns to preserve:**
- `vDSP_create_fftsetup` / `vDSP_destroy_fftsetup` lifecycle
- `vDSP_vmul` for window application
- `vDSP_dotpr` for dot products (used in similarity)
- `withUnsafeMutableBufferPointer` for memory safety

**Do NOT rewrite** - this code is already production-quality.

---

### Why GRDB over SwiftData?

**Evidence:** Current Go schema uses advanced SQLite features:

```sql
-- internal/storage/migrations/001_initial.sql
-- WAL mode enabled in connection string
db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")

-- Upsert with ON CONFLICT (not available in SwiftData)
INSERT INTO analyses (...) VALUES (...)
ON CONFLICT(track_id, version) DO UPDATE SET
    status = excluded.status,
    bpm = excluded.bpm,
    -- ... 20+ columns
    updated_at = CURRENT_TIMESTAMP
```

**SwiftData limitations:**
- No raw SQL for complex queries
- No `ON CONFLICT` upserts
- No explicit WAL control
- Schema migrations are opaque

**GRDB advantages:**
- Drop-in replacement for existing SQL
- Explicit migration control (matches Go pattern)
- Full WAL configuration
- Can coexist with SwiftData if needed later

---

### ML Pipeline Evidence

**Current Core ML integration is production-ready:**

```swift
// analyzer-swift/Sources/AnalyzerSwift/ML/OpenL3Embedder.swift
public final class OpenL3Embedder: @unchecked Sendable {
    private let model: MLModel?

    public init() {
        let config = MLModelConfiguration()
        config.computeUnits = .all  // Enable ANE when available

        if let modelURL = Bundle.module.url(forResource: "OpenL3Music",
                                           withExtension: "mlpackage") {
            let compiledURL = try MLModel.compileModel(at: modelURL)
            self.model = try MLModel(contentsOf: compiledURL, configuration: config)
        }
    }

    private func runInference(_ melSpec: [[Float]], model: MLModel) -> [Float] {
        // Create MLMultiArray [1, 128, 199, 1]
        let inputArray = try MLMultiArray(shape: [1, 128, 199, 1], dataType: .float32)

        // Fill: melSpec is [199 frames][128 bands]
        for frameIdx in 0..<timeFrames {
            for bandIdx in 0..<melBands {
                inputArray[[0, bandIdx, frameIdx, 0] as [NSNumber]] = value
            }
        }

        // Prediction
        let input = try MLDictionaryFeatureProvider(dictionary: ["melspectrogram": inputArray])
        let output = try model.prediction(from: input)

        // Extract 512-dim output
        if let embeddingArray = output.featureValue(for: "var_227")?.multiArrayValue {
            // ... extract embedding
        }
    }
}
```

**Key patterns:**
- `computeUnits = .all` enables ANE (Neural Engine) automatically
- Bundle resource loading for `.mlpackage`
- MLMultiArray population pattern
- Fallback when model unavailable

---

### Similarity Scoring Evidence

**Go implementation to port:**

```go
// internal/similarity/similarity.go
const (
    WeightOpenL3  = 0.50 // OpenL3 embedding cosine similarity
    WeightTempo   = 0.20 // BPM compatibility
    WeightKey     = 0.20 // Key compatibility
    WeightEnergy  = 0.10 // Energy level similarity
)

func computeCosineSimilarity(a, b []float32) float64 {
    var dotProduct, normA, normB float64
    for i := 0; i < len(a); i++ {
        dotProduct += float64(a[i]) * float64(b[i])
        normA += float64(a[i]) * float64(a[i])
        normB += float64(b[i]) * float64(b[i])
    }
    // Normalize from [-1, 1] to [0, 1]
    sim := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
    return (sim + 1) / 2
}

func computeTempoSimilarity(bpmA, bpmB float64) float64 {
    diff := math.Abs(bpmA - bpmB)
    // Support half/double tempo (64 BPM ≈ 128 BPM)
    halfDiff := math.Min(math.Abs(bpmA - bpmB*2), math.Abs(bpmA*2 - bpmB))
    diff = math.Min(diff, halfDiff)

    if diff <= 1 { return 1.0 }
    if diff >= 10 { return 0.0 }
    return 1.0 - (diff / 10.0)
}

func computeKeySimilarity(keyA, keyB string) (float64, string) {
    // Camelot wheel: same key = 1.0, relative = 0.9, ±1 = 0.85, clash = 0.2
    if keyA == keyB { return 1.0, "same" }
    numA, modeA := parseCamelot(keyA)
    numB, modeB := parseCamelot(keyB)
    if numA == numB && modeA != modeB { return 0.9, "relative" }
    // ... more rules
}
```

**Swift port will use vDSP for cosine similarity (already exists in OpenL3Embedder):**

```swift
// Already in OpenL3Embedding struct
public func similarity(to other: OpenL3Embedding) -> Float {
    vDSP_dotpr(vector, 1, other.vector, 1, &dotProduct, vDSP_Length(vector.count))
    vDSP_svesq(vector, 1, &normA, vDSP_Length(vector.count))
    vDSP_svesq(other.vector, 1, &normB, vDSP_Length(other.vector.count))
    return dotProduct / (sqrt(normA) * sqrt(normB))
}
```

---

### Export Format Evidence

**Rekordbox XML format:**

```go
// internal/exporter/rekordbox.go
type RekordboxPositionMark struct {
    Name  string `xml:"Name,attr"`
    Type  int    `xml:"Type,attr"`    // 0=cue, 1=fade-in, 2=fade-out, 4=loop
    Start string `xml:"Start,attr"`   // Seconds with 3 decimal places
    Num   int    `xml:"Num,attr"`
    Red   int    `xml:"Red,attr"`
    Green int    `xml:"Green,attr"`
    Blue  int    `xml:"Blue,attr"`
}

// Color mapping per cue type
func cueTypeToColor(cueType string) [3]int {
    colors := map[string][3]int{
        "CUE_TYPE_INTRO":    {40, 226, 20},   // Green
        "CUE_TYPE_DROP":     {230, 20, 20},   // Red
        "CUE_TYPE_BREAK":    {20, 130, 230},  // Blue
        "CUE_TYPE_BUILD":    {230, 150, 20},  // Orange
        "CUE_TYPE_OUTRO":    {200, 20, 200},  // Purple
    }
}
```

**Serato binary format:**

```go
// internal/exporter/serato.go
func EncodeSeratoMarkers(tracks []TrackExport) map[string][]byte {
    buf.WriteByte(0x02)              // Version
    buf.WriteByte(byte(len(cues)))   // Cue count (max 8)

    for i, cue := range cues {
        buf.WriteByte(byte(i))       // Cue index
        // Position in milliseconds (big endian)
        posMs := uint32(cue.GetTime().AsDuration().Milliseconds())
        binary.Write(&buf, binary.BigEndian, posMs)
        // RGB color
        buf.Write(color[:])
    }
}

// UTF-16BE path encoding
func encodeUTF16BE(s string) []byte {
    u16 := utf16.Encode([]rune(s))
    buf := make([]byte, len(u16)*2)
    for i, r := range u16 {
        buf[i*2] = byte(r >> 8)      // High byte first (big endian)
        buf[i*2+1] = byte(r)
    }
    return buf
}
```

**Traktor NML format:**

```go
// internal/exporter/traktor.go
type TraktorCueV2 struct {
    Name   string  `xml:"NAME,attr"`
    Displ  int     `xml:"DISPL_ORDER,attr"`
    Type   int     `xml:"TYPE,attr"`       // 0=cue, 1=fade-in, 2=fade-out, 4=loop
    Start  float64 `xml:"START,attr"`      // Milliseconds (not seconds!)
    Hotcue int     `xml:"HOTCUE,attr"`
}

// Path encoding: forward slashes become "/:"-delimited
locationKey := "/:file://localhost" + strings.ReplaceAll(absPath, "/", "/:")

// Key mapping to Traktor's 0-23 system
func camelotToTraktorKey(camelot string) int {
    camelotMap := map[string]int{
        "8A": 21,  // Am
        "8B": 0,   // C (root)
        "12A": 13, // C#m
    }
}
```

**Golden test pattern:**

```go
// internal/exporter/golden_test.go
func TestRekordboxGolden(t *testing.T) {
    actual, _ := WriteRekordbox(dir, "golden-set", tracks)

    if *updateGolden {
        os.WriteFile(goldenPath, actual, 0644)
        return
    }

    expected, _ := os.ReadFile(goldenPath)
    compareXML(t, "Rekordbox", expected, actual)
}
```

---

## Target Architecture

```mermaid
flowchart LR
  subgraph UI[Damascus.app (SwiftUI)]
    V(Views):::ui
    Domain[Domain layer\n(actors + reducers)]:::logic
    ExportUI[Export & set builder]:::ui
    SimilarityUI[Similarity search]:::ui
  end

  Domain -->|Async intents| Jobs[Analysis Orchestrator\n(TaskGroup owner)]:::logic
  Jobs -->|XPC| XPCSvc[XPC: AnalyzerXPC.xpc]:::xpc
  XPCSvc --> DSP[Accelerate/Metal\nCore ML (OpenL3)]:::dsp
  Jobs --> Store[(SQLite / GRDB\nversioned analysis)]:::db
  Store <--> ExportUI
  Store <--> SimilarityUI

  classDef ui fill:#0b1b2b,stroke:#0b1b2b,color:#e6f1ff;
  classDef logic fill:#12355b,stroke:#0b1b2b,color:#e6f1ff;
  classDef xpc fill:#0f4c3a,stroke:#0b1b2b,color:#e6f1ff;
  classDef dsp fill:#174a7a,stroke:#0b1b2b,color:#e6f1ff;
  classDef db fill:#003b57,stroke:#0b1b2b,color:#e6f1ff;
```

### Analysis Dataflow

```mermaid
graph TD
  Ingest[Ingest queue] --> Decode[Decode (AVFoundation)]
  Decode --> FFT[FFT/STFT (Accelerate/vDSP)]
  FFT --> Features[Features: onset, chroma, MFCC]
  Features --> ML[ML Inference (Core ML on ANE/GPU)]
  ML --> Sections[Section labeling]
  FFT --> Beat[Tempo + Beatgrid]
  Decode --> OpenL3[OpenL3 Embeddings (512d)]
  OpenL3 --> Similarity[Similarity Index]
  Sections --> Cues[Cue candidates]
  Beat --> Cues
  Decode --> Loudness[Loudness (EBU R128)]
  Features --> Key[Key Detection (Camelot)]
  Features --> Energy[Energy Curve (1-10)]
  Cues --> Store[(SQLite/GRDB)]
  Beat --> Store
  Sections --> Store
  OpenL3 --> Store
  Key --> Store
  Energy --> Store
  Loudness --> Store
  Similarity --> Store
  Store --> Export[Exports (M3U/JSON/CSV/RB/Serato/Traktor)]
  Export --> QA[Golden fixtures + checksum]
```

---

## Current → Target Gap Analysis

| Area | Current | Target | Gap / Work |
|---|---|---|---|
| **UI** | React 19 + Vite + Zustand | SwiftUI + `@Observable` | Rebuild screens; port state management patterns |
| **Orchestration** | Go engine + raw HTTP | Swift actors + TaskGroups | Port job scheduler; add structured cancellation |
| **Compute** | `analyzer-swift` with HTTP server | XPC service | Replace socket server with XPC endpoint; keep all DSP/ML |
| **Storage** | SQLite WAL via Go | GRDB for Swift | Port schema + 2 migrations; keep upsert patterns |
| **IPC** | HTTP over localhost | XPC | Define XPC protocol wrapping same data types |
| **Similarity** | Go `similarity.go` | Swift actor | Port weighted scoring; use vDSP for cosine |
| **Exporters** | Go XML/binary writers | Swift Codable + binary | Port Rekordbox/Serato/Traktor formats exactly |

---

## Schema to Port

```sql
-- internal/storage/migrations/001_initial.sql
CREATE TABLE tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_hash TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL,
    title TEXT, artist TEXT, album TEXT,
    file_size INTEGER, file_modified_at DATETIME
);

CREATE TABLE analyses (
    id INTEGER PRIMARY KEY,
    track_id INTEGER REFERENCES tracks(id) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1,
    status TEXT DEFAULT 'pending',  -- pending|analyzing|complete|failed
    duration_seconds REAL, bpm REAL, bpm_confidence REAL,
    key_value TEXT, key_format TEXT, key_confidence REAL,
    energy_global INTEGER,
    integrated_lufs REAL, true_peak_db REAL,
    beatgrid_json TEXT, sections_json TEXT, cue_points_json TEXT,
    embedding BLOB, openl3_embedding BLOB,
    UNIQUE(track_id, version)
);

CREATE TABLE cue_edits (  -- User overrides, never deleted by re-analysis
    track_id INTEGER REFERENCES tracks(id),
    cue_index INTEGER, beat_index INTEGER, cue_type TEXT, label TEXT
);

-- internal/storage/migrations/002_openl3_embeddings.sql
CREATE TABLE embedding_similarity (
    track_a_id INTEGER, track_b_id INTEGER,
    openl3_similarity REAL, combined_score REAL,
    tempo_similarity REAL, key_similarity REAL, energy_similarity REAL,
    explanation TEXT,
    PRIMARY KEY (track_a_id, track_b_id)
);

CREATE TABLE openl3_windows (
    track_id INTEGER, analysis_version INTEGER, window_index INTEGER,
    timestamp_seconds REAL, duration_seconds REAL,
    embedding BLOB  -- 512 x float32 = 2KB per window
);
```

---

## Phased Plan

```mermaid
gantt
    dateFormat  YYYY-MM-DD
    title  macOS-Native Conversion
    section Foundation (done)
    analyzer-swift DSP/ML         :done,    f1, 2026-01-15,14d
    SQLite schema + migrations    :done,    f2, 2026-01-15,7d
    Go exporters (RB/Serato/TK)   :done,    f3, 2026-01-20,10d
    OpenL3 similarity search      :done,    f4, 2026-01-25,5d
    section Shell & State
    SwiftUI app skeleton          :active,  a1, 2026-01-29,5d
    Domain reducers/actors        :         a2, 2026-02-03,7d
    section XPC Migration
    XPC service scaffold          :         a3, 2026-02-05,5d
    Wrap existing Swift analyzers :         a4, 2026-02-10,5d
    section Data & Export
    Port schema to GRDB           :         a6, 2026-02-10,7d
    Port exporters to Swift       :         a7, 2026-02-17,8d
    Port similarity to Swift      :         a8, 2026-02-17,5d
    section UX & QA
    Library + track detail        :         a9, 2026-02-20,8d
    Set builder + rehearsal       :         a10, 2026-02-28,8d
    Golden corpus replay          :         a12, 2026-03-14,5d
```

---

## Risks & Mitigations

| Risk | Mitigation |
|---|---|
| **SwiftData immaturity** | Start with GRDB; same SQL as Go; feature-flag SwiftData |
| **XPC crash loops** | Launchd plist with exponential backoff; circuit breaker in domain actor |
| **Model drift** | Version analysis artifacts; invalidate cache on version change |
| **Perf regressions** | Benchmarks for decode→beatgrid→sections; cap TaskGroup concurrency |
| **Export fidelity** | Golden round-trips per DJ format; checksum exports; CI regression |

---

## Validation Strategy

1. **Unit:** XCTest for FFT, section classifier, cue placement (extend `analyzer-swift/Tests/`)
2. **Integration:** Golden corpus through XPC; compare beats/sections/embeddings to JSON
3. **Export fixtures:** Port `golden_test.go` pattern to Swift; verify XML/binary byte-exact
4. **Similarity:** Verify cosine similarity matches Go within epsilon; verify weighted formula
5. **UI:** Xcode UI tests for import → analyze → cues → export flows

---

## Success Criteria

- **TTI:** Track-to-insight under 5s for 5-minute track on M3 MacBook Pro
- **Responsiveness:** Main thread tasks < 10ms during 10 concurrent analyses
- **Crash resilience:** Zero data loss on analyzer crash; job resumes from checkpoint
- **Export determinism:** Round-trips deterministic; fixtures green in CI

---

## Appendix: Existing Codebase Reference

| Component | Path | Key Code |
|---|---|---|
| FFT/DSP | `analyzer-swift/.../DSP/FFT.swift` | `vDSP_fft_zrip`, `vDSP_vmul`, `vDSP_zvmags` |
| Audio Decode | `analyzer-swift/.../DSP/AudioDecoder.swift` | `AVAudioFile`, `AVAudioConverter` |
| OpenL3 ML | `analyzer-swift/.../ML/OpenL3Embedder.swift` | `MLModel`, `MLMultiArray`, `computeUnits = .all` |
| Beatgrid | `analyzer-swift/.../Analysis/BeatgridDetector.swift` | Autocorrelation tempo estimation |
| Key Detection | `analyzer-swift/.../Analysis/KeyDetector.swift` | Krumhansl-Schmuckler chroma profiles |
| Sections | `analyzer-swift/.../Analysis/SectionDetector.swift` | Energy-based classification |
| Similarity | `internal/similarity/similarity.go` | Weighted cosine (0.5 OpenL3 + 0.2 tempo + 0.2 key + 0.1 energy) |
| Rekordbox | `internal/exporter/rekordbox.go` | XML with POSITION_MARK, TEMPO elements |
| Serato | `internal/exporter/serato.go` | Binary chunks, UTF-16BE paths, big-endian ms |
| Traktor | `internal/exporter/traktor.go` | NML XML, `/:` path encoding, 0-23 key system |
| Golden Tests | `internal/exporter/golden_test.go` | `-update-golden` flag pattern |
| Schema | `internal/storage/migrations/*.sql` | WAL mode, upserts, embedding BLOBs |
