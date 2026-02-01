# CI/CD Strategy (Local-First, No GitHub Runners)

Per project direction, we **do not run CI on GitHub** (paid minutes avoided). All validation happens locally on a Mac before publishing releases.

## Local Pipelines

1) **UI build & lint**
```bash
cd web
npm ci
npm run build
```

2) **Go tests**
```bash
go test ./...
```

3) **Swift analyzer tests**
```bash
cd analyzer-swift
swift test
```

4) **Full signed + notarized macOS build**
```bash
bash scripts/build-and-notarize.sh
```
Outputs:
- `build/export/Algiers.app`
- `build/Algiers-v1.6-beta-AppleSilicon.dmg` (signed, notarized, stapled)

## Publishing

- Releases are created manually via `gh release create ...` after the local build passes.
- Gatekeeper acceptance is verified locally (`spctl --assess --verbose=4 build/export/Algiers.app`).

## Rationale

- Avoid GitHub Actions costs.
- Keep secrets (Developer ID, notary creds) local-only.
- Fast feedback on Apple Silicon hardware that matches our target users.

## Optional Local Automation

- You can wrap the above commands in a local `make ci` target if desired; no remote runners involved.
