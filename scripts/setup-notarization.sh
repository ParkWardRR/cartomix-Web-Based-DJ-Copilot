#!/bin/bash
# Setup notarization credentials for Algiers
set -e

TEAM_ID="6U62M4232W"
PROFILE_NAME="notary-api"

echo ""
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║           Apple Notarization Setup for Algiers                ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# Check if profile already exists
if xcrun notarytool history --keychain-profile "$PROFILE_NAME" > /dev/null 2>&1; then
    echo "✓ Notarization credentials already configured!"
    echo ""
    echo "Run the build script:"
    echo "  bash scripts/build-and-notarize.sh"
    exit 0
fi

echo "To notarize apps for Gatekeeper, you need:"
echo ""
echo "  1. Your Apple ID email (developer account)"
echo "  2. An App-Specific Password"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "CREATE APP-SPECIFIC PASSWORD:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  1. Open: https://appleid.apple.com"
echo "  2. Sign in → 'Sign-In and Security'"
echo "  3. Click 'App-Specific Passwords' → Generate"
echo "  4. Name: 'Algiers Notarization'"
echo "  5. Copy the password (xxxx-xxxx-xxxx-xxxx)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
read -p "Press ENTER when you have your app-specific password... "
echo ""

echo "Storing credentials in macOS Keychain..."
echo "(Your password will be stored securely and never shown)"
echo ""

# Run the interactive credential setup
xcrun notarytool store-credentials "$PROFILE_NAME" \
    --team-id "$TEAM_ID"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Verifying credentials..."
if xcrun notarytool history --keychain-profile "$PROFILE_NAME" > /dev/null 2>&1; then
    echo ""
    echo "✅ Credentials verified!"
    echo ""
    echo "Now run:"
    echo "  bash scripts/build-and-notarize.sh"
    echo ""
else
    echo ""
    echo "❌ Verification failed. Please try again."
    exit 1
fi
