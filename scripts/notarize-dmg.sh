#!/bin/bash
# Notarize an existing signed DMG
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
PROFILE_NAME="notary-api"
DMG_PATH="${1:-$BUILD_DIR/Algiers-v0.6-beta-AppleSilicon.dmg}"

echo "=== Notarize DMG ==="
echo ""

# Check DMG exists
if [ ! -f "$DMG_PATH" ]; then
    echo "❌ DMG not found: $DMG_PATH"
    exit 1
fi

echo "DMG: $DMG_PATH"
echo ""

# Check credentials
if ! xcrun notarytool history --keychain-profile "$PROFILE_NAME" > /dev/null 2>&1; then
    echo "❌ Notarization credentials not found."
    echo ""
    echo "Run first: bash scripts/setup-notarization.sh"
    exit 1
fi

echo ">>> Submitting for notarization..."
echo "    This typically takes 2-5 minutes..."
echo ""

xcrun notarytool submit "$DMG_PATH" \
    --keychain-profile "$PROFILE_NAME" \
    --wait

echo ""
echo ">>> Stapling notarization ticket..."
xcrun stapler staple "$DMG_PATH"

echo ""
echo ">>> Verifying Gatekeeper..."
if spctl --assess --type open --context context:primary-signature -v "$DMG_PATH" 2>&1 | grep -q "accepted"; then
    echo "✅ Gatekeeper: ACCEPTED"
else
    spctl --assess --type open --context context:primary-signature -v "$DMG_PATH" 2>&1
fi

echo ""
echo "=== Notarization Complete ==="
echo ""
echo "DMG: $DMG_PATH"
echo ""
echo "Users can now download and open without Gatekeeper warnings!"
