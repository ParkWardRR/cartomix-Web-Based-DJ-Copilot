#!/bin/bash
# Build, sign, notarize, and package Algiers for distribution
#
# SETUP REQUIRED: Before running this script, you must configure notarization credentials.
# See docs/SIGNING.md for full instructions, or run:
#
#   xcrun notarytool store-credentials "notary-api" \
#       --apple-id "YOUR_APPLE_ID" \
#       --team-id "YOUR_TEAM_ID" \
#       --password "YOUR_APP_SPECIFIC_PASSWORD"
#
# Get an app-specific password at: https://appleid.apple.com → Security → App-Specific Passwords
#
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"

# Configuration
TEAM_ID="6U62M4232W"
DEVELOPER_ID="Developer ID Application: Twesh Deshetty (6U62M4232W)"
BUNDLE_ID="com.algiers.app"
VERSION="1.7-beta"
PROFILE_NAME="notary-api"

echo "=== Algiers Build & Notarize ==="
echo ""

# Check for notarization credentials
check_credentials() {
    if ! xcrun notarytool history --keychain-profile "$PROFILE_NAME" > /dev/null 2>&1; then
        echo "⚠️  Notarization credentials not found."
        echo ""
        echo "To enable notarization, you need to set up credentials:"
        echo ""
        echo "1. Go to https://appleid.apple.com"
        echo "2. Sign in and go to Security → App-Specific Passwords"
        echo "3. Generate a new password for 'Algiers Notarization'"
        echo "4. Run this command:"
        echo ""
        echo "   xcrun notarytool store-credentials \"$PROFILE_NAME\" \\"
        echo "       --apple-id YOUR_APPLE_ID_EMAIL \\"
        echo "       --team-id $TEAM_ID \\"
        echo "       --password YOUR_APP_SPECIFIC_PASSWORD"
        echo ""
        echo "Then run this script again."
        echo ""
        return 1
    fi
    return 0
}

# Step 1: Build web assets
echo ">>> Building web assets..."
cd "$PROJECT_ROOT/web"
npm install --silent 2>/dev/null
npm run build
echo "✓ Web assets built"

# Step 2: Build Go engine
echo ""
echo ">>> Building Go engine..."
cd "$PROJECT_ROOT"
CGO_ENABLED=1 go build -ldflags="-s -w" -o "$BUILD_DIR/algiers-engine" ./cmd/engine
echo "✓ Engine built"

# Step 3: Build Swift analyzer
echo ""
echo ">>> Building Swift analyzer..."
cd "$PROJECT_ROOT/analyzer-swift"
swift build -c release 2>&1 | grep -E "Build complete|error:" || true
cp ".build/release/analyzer-swift" "$BUILD_DIR/analyzer-swift"
echo "✓ Analyzer built"

# Step 4: Archive with Xcode
echo ""
echo ">>> Archiving Xcode project..."
cd "$PROJECT_ROOT/Algiers"
rm -rf "$BUILD_DIR/Algiers.xcarchive"
xcodebuild -project Algiers.xcodeproj \
    -scheme Algiers \
    -configuration Release \
    -archivePath "$BUILD_DIR/Algiers.xcarchive" \
    archive 2>&1 | grep -E "ARCHIVE|error:" || true
echo "✓ Archive created"

# Step 5: Export archive
echo ""
echo ">>> Exporting archive..."
rm -rf "$BUILD_DIR/export"
xcodebuild -exportArchive \
    -archivePath "$BUILD_DIR/Algiers.xcarchive" \
    -exportPath "$BUILD_DIR/export" \
    -exportOptionsPlist "$PROJECT_ROOT/scripts/ExportOptions.plist" 2>&1 | grep -E "EXPORT|error:" || true
echo "✓ Archive exported"

# Step 6: Add helpers to app bundle
echo ""
echo ">>> Adding helpers to app bundle..."
APP_PATH="$BUILD_DIR/export/Algiers.app"
HELPERS_DIR="$APP_PATH/Contents/Helpers"
RESOURCES_DIR="$APP_PATH/Contents/Resources"

mkdir -p "$HELPERS_DIR"
mkdir -p "$RESOURCES_DIR/web"
mkdir -p "$RESOURCES_DIR/Models"

