# Algiers API Reference

## Swift Analyzer API

The Swift analyzer provides both CLI and HTTP API interfaces.

### CLI Commands

```bash
# Analyze a single track
analyzer-swift analyze <path> [options]

Options:
  -f, --format <format>    Output format: json, summary (default: json)
  -o, --output <file>      Output file (default: stdout)
  -p, --progress           Show progress to stderr

# Start HTTP server
analyzer-swift serve [options]

Options:
  -p, --port <port>        Port to listen on (default: 9090)
  --proto <protocol>       Protocol: http, grpc (default: http)

# Health check
analyzer-swift healthcheck
```

### HTTP Endpoints

#### Health Check

```http
GET /health
GET /healthz
```

Response:
```json
{
  "status": "ok",
  "version": "0.1.0"
}
```

#### Analyze Track

```http
POST /analyze
Content-Type: application/json

{
  "path": "/path/to/audio/file.flac"
}
```

Response:
```json
{
  "path": "/path/to/audio/file.flac",
  "duration": 369.61,
  "bpm": 117.19,
  "beatgridConfidence": 0.99,
  "key": {
    "name": "F#",
    "camelot": "2B",
    "openKey": "7d",
    "confidence": 0.92
  },
  "energy": 7,
  "loudness": {
    "integratedLUFS": -14.0,
    "loudnessRange": 3.7,
    "shortTermMax": -11.2,
    "momentaryMax": -9.4,
    "truePeak": 0.45,
    "samplePeak": 0.45
  },
  "sections": [...],
  "cues": [...],
  "waveformSummary": [...],
  "embedding": {
    "vector": [...],
    "spectralCentroid": 628.0,
    "spectralRolloff": 955.7,
    "zeroCrossingRate": 0.049,
    "spectralFlatness": 0.013,
    "tempoStability": 0.0,
    "harmonicRatio": 0.52
  }
}
```

---

## Go Engine API

The Go engine provides gRPC and HTTP REST APIs.

### HTTP REST Endpoints

Base URL: `http://localhost:8080/api`

#### Library

```http
GET /api/library
```

Returns all tracks in the library.

```http
POST /api/library/scan
Content-Type: application/json

{
  "path": "/path/to/music/folder"
}
```

Scans a folder for audio files and adds them to the library.

#### Tracks

```http
GET /api/tracks/{id}
```

Returns a single track by ID.

```http
GET /api/audio?path=/path/to/file.flac
```

Streams audio file for Web Audio playback.

#### Set Planning

```http
POST /api/set/plan
Content-Type: application/json

{
  "trackIds": ["id1", "id2", "id3"],
  "constraints": {
    "maxBpmDelta": 8,
    "preferCamelotCompatible": true
  }
}
```

Plans optimal track ordering with transition scoring.

#### Exports

```http
POST /api/export/rekordbox
Content-Type: application/json

{
  "trackIds": ["id1", "id2", "id3"],
  "playlistName": "My Set"
}
```

Exports to Rekordbox XML format.

```http
POST /api/export/serato
```

Exports to Serato crate format.

```http
POST /api/export/traktor
```

Exports to Traktor NML format.

```http
POST /api/export/m3u
```

Exports to M3U8 playlist.

---

## Data Types

### Track

```typescript
interface Track {
  id: string;
  path: string;
  title: string;
  artist: string;
  duration: number;      // seconds
  bpm: number;
  key: string;           // Camelot notation (e.g., "2B")
  energy: number;        // 1-10 scale
  waveformSummary: number[];
  sections: Section[];
  cues: Cue[];
  transitionWindows: TransitionWindow[];
}
```

### Section

```typescript
interface Section {
  type: 'intro' | 'verse' | 'build' | 'drop' | 'breakdown' | 'outro';
  startBeat: number;
  endBeat: number;
  startTime: number;     // seconds
  endTime: number;       // seconds
  confidence: number;    // 0-1
}
```

### Cue Point

```typescript
interface Cue {
  type: string;          // CUE_LOAD, CUE_DROP, CUE_BREAKDOWN, etc.
  beatIndex: number;
  time: number;          // seconds
  label: string;
  color: number;         // RGB hex color
}
```

### Loudness (EBU R128)

```typescript
interface Loudness {
  integratedLUFS: number;  // Integrated loudness in LUFS
  loudnessRange: number;   // Loudness range in LU
  shortTermMax: number;    // Maximum short-term loudness (3s)
  momentaryMax: number;    // Maximum momentary loudness (400ms)
  truePeak: number;        // True peak in dBTP
  samplePeak: number;      // Sample peak in dBFS
}
```

### Audio Embedding

```typescript
interface Embedding {
  vector: number[];        // 128-dimensional embedding
  spectralCentroid: number; // Brightness (Hz)
  spectralRolloff: number;  // Rolloff frequency (Hz)
  zeroCrossingRate: number; // Percussiveness (0-1)
  spectralFlatness: number; // Noise vs tonal (0-1)
  tempoStability: number;   // Tempo consistency (0-1)
  harmonicRatio: number;    // Harmonic content (0-1)
}
```

### Similarity Functions

The embedding can be used to find similar tracks:

```typescript
// Cosine similarity between two embeddings (0-1)
function similarity(a: Embedding, b: Embedding): number;

// Weighted "vibe" similarity for DJ mixing (0-1)
// Considers: vector similarity, brightness, flatness, harmonics, percussiveness
function vibeSimilarity(a: Embedding, b: Embedding): number;
```

---

## Error Handling

All endpoints return standard HTTP status codes:

| Code | Description |
|------|-------------|
| 200  | Success |
| 400  | Bad request (invalid parameters) |
| 404  | Resource not found |
| 500  | Internal server error |

Error responses include a JSON body:

```json
{
  "error": "description of the error"
}
```
