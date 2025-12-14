#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
NPM_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

VERSION=$(cat "$NPM_DIR/version.txt")

if [ -z "$VERSION" ]; then
  echo "Error: version.txt is empty or missing" >&2
  exit 1
fi

echo "Syncing version $VERSION to all packages..."

# Update main package (@claudex/cli)
jq ".version = \"$VERSION\" | .optionalDependencies[\"@claudex/darwin-arm64\"] = \"$VERSION\" | .optionalDependencies[\"@claudex/darwin-x64\"] = \"$VERSION\" | .optionalDependencies[\"@claudex/linux-x64\"] = \"$VERSION\" | .optionalDependencies[\"@claudex/linux-arm64\"] = \"$VERSION\"" \
  "$NPM_DIR/@claudex/cli/package.json" > "$NPM_DIR/@claudex/cli/package.json.tmp" && \
  mv "$NPM_DIR/@claudex/cli/package.json.tmp" "$NPM_DIR/@claudex/cli/package.json"
echo "✓ Updated @claudex/cli"

# Update platform packages
for platform in darwin-arm64 darwin-x64 linux-x64 linux-arm64; do
  pkg_json="$NPM_DIR/@claudex/$platform/package.json"
  jq ".version = \"$VERSION\"" "$pkg_json" > "$pkg_json.tmp" && mv "$pkg_json.tmp" "$pkg_json"
  echo "✓ Updated @claudex/$platform"
done

echo "Version sync complete: $VERSION"
