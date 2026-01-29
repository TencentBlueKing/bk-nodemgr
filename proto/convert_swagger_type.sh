#!/usr/bin/env bash

set -euo pipefail

SWAGGER_ROOT="${1:-}"

if [[ -z "${SWAGGER_ROOT}" ]]; then
  echo "usage: $(basename "$0") <swagger_root_dir>" >&2
  exit 2
fi

if [[ ! -d "${SWAGGER_ROOT}" ]]; then
  echo "swagger_root_dir not found: ${SWAGGER_ROOT}" >&2
  exit 2
fi

sed_inplace() {
  # macOS/BSD sed requires an explicit empty backup suffix.
  # GNU sed accepts `-i` without a suffix.
  if sed --version >/dev/null 2>&1; then
    sed -i -E "$@"
  else
    sed -i '' -E "$@"
  fi
}

# Recursively process all *.swagger.json files.
find "${SWAGGER_ROOT}" -type f -name "*.swagger.json" -print0 | while IFS= read -r -d '' file; do
  sed_inplace '
    /[[:space:]]*"type": "string",/ {
      N
      /[[:space:]]*"type": "string",\n[[:space:]]*"format": "int64"/ {
        s/"type": "string"/"type": "integer"/
      }
    }
    /[[:space:]]*"type": "string",[[:space:]]*"format": "int64"/ {
      s/"type": "string"/"type": "integer"/
    }
  ' "${file}"
done
