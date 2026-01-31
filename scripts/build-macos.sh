#!/bin/bash
# Build script for Algiers macOS standalone app
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
APP_DIR="$BUILD_DIR/Algiers.app"

# Code signing identity
DEVELOPER_ID="Developer ID Application: Twesh Deshetty (6U62M4232W)"
TEAM_ID="6U62M4232W"

echo "=== Algiers macOS Build ==="
echo "Project root: $PROJECT_ROOT"
echo ""

# Parse arguments
CONFIGURATION="Release"
SKIP_WEB=false
SKIP_ENGINE=false
SKIP_ANALYZER=false
SKIP_SIGN=false
SKIP_NOTARIZE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)
            CONFIGURATION="Debug"
            shift
            ;;
        --skip-web)
            SKIP_WEB=true
            shift
            ;;
        --skip-engine)
            SKIP_ENGINE=true
            shift
            ;;
        --skip-analyzer)
            SKIP_ANALYZER=true
            shift
            ;;
        --skip-sign)
            SKIP_SIGN=true
            shift
            ;;
        --skip-notarize)
            SKIP_NOTARIZE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create build directory
mkdir -p "$BUILD_DIR"

# Step 1: Build web assets
if [ "$SKIP_WEB" = false ]; then
    echo ">>> Building web assets..."
    cd "$PROJECT_ROOT/web"
    npm install --silent
    npm run build
    echo "Web assets built: $PROJECT_ROOT/web/dist"
fi

# Step 2: Build Go engine
if [ "$SKIP_ENGINE" = false ]; then
    echo ""
    echo ">>> Building Go engine..."
    cd "$PROJECT_ROOT"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BUILD_DIR/algiers-engine" ./cmd/engine
    echo "Engine built: $BUILD_DIR/algiers-engine"
fi

# Step 3: Build Swift analyzer (required)
if [ "$SKIP_ANALYZER" = false ]; then
    echo ""
    echo ">>> Building Swift analyzer..."
    cd "$PROJECT_ROOT/analyzer-swift"
    swift build -c release
    cp ".build/release/analyzer-swift" "$BUILD_DIR/analyzer-swift"
    echo "Analyzer built: $BUILD_DIR/analyzer-swift"
fi

# Step 4: Build Xcode project with code signing
echo ""
echo ">>> Building Xcode project..."
cd "$PROJECT_ROOT/Algiers"

# Build without Xcode signing - we'll sign manually after adding helpers
xcodebuild -project Algiers.xcodeproj \
    -scheme Algiers \
    -configuration "$CONFIGURATION" \
    -derivedDataPath "$BUILD_DIR/DerivedData" \
    SYMROOT="$BUILD_DIR" \
    CODE_SIGN_IDENTITY="-" \
    CODE_SIGNING_REQUIRED=NO \
    CODE_SIGNING_ALLOWED=NO \
    build 2>&1 | grep -E "^(Build|Compile|Link|error:|warning:|\*\*)" || true

# Find the built app
BUILT_APP="$BUILD_DIR/$CONFIGURATION/Algiers.app"
if [ ! -d "$BUILT_APP" ]; then
    echo "Error: Built app not found at $BUILT_APP"
    echo "Checking alternative locations..."
    find "$BUILD_DIR" -name "Algiers.app" -type d
    exit 1
fi

# Step 5: Copy helpers into app bundle
echo ""
echo ">>> Packaging app bundle..."
HELPERS_DIR="$BUILT_APP/Contents/Helpers"
RESOURCES_DIR="$BUILT_APP/Contents/Resources"
mkdir -p "$HELPERS_DIR"
mkdir -p "$RESOURCES_DIR/web"

# Copy engine
cp "$BUILD_DIR/algiers-engine" "$HELPERS_DIR/"
chmod +x "$HELPERS_DIR/algiers-engine"

# Copy analyzer binary
cp "$BUILD_DIR/analyzer-swift" "$HELPERS_DIR/"
chmod +x "$HELPERS_DIR/analyzer-swift"

