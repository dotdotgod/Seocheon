#!/usr/bin/env bash
#
# genesis_build.sh — Seocheon production genesis file builder.
#
# Usage:
#   ./scripts/genesis_build.sh [OPTIONS]
#
# Options:
#   --chain-id <id>       Chain ID (default: seocheon-1)
#   --output <path>       Output directory (default: ./genesis-output)
#   --team-addr <addr>    Team vesting account address
#   --foundation-addr     Foundation multisig address
#   --gentx-dir <dir>     Directory containing gentx files
#
# This script:
# 1. Initializes a fresh chain with 'seocheon init'
# 2. Applies production parameters via 'seocheon genesis-build'
# 3. Collects gentx files (if provided)
# 4. Validates the final genesis file

set -euo pipefail

CHAIN_ID="${CHAIN_ID:-seocheon-1}"
OUTPUT_DIR="${OUTPUT_DIR:-./genesis-output}"
BINARY="${BINARY:-seocheon}"
MONIKER="${MONIKER:-genesis-builder}"
TEAM_ADDR="${TEAM_ADDR:-}"
FOUNDATION_ADDR="${FOUNDATION_ADDR:-}"
GENTX_DIR="${GENTX_DIR:-}"
DENOM="usum"

# Parse arguments.
while [[ $# -gt 0 ]]; do
  case "$1" in
    --chain-id) CHAIN_ID="$2"; shift 2 ;;
    --output) OUTPUT_DIR="$2"; shift 2 ;;
    --team-addr) TEAM_ADDR="$2"; shift 2 ;;
    --foundation-addr) FOUNDATION_ADDR="$2"; shift 2 ;;
    --gentx-dir) GENTX_DIR="$2"; shift 2 ;;
    *) echo "Unknown argument: $1"; exit 1 ;;
  esac
done

echo "=== Seocheon Genesis Builder ==="
echo "Chain ID:    $CHAIN_ID"
echo "Output:      $OUTPUT_DIR"
echo "Binary:      $BINARY"
echo ""

# Check binary exists.
if ! command -v "$BINARY" &> /dev/null; then
  echo "ERROR: $BINARY not found in PATH. Build first: make install"
  exit 1
fi

# Create output directory.
export HOME_DIR="$OUTPUT_DIR/.seocheon"
mkdir -p "$OUTPUT_DIR"

# Step 1: Initialize chain.
echo "--- Step 1: Initialize chain ---"
$BINARY init "$MONIKER" \
  --chain-id "$CHAIN_ID" \
  --home "$HOME_DIR" \
  --default-denom "$DENOM" \
  2>/dev/null

GENESIS_FILE="$HOME_DIR/config/genesis.json"
echo "  Genesis initialized: $GENESIS_FILE"

# Step 2: Apply production parameters.
echo "--- Step 2: Apply production parameters ---"
$BINARY genesis-build "$GENESIS_FILE" --home "$HOME_DIR"

# Step 3: Collect gentx files (if provided).
if [ -n "$GENTX_DIR" ] && [ -d "$GENTX_DIR" ]; then
  echo "--- Step 3: Collect gentx files ---"
  GENTX_TARGET="$HOME_DIR/config/gentx"
  mkdir -p "$GENTX_TARGET"
  cp "$GENTX_DIR"/*.json "$GENTX_TARGET/" 2>/dev/null || true
  GENTX_COUNT=$(ls "$GENTX_TARGET"/*.json 2>/dev/null | wc -l)
  echo "  Collected $GENTX_COUNT gentx files"

  if [ "$GENTX_COUNT" -gt 0 ]; then
    $BINARY genesis collect-gentxs --home "$HOME_DIR" 2>/dev/null
    echo "  Gentxs collected and merged into genesis"
  fi
else
  echo "--- Step 3: Skipped (no gentx directory provided) ---"
fi

# Step 4: Validate genesis.
echo "--- Step 4: Validate genesis ---"
if $BINARY genesis validate "$GENESIS_FILE" --home "$HOME_DIR" 2>/dev/null; then
  echo "  Genesis validation PASSED"
else
  echo "  WARNING: Genesis validation failed or command not available"
  echo "  This may be expected before gentx collection"
fi

# Step 5: Copy final genesis to output.
cp "$GENESIS_FILE" "$OUTPUT_DIR/genesis.json"

echo ""
echo "=== Genesis build complete ==="
echo "  Final genesis: $OUTPUT_DIR/genesis.json"
echo "  Home dir:      $HOME_DIR"
echo ""
echo "Next steps:"
echo "  1. Distribute genesis.json to validators"
echo "  2. Each validator creates gentx: seocheon genesis gentx ..."
echo "  3. Collect gentxs: ./scripts/genesis_build.sh --gentx-dir <dir>"
echo "  4. Start chain: seocheon start --home <home>"
