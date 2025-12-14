#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NPM_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Check if logged in to npm
if ! npm whoami &>/dev/null; then
  echo "Error: Not logged in to npm. Run 'npm login' first." >&2
  exit 1
fi

# Sync versions first
"$SCRIPT_DIR/sync-version.sh"

echo ""
echo "Publishing platform packages first (main package depends on them)..."

# Publish platform packages
for platform in darwin-arm64 darwin-x64 linux-x64 linux-arm64; do
  echo "Publishing @claudex/$platform..."
  (cd "$NPM_DIR/@claudex/$platform" && npm publish --access public)
done

echo ""
echo "Publishing main package (@claudex/cli)..."
(cd "$NPM_DIR/@claudex/cli" && npm publish --access public)

echo ""
echo "✓ All packages published successfully!"