# Copy engine
cp "$BUILD_DIR/algiers-engine" "$HELPERS_DIR/"
chmod +x "$HELPERS_DIR/algiers-engine"

# Copy analyzer
cp "$BUILD_DIR/analyzer-swift" "$HELPERS_DIR/"
chmod +x "$HELPERS_DIR/analyzer-swift"

# Copy OpenL3 model
OPENL3_MODEL="$PROJECT_ROOT/analyzer-swift/.build/release/AnalyzerSwift_AnalyzerSwift.bundle/OpenL3Music.mlpackage"
if [ -d "$OPENL3_MODEL" ]; then
    cp -r "$OPENL3_MODEL" "$RESOURCES_DIR/Models/"
    echo "  ✓ OpenL3 model copied"
fi

# Copy web assets
cp -r "$PROJECT_ROOT/web/dist/"* "$RESOURCES_DIR/web/"
echo "  ✓ Web assets copied"

# Step 7: Sign all components
echo ""
echo ">>> Code signing..."

# Sign helpers with hardened runtime
codesign --force --options runtime --sign "$DEVELOPER_ID" "$HELPERS_DIR/algiers-engine"
echo "  ✓ Signed: algiers-engine"

codesign --force --options runtime --sign "$DEVELOPER_ID" "$HELPERS_DIR/analyzer-swift"
echo "  ✓ Signed: analyzer-swift"

# Re-sign the entire app bundle (deep sign)
codesign --force --options runtime --deep --sign "$DEVELOPER_ID" "$APP_PATH"
echo "  ✓ Signed: Algiers.app"

# Verify signature
codesign --verify --deep --strict "$APP_PATH"
echo "  ✓ Signature verified"

# Step 8: Create DMG
echo ""
echo ">>> Creating DMG..."
DMG_NAME="Algiers-v${VERSION}-AppleSilicon.dmg"
rm -f "$BUILD_DIR/$DMG_NAME"
# Use unique volume name to avoid conflicts with previous mounts
hdiutil create -volname "Algiers v${VERSION}" -srcfolder "$APP_PATH" -ov -format UDZO "$BUILD_DIR/$DMG_NAME"
echo "  ✓ DMG created"

# Sign DMG
codesign --force --sign "$DEVELOPER_ID" "$BUILD_DIR/$DMG_NAME"
echo "  ✓ DMG signed"

# Step 9: Notarize
echo ""
echo ">>> Checking notarization credentials..."

if check_credentials; then
    echo ""
    echo ">>> Submitting for notarization..."
    echo "  This may take 2-5 minutes..."

    # Submit for notarization and wait
    xcrun notarytool submit "$BUILD_DIR/$DMG_NAME" \
        --keychain-profile "$PROFILE_NAME" \
        --wait

    NOTARIZE_STATUS=$?

    if [ $NOTARIZE_STATUS -eq 0 ]; then
        echo ""
        echo ">>> Stapling notarization ticket..."
        xcrun stapler staple "$BUILD_DIR/$DMG_NAME"
        echo "  ✓ Notarization ticket stapled"

        # Verify Gatekeeper assessment
        echo ""
        echo ">>> Verifying Gatekeeper..."
        if spctl --assess --verbose=4 "$APP_PATH" 2>&1 | grep -q "accepted"; then
            echo "  ✓ Gatekeeper: ACCEPTED"
        else
            echo "  ⚠ Gatekeeper check - run 'spctl --assess' manually"
        fi
    else
        echo ""
        echo "⚠️  Notarization failed. Check logs with:"
        echo "   xcrun notarytool log <submission-id> --keychain-profile $PROFILE_NAME"
    fi
else
    echo ""
    echo "⚠️  Skipping notarization (credentials not set up)"
    echo "The DMG is signed but not notarized."
    echo "Users will need to right-click > Open to launch."
fi

# Final output
echo ""
echo "=== Build Complete ==="
echo ""
echo "App:  $APP_PATH"
echo "DMG:  $BUILD_DIR/$DMG_NAME"
echo ""
echo "To test locally:"
echo "  open $BUILD_DIR/$DMG_NAME"
