#!/bin/bash
# Build script for Algiers macOS standalone app
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_DIR="$PROJECT_ROOT/build"
APP_DIR="$BUILD_DIR/Algiers.app"

echo "=== Algiers macOS Build ==="
echo "Project root: $PROJECT_ROOT"
echo ""

# Parse arguments
CONFIGURATION="Release"
SKIP_WEB=false
SKIP_ENGINE=false
SKIP_ANALYZER=false

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

# Step 4: Build Xcode project
echo ""
echo ">>> Building Xcode project..."
cd "$PROJECT_ROOT/Algiers"
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

# Copy analyzer resources bundle (contains OpenL3 model)
ANALYZER_BUNDLE="$PROJECT_ROOT/analyzer-swift/.build/release/AnalyzerSwift_AnalyzerSwift.bundle"
if [ -d "$ANALYZER_BUNDLE" ]; then
    cp -r "$ANALYZER_BUNDLE" "$HELPERS_DIR/"
    echo "OpenL3 model bundle copied"
fi

# Copy web assets
cp -r "$PROJECT_ROOT/web/dist/"* "$RESOURCES_DIR/web/"

echo "App packaged: $BUILT_APP"

# Step 6: Copy to top-level build directory
cp -r "$BUILT_APP" "$BUILD_DIR/Algiers.app" 2>/dev/null || true

echo ""
echo "=== Build Complete ==="
echo "App location: $BUILD_DIR/Algiers.app"
echo ""
echo "To run:"
echo "  open $BUILD_DIR/Algiers.app"
echo ""
echo "To sign for distribution:"
echo "  1. Open Algiers/Algiers.xcodeproj in Xcode"
echo "  2. Select your Team in Signing & Capabilities"
echo "  3. Archive and export for distribution"
