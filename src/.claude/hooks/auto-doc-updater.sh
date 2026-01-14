#!/bin/bash
# auto-doc-updater.sh - Shell proxy for Go hook implementation
# This script calls the claudex-hooks binary which contains the actual logic.

# Find the hooks binary (installed alongside claudex)
HOOKS_BIN="${CLAUDEX_HOOKS_BIN:-claudex-hooks}"

# Execute the appropriate subcommand, passing stdin through
exec "$HOOKS_BIN" auto-doc
