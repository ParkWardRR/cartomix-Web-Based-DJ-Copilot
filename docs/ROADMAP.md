# Roadmap — Imported Snapshot
Source: `macos-finder-quick-actions-video-tools-and-batch-playback-scripts/tallinn-v1/roadmap.md` (captured 2026-01-29). Kept here verbatim for traceability and to reuse the hardening checklist for our local-first media analysis stack.

## Current Snapshot
- Go service `qacache` builds and unit tests pass (`go test ./...`).
- Automator Quick Actions and shell-based cache (`cache_lib.sh`) are still the default on macOS; the Go service is deployed on the AlmaLinux host but not yet integrated with the macOS workflows.
- Systemd service runs with hardened settings; Docker/Podman assets and deployment scripts exist. Logrotate config targets `/var/log/qacache/qacache.log` while the service logs to stdout/journald.

## Immediate Stabilization (0–2 weeks)
- Config/Env parity: align env vars used in `qacache.service` with code (`QACACHE_SERVER_ADDR`, `QACACHE_FFPROBE_CONCURRENCY`, `QACACHE_DB_CHECKPOINT_EVERY` currently ignored). Add explicit env overrides for busy_timeout, vacuum/backup intervals, log output, metrics path.
- Schema migrations: implement real ALTER/ADD migrations tied to `schema_version`; fail fast when DB is older than supported; add migration test.
- Scanner robustness:
  - Stream files instead of materializing the full list (current `discoverFiles` builds an in-memory slice, risky on large libraries).
  - Handle `PutBatch` errors (currently dropped) and surface in scan progress and Prometheus metrics.
  - Add backpressure when the result channel is slow; ensure graceful cancellation flushes pending batches.
- Stronger cache keys: actually compute and persist a content hash (fast chunk hash) to detect renamed/mtime-copied files; gate writes on stability check + hash option.
- Readiness/health: expand `/ready` to check NFS readability, free space at DB path, WAL checkpoint ability, and ffprobe availability. Fail if the scan root is missing or read-only.
- Security:
  - Restrict `/metrics` and scan control endpoints (IP allowlist or token/Bearer auth); tighten CORS (currently `*`).
  - Consider separate listener/port for internal ops.
- Logging/rotation: either log to `/var/log/qacache/qacache.log` to match logrotate, or update logrotate to read journald; add structured fields (scan_id, file) to scanner logs.
- Tests: add integration tests for HTTP handlers (lookup/ready/scan lifecycle) and scanner rate-limit/unstable-file paths; add config env override tests.

## Workflow Integration (2–4 weeks)
- remove local cahche option. Server cache only 
- Update Automator workflows to call the client/HTTP lookup instead of scanning with `find` + ffprobe; preserve offline fallback.
- Replace `cache_lib.sh` with a thin HTTP client shim; keep portable DB only as cache-miss fallback.
- Path translation: extend `pathtranslate` to handle multiple mount prefixes and trailing-slash differences; auto-detect active volume and cache the mapping.
- Offline/latency handling: add client-side timeouts and notification UX when the server is unreachable.

## Observability & Operations (4–6 weeks)
- Prometheus/Grafana: create dashboards for scan throughput, cache hit ratios, ffprobe latency, WAL checkpoints, and error types; expose build/version labels.
- SSE/WS health: add metrics for SSE subscriber counts and broadcast failures; expose `/healthz` for load balancers.
- Maintenance jobs: schedule WAL checkpoints and VACUUM via systemd timers; add backup retention/rotation policy with integrity checks.
- Alerting hooks: optional webhook/email on scan failure or high ffprobe error rate.

## Feature Enhancements (6–10 weeks)
- API additions: endpoints for top-N largest videos by orientation, recent failures with retry controls, and manual re-probe of a path.
- Dedup/diff tooling: surface content-hash duplicates and size deltas between scans.
- Dashboard UX: move Tailwind CDN to vendored assets, add dark/light toggle, and provide filterable tables with search/sort.
- Rate limiting & QoS: configurable per-scan rate limits and max concurrent ffprobe per mount to protect NAS.
- Export/import: allow exporting cache segments as NDJSON/CSV and importing on macOS for offline lookup.

## Packaging & Delivery (parallel)
- Podman/Docker: publish a hardened image with non-root user, read-only rootfs, configurable volume mounts; include example `podman generate systemd` unit.

## Cleanup / Backlog
- move stray transcript file `spec.sh` or move to docs.
- Document supported media extensions in one place (Go config and shell scripts differ slightly).
- deprecate cache dashboard script points to the Go service or is deprecated once the web UI suffices.
