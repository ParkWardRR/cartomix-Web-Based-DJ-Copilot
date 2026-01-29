# gRPC Migration Guide

This document provides guidance for migrating from the HTTP REST API to the gRPC API.

## Why gRPC?

The Algiers engine is transitioning from HTTP REST to gRPC for several benefits:

| Feature | HTTP REST | gRPC |
|---------|-----------|------|
| **Protocol** | HTTP/1.1 + JSON | HTTP/2 + Protocol Buffers |
| **Streaming** | Limited (SSE) | Full bidirectional streaming |
| **Performance** | ~100ms latency | ~10ms latency |
| **Type Safety** | Manual validation | Generated type-safe clients |
| **Code Generation** | Manual | Automatic for 10+ languages |
| **Binary Size** | Large JSON payloads | Compact binary format |

## Deprecation Timeline

| Date | Status |
|------|--------|
| **2026-01-29** | HTTP API marked deprecated (headers added) |
| **2026-04-01** | gRPC becomes primary API, HTTP receives no new features |
| **2026-07-01** | HTTP API sunset, removed from codebase |

## Connection Setup

### gRPC (Recommended)

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    engine "github.com/cartomix/cancun/gen/go/engine"
)

// Connect to gRPC server
conn, err := grpc.NewClient(
    "localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := engine.NewEngineAPIClient(conn)
```

### HTTP (Deprecated)

```javascript
// Old HTTP approach - DEPRECATED
const response = await fetch('http://localhost:8080/api/tracks');
const tracks = await response.json();
```

## API Mapping

### Track Operations

| HTTP Endpoint | gRPC Method |
|--------------|-------------|
| `GET /api/tracks` | `ListTracks` (streaming) |
| `GET /api/tracks/{id}` | `GetTrack` |
| `POST /api/scan` | `ScanLibrary` (streaming) |
| `POST /api/analyze` | `AnalyzeTracks` (streaming) |

### Set Planning

| HTTP Endpoint | gRPC Method |
|--------------|-------------|
| `POST /api/set/propose` | `ProposeSet` |
| `POST /api/export` | `ExportSet` |

### ML & Similarity

| HTTP Endpoint | gRPC Method |
|--------------|-------------|
| `GET /api/tracks/{id}/similar` | `GetSimilarTracks` |
| `GET /api/ml/settings` | `GetMLSettings` |
| `PUT /api/ml/settings` | `UpdateMLSettings` |

### Training (Layer 3)

| HTTP Endpoint | gRPC Method |
|--------------|-------------|
| `GET /api/training/labels` | `ListTrainingLabels` |
| `POST /api/training/labels` | `AddTrainingLabel` |
| `DELETE /api/training/labels/{id}` | `DeleteTrainingLabel` |
| `GET /api/training/labels/stats` | `GetTrainingLabelStats` |
| `POST /api/training/start` | `StartTraining` |
| `GET /api/training/jobs` | `ListTrainingJobs` |
| `GET /api/training/jobs/{id}` | `GetTrainingJob` |
| N/A | `StreamTrainingProgress` (streaming) |
| `GET /api/training/models` | `ListModelVersions` |
| `POST /api/training/models/{v}/activate` | `ActivateModelVersion` |
| `DELETE /api/training/models/{v}` | `DeleteModelVersion` |

### Health Check

| HTTP Endpoint | gRPC Method |
|--------------|-------------|
| `GET /api/health` | `HealthCheck` |

## Code Examples

### List Tracks (Go)

**HTTP (Deprecated):**
```go
resp, _ := http.Get("http://localhost:8080/api/tracks")
var tracks []Track
json.NewDecoder(resp.Body).Decode(&tracks)
```

**gRPC (Recommended):**
```go
stream, err := client.ListTracks(ctx, &engine.ListTracksRequest{
    Query: "artist:deadmau5",
    Limit: 50,
})
for {
    track, err := stream.Recv()
    if err == io.EOF {
        break
    }
    fmt.Printf("Track: %s - %s\n", track.Artist, track.Title)
}
```

### Similar Tracks (Go)

**HTTP (Deprecated):**
```go
resp, _ := http.Get("http://localhost:8080/api/tracks/123/similar")
var similar []SimilarTrack
json.NewDecoder(resp.Body).Decode(&similar)
```

**gRPC (Recommended):**
```go
resp, err := client.GetSimilarTracks(ctx, &engine.SimilarTracksRequest{
    TrackId: &common.TrackId{ContentHash: "abc123"},
    Limit:   10,
    MinScore: 0.7,
    Constraints: &engine.SimilarityConstraints{
        MaxBpmDelta:    5,
        SameKeyOnly:    false,
        MaxEnergyDelta: 2,
    },
})
for _, track := range resp.Similar {
    fmt.Printf("%.0f%% match: %s (%s)\n",
        track.Score*100, track.Title, track.Explanation)
}
```

### Training Progress Streaming (Go)

**HTTP (Not available):**
Training progress streaming was not available in the HTTP API.

**gRPC (New feature):**
```go
stream, err := client.StreamTrainingProgress(ctx, &engine.GetJobRequest{
    JobId: "job-123",
})
for {
    update, err := stream.Recv()
    if err == io.EOF {
        break
    }
    fmt.Printf("Epoch %d/%d, Loss: %.4f, Progress: %.1f%%\n",
        update.CurrentEpoch, update.TotalEpochs,
        update.CurrentLoss, update.Progress*100)
}
```

## Enhanced Streaming Progress (v1.0)

All streaming methods now include detailed progress information for better UX during long operations:

### ScanProgress Fields

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Full path being scanned |
| `current_file` | string | Filename only (no path) |
| `status` | string | queued, processing, done, skipped, error |
| `percent` | float | Overall progress 0-100 |
| `elapsed_ms` | int64 | Elapsed time in milliseconds |
| `eta_ms` | int64 | Estimated time remaining |
| `processed` | int64 | Files processed so far |
| `total` | int64 | Total files to process |
| `new_tracks_found` | int64 | Count of new tracks discovered |
| `skipped_cached` | int64 | Tracks skipped (already in DB) |
| `bytes_processed` | int64 | Total bytes processed |
| `bytes_total` | int64 | Total bytes to process |

### AnalyzeProgress Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | TrackId | Track identifier |
| `stage` | string | Current analysis stage |
| `status` | string | processing, done, error |
| `percent` | float | Stage progress 0-100 |
| `current_file` | string | Filename being analyzed |
| `track_index` | int32 | Current track number (1-based) |
| `total_tracks` | int32 | Total tracks in job |
| `overall_percent` | float | Overall job progress 0-100 |
| `elapsed_ms` | int64 | Elapsed time in milliseconds |
| `eta_ms` | int64 | Estimated time remaining |
| `stage_timings` | StageTiming[] | Per-stage timing breakdown |
| `stage_message` | string | Human-readable stage status |
| `title` | string | Track title (if available) |
| `artist` | string | Track artist (if available) |
| `duration_seconds` | float | Track duration |

### TrainingProgressUpdate Fields

| Field | Type | Description |
|-------|------|-------------|
| `job_id` | string | Training job identifier |
| `status` | string | queued, training, evaluating, completed, failed |
| `progress` | float | Overall progress 0-1 |
| `current_epoch` | int32 | Current training epoch |
| `total_epochs` | int32 | Total epochs configured |
| `current_loss` | float | Training loss value |
| `current_accuracy` | float | Training accuracy value |
| `message` | string | Human-readable status message |
| `elapsed_ms` | int64 | Elapsed time in milliseconds |
| `eta_ms` | int64 | Estimated time remaining |
| `validation_loss` | float | Validation set loss |
| `validation_accuracy` | float | Validation set accuracy |
| `samples_processed` | int32 | Training samples processed |
| `total_samples` | int32 | Total training samples |
| `current_stage` | string | Current pipeline stage |
| `stage_timings` | StageTiming[] | Per-stage timing breakdown |

### StageTiming Message

```protobuf
message StageTiming {
    string stage = 1;      // Stage name (e.g., "decode", "beatgrid", "key")
    int64 duration_ms = 2; // Time spent in this stage
    bool completed = 3;    // Whether stage is complete
}
```

### Example: Displaying Analysis Progress

```go
stream, _ := client.AnalyzeTracks(ctx, &engine.AnalyzeRequest{...})
for {
    progress, err := stream.Recv()
    if err == io.EOF {
        break
    }

    // Display overall job progress
    fmt.Printf("[%d/%d] %.1f%% - %s\n",
        progress.TrackIndex, progress.TotalTracks,
        progress.OverallPercent, progress.CurrentFile)

    // Display current stage
    fmt.Printf("  Stage: %s (%.0f%%) - %s\n",
        progress.Stage, progress.Percent, progress.StageMessage)

    // Display ETA
    if progress.EtaMs > 0 {
        eta := time.Duration(progress.EtaMs) * time.Millisecond
        fmt.Printf("  ETA: %v\n", eta.Round(time.Second))
    }

    // Display stage breakdown when track completes
    if progress.Status == "done" {
        fmt.Println("  Stage timings:")
        for _, st := range progress.StageTimings {
            fmt.Printf("    %s: %dms\n", st.Stage, st.DurationMs)
        }
    }
}
```

## Client Generation

### Go
```bash
buf generate
```

### TypeScript (gRPC-Web)
```bash
buf generate --template buf.gen.web.yaml
```

### Python
```bash
buf generate --template buf.gen.python.yaml
```

## Testing with grpcurl

```bash
# List all services
grpcurl -plaintext localhost:50051 list

# List methods in EngineAPI
grpcurl -plaintext localhost:50051 list cartomix.engine.EngineAPI

# Health check
grpcurl -plaintext localhost:50051 cartomix.engine.EngineAPI/HealthCheck

# Get similar tracks
grpcurl -plaintext -d '{
  "track_id": {"content_hash": "abc123"},
  "limit": 5,
  "min_score": 0.7
}' localhost:50051 cartomix.engine.EngineAPI/GetSimilarTracks

# List tracks (streaming)
grpcurl -plaintext -d '{"query": "", "limit": 10}' \
  localhost:50051 cartomix.engine.EngineAPI/ListTracks
```

## Performance Comparison

Benchmarks on 1000 track library (M3 MacBook Pro):

| Operation | HTTP | gRPC | Improvement |
|-----------|------|------|-------------|
| List 100 tracks | 45ms | 8ms | 5.6x faster |
| Get similar tracks | 120ms | 25ms | 4.8x faster |
| Export set (10 tracks) | 200ms | 40ms | 5x faster |
| Training progress stream | N/A | Real-time | New feature |

## Troubleshooting

### Connection Refused
```
Error: connection refused
```
Ensure the gRPC server is running on the correct port (default: 50051).

### TLS Certificate Errors
For production, use proper TLS certificates:
```go
creds, _ := credentials.NewClientTLSFromFile("ca.crt", "")
conn, _ := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
```

### Deadline Exceeded
For long-running operations, increase the context timeout:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
```

## Support

- GitHub Issues: https://github.com/cartomix/cancun/issues
- gRPC Documentation: https://grpc.io/docs/
- Protocol Buffers: https://protobuf.dev/
