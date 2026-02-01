# Code Signing & Notarization Guide

This guide explains how to set up Apple code signing and notarization for Algiers.

## Prerequisites

1. **Apple Developer Account** (paid, $99/year)
2. **Developer ID Application certificate** installed in Keychain
3. **App-specific password** for notarization

---

## One-Time Setup

### 1. Get Your Team ID

Your Team ID is visible in the Apple Developer portal or via:

```bash
security find-identity -v -p codesigning | grep "Developer ID Application"
```

Output looks like:
```
1) ABC123XYZ "Developer ID Application: Your Name (ABC123XYZ)"
```

The Team ID is the part in parentheses: `ABC123XYZ`

### 2. Create an App-Specific Password

1. Go to https://appleid.apple.com
2. Sign in with your Apple ID
3. Navigate to **Security** → **App-Specific Passwords**
4. Click **Generate an app-specific password**
5. Name it something like "Algiers Notarization"
6. Copy the generated password (format: `xxxx-xxxx-xxxx-xxxx`)

### 3. Store Notarization Credentials in Keychain

Run this command to securely store your credentials:

```bash
xcrun notarytool store-credentials "notary-api" \
    --apple-id "your-apple-id@example.com" \
    --team-id "YOUR_TEAM_ID" \
    --password "xxxx-xxxx-xxxx-xxxx"
```

Replace:
- `your-apple-id@example.com` with your Apple ID email
- `YOUR_TEAM_ID` with your Team ID (e.g., `6U62M4232W`)
- `xxxx-xxxx-xxxx-xxxx` with your app-specific password

This stores the credentials in your macOS Keychain under the profile name `notary-api`.

### 4. Verify Credentials Work

```bash
xcrun notarytool history --keychain-profile "notary-api"
```

If this runs without error, you're set up correctly.

---

## Build Script Configuration

The build script (`scripts/build-and-notarize.sh`) uses these variables:

```bash
TEAM_ID="6U62M4232W"                                          # Your Team ID
DEVELOPER_ID="Developer ID Application: Your Name (TEAM_ID)"  # Full certificate name
BUNDLE_ID="com.algiers.app"                                   # App bundle identifier
PROFILE_NAME="notary-api"                                     # Keychain profile name
```

Update these values in the script if your credentials differ.

---

## Manual Signing Commands

### Sign a Binary

```bash
codesign --force --options runtime --sign "Developer ID Application: Your Name (TEAM_ID)" /path/to/binary
```

### Sign an App Bundle (deep)

```bash
codesign --force --options runtime --deep --sign "Developer ID Application: Your Name (TEAM_ID)" /path/to/App.app
```

### Verify Signature

```bash
codesign --verify --deep --strict /path/to/App.app
```

---

## Manual Notarization Commands

### Submit for Notarization

```bash
xcrun notarytool submit /path/to/App.dmg \
    --keychain-profile "notary-api" \
    --wait
```

The `--wait` flag blocks until notarization completes (usually 2-5 minutes).

### Check Submission Status

```bash
xcrun notarytool info <submission-id> --keychain-profile "notary-api"
```

### View Notarization Log (for debugging failures)

```bash
xcrun notarytool log <submission-id> --keychain-profile "notary-api"
```

### Staple the Ticket

After successful notarization, staple the ticket to the DMG:

```bash
xcrun stapler staple /path/to/App.dmg
```

---

## Gatekeeper Verification

After notarization, verify Gatekeeper accepts the app:

```bash
spctl --assess --verbose=4 /path/to/App.app
```

Expected output: `accepted`

---

## Troubleshooting

### "No identity found"

Your Developer ID certificate isn't installed. Download it from:
https://developer.apple.com/account/resources/certificates/list

### "The signature is invalid"

Re-sign with `--force` flag and ensure hardened runtime is enabled (`--options runtime`).

### Notarization "Invalid" status

Check the log for details:
```bash
xcrun notarytool log <submission-id> --keychain-profile "notary-api"
```

Common issues:
- Missing hardened runtime entitlements
- Unsigned helper binaries
- Invalid Info.plist

### "Profile not found"

Re-run the `store-credentials` command to set up the keychain profile.

---

## Current Configuration

| Setting | Value |
|---------|-------|
| Team ID | `6U62M4232W` |
| Bundle ID | `com.algiers.app` |
| Keychain Profile | `notary-api` |
| Certificate | Developer ID Application: Twesh Deshetty |

---

## Quick Reference

```bash
# Full build with signing and notarization
bash scripts/build-and-notarize.sh

# Just sign a DMG
codesign --force --sign "Developer ID Application: Twesh Deshetty (6U62M4232W)" build/Algiers.dmg

# Just notarize (after signing)
xcrun notarytool submit build/Algiers.dmg --keychain-profile "notary-api" --wait
xcrun stapler staple build/Algiers.dmg

# Verify everything
spctl --assess -v build/export/Algiers.app
```
