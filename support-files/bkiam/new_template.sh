#!/bin/bash
# Create a new IAM migration template file
#
# Usage: ./new_template.sh <description>
# Example: ./new_template.sh add_resource_type

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <description>"
    echo "Example: $0 add_resource_type"
    exit 1
fi

DESCRIPTION="$1"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
TEMPLATES_DIR="${SCRIPT_DIR}/templates"

# Find the next sequence number
LAST_NUM=$(ls -1 "$TEMPLATES_DIR"/*.json.tpl 2>/dev/null | sed 's/.*\/\([0-9]*\)_.*/\1/' | sort -n | tail -1)
if [ -z "$LAST_NUM" ]; then
    NEXT_NUM="0001"
else
    NEXT_NUM=$(printf "%04d" $((10#$LAST_NUM + 1)))
fi

FILENAME="${NEXT_NUM}_bk_nodemgr_${DESCRIPTION}.json.tpl"
FILEPATH="${TEMPLATES_DIR}/${FILENAME}"

cat > "$FILEPATH" << 'EOF'
{
    "system_id": "bk-nodemgr",
    "operations": [
        {
            "operation": "upsert_resource_type",
            "data": {
                "id": "your_resource_type",
                "name": "资源类型名称",
                "name_en": "Resource Type Name",
                "description": "资源类型描述",
                "description_en": "Resource type description",
                "version": 1
            }
        }
    ]
}
EOF

echo "Created: $FILENAME"
