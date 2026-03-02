#!/usr/bin/env bash
# gen-mock-config.sh — Merge CMDB mock data into mock-server config.
# Usage: gen-mock-config.sh <base-config> <output-path>
# Reads test/data/cmdb.yaml and appends it as mockData.cmdb section.

set -euo pipefail

usage () {
    cat <<EOF
Usage:
    $0 <base-config> <output-path>
EOF
}

usage_and_exit () {
    usage
    exit "$1"
}

error () {
    echo "$@" 1>&2
    exit 1
}

(( $# != 2 )) && usage_and_exit 1

BASE_CONFIG="$1"
OUTPUT="$2"
SCRIPT_DIR="$(realpath "$(dirname "$0")")"
DATA_DIR="${SCRIPT_DIR}/../data"
CMDB_DATA="${DATA_DIR}/cmdb.yaml"

if ! [[ -f "$BASE_CONFIG" ]]; then
    error "base config not found: $BASE_CONFIG"
fi

if ! [[ -f "$CMDB_DATA" ]]; then
    error "cmdb data not found: $CMDB_DATA"
fi

# create output directory if it doesn't exist
OUTPUT_DIR="$(dirname "$OUTPUT")"
mkdir -p "$OUTPUT_DIR"

BASE_REAL="$(realpath "$BASE_CONFIG")"
OUTPUT_REAL="$(realpath -m "$OUTPUT")"
if [[ "$BASE_REAL" != "$OUTPUT_REAL" ]]; then
    cp "$BASE_CONFIG" "$OUTPUT"
fi

# append mockData.cmdb section to the output file.
echo "" >> "$OUTPUT"
echo "mockData:" >> "$OUTPUT"
echo "  cmdb:" >> "$OUTPUT"
sed 's/^/    /' "$CMDB_DATA" >> "$OUTPUT"
