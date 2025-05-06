#!/bin/bash

SWAGGER_ROOT="$1"

# recursively process all Swagger.JSON files.
find "${SWAGGER_ROOT}" -type f -name "*.swagger.json" -print0 | while IFS= read -r -d '' file; do
  # Modify files in place using GNU sed
  sed -i -E '
    /[[:space:]]*"type": "string",/ {
      N  
      /[[:space:]]*"type": "string",\n[[:space:]]*"format": "int64"/ {
        s/"type": "string"/"type": "integer"/  # replace type
      }
    }
    # Handle the format of the same line (type and format on the same line)
    /[[:space:]]*"type": "string",[[:space:]]*"format": "int64"/ {
      s/"type": "string"/"type": "integer"/  # replace type
    }
  ' "${file}"
done