# Copy OpenL3 model directly (avoid bundle format issues with code signing)
OPENL3_MODEL="$PROJECT_ROOT/analyzer-swift/.build/release/AnalyzerSwift_AnalyzerSwift.bundle/OpenL3Music.mlpackage"
if [ -d "$OPENL3_MODEL" ]; then
    mkdir -p "$RESOURCES_DIR/Models"
    cp -r "$OPENL3_MODEL" "$RESOURCES_DIR/Models/"
    echo "OpenL3 model copied to Resources/Models"
fi

# Copy web assets
cp -r "$PROJECT_ROOT/web/dist/"* "$RESOURCES_DIR/web/"

echo "App packaged: $BUILT_APP"

# Step 6: Sign all helper binaries and the app
if [ "$SKIP_SIGN" = false ]; then
    echo ""
    echo ">>> Code signing..."

    # Sign helper binaries with hardened runtime
    codesign --force --options runtime --sign "$DEVELOPER_ID" "$HELPERS_DIR/algiers-engine"
    echo "  Signed: algiers-engine"

    codesign --force --options runtime --sign "$DEVELOPER_ID" "$HELPERS_DIR/analyzer-swift"
    echo "  Signed: analyzer-swift"

    # Note: AnalyzerSwift_AnalyzerSwift.bundle is a resources bundle (mlpackage)
    # It will be signed as part of the app bundle

    # Re-sign the entire app bundle
    codesign --force --options runtime --deep --sign "$DEVELOPER_ID" "$BUILT_APP"
    echo "  Signed: Algiers.app"

    # Verify signature
    echo ""
    echo ">>> Verifying signature..."
    codesign --verify --deep --strict "$BUILT_APP"
    echo "  Signature verified!"
fi

# Step 7: Copy to top-level build directory
rm -rf "$BUILD_DIR/Algiers.app" 2>/dev/null || true
cp -r "$BUILT_APP" "$BUILD_DIR/Algiers.app"

# Step 8: Create DMG
echo ""
echo ">>> Creating DMG..."
DMG_NAME="Algiers-v0.6-beta-AppleSilicon.dmg"
rm -f "$BUILD_DIR/$DMG_NAME"
hdiutil create -volname "Algiers" -srcfolder "$BUILD_DIR/Algiers.app" -ov -format UDZO "$BUILD_DIR/$DMG_NAME"
echo "DMG created: $BUILD_DIR/$DMG_NAME"

# Step 9: Sign DMG
if [ "$SKIP_SIGN" = false ]; then
    echo ""
    echo ">>> Signing DMG..."
    codesign --force --sign "$DEVELOPER_ID" "$BUILD_DIR/$DMG_NAME"
    echo "  DMG signed!"
fi

# Step 10: Notarize (if not skipped)
if [ "$SKIP_SIGN" = false ] && [ "$SKIP_NOTARIZE" = false ]; then
    echo ""
    echo ">>> Submitting for notarization..."
    echo "  This may take a few minutes..."

    # Submit for notarization
    xcrun notarytool submit "$BUILD_DIR/$DMG_NAME" \
        --keychain-profile "notarytool-profile" \
        --wait 2>&1 || {
        echo ""
        echo "⚠️  Notarization failed or profile not found."
        echo "To set up notarization, run:"
        echo "  xcrun notarytool store-credentials notarytool-profile --apple-id YOUR_APPLE_ID --team-id $TEAM_ID"
        echo ""
        echo "The DMG is signed but not notarized. Users will need to right-click > Open."
    }

    # Staple the notarization ticket if successful
    if xcrun stapler staple "$BUILD_DIR/$DMG_NAME" 2>/dev/null; then
        echo "  Notarization ticket stapled!"
    fi
fi

echo ""
echo "=== Build Complete ==="
echo "App location: $BUILD_DIR/Algiers.app"
echo "DMG location: $BUILD_DIR/$DMG_NAME"
echo ""
echo "To run:"
echo "  open $BUILD_DIR/Algiers.app"
