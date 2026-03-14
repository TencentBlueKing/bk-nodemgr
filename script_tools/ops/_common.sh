#!/bin/sh
# Shared helper functions for ops scripts.
# Source this file: source "${SCRIPT_DIR}/_common.sh"

# parse_yaml_value extracts a scalar value under a given top-level section.
# $1 = file, $2 = section (e.g. "etcd"), $3 = key (e.g. "username")
parse_yaml_value() {
    local file="$1" section="$2" key="$3"
    awk -v section="$section" -v key="$key" '
    BEGIN { in_section=0 }
    {
        if ($0 ~ "^"section":") { in_section=1; next }
        if (in_section) {
            if (/^[a-zA-Z]/ && $0 !~ "^"section":") { in_section=0; next }
            if (match($0, "^[[:space:]]+"key":[[:space:]]*(.*)$")) {
                val = $0
                sub("^[[:space:]]+"key":[[:space:]]*", "", val)
                gsub(/^["'\''"]|["'\''"]$/, "", val)
                print val
                exit
            }
        }
    }' "$file"
}

# parse_yaml_nested_value extracts a value nested two levels: section.parent.key
# Dynamically detects parent indent level instead of hardcoding.
# $1 = file, $2 = section, $3 = parent, $4 = key
parse_yaml_nested_value() {
    local file="$1" section="$2" parent="$3" key="$4"
    awk -v section="$section" -v parent="$parent" -v key="$key" '
    BEGIN { in_section=0; in_parent=0; parent_indent=-1 }
    {
        if ($0 ~ "^"section":") { in_section=1; next }
        if (in_section) {
            if (/^[a-zA-Z]/ && $0 !~ "^"section":") { in_section=0; next }
            if (in_section && match($0, "^[[:space:]]+"parent":")) {
                in_parent=1
                tmp = $0; gsub(/[^ ].*/, "", tmp); parent_indent = length(tmp)
                next
            }
            if (in_parent) {
                if (match($0, /^[[:space:]]+[a-zA-Z]/)) {
                    line = $0; gsub(/^[[:space:]]+/, "", line)
                    spaces = length($0) - length(line)
                    if (spaces <= parent_indent) { in_parent=0; next }
                }
                if (match($0, "^[[:space:]]+"key":[[:space:]]*(.*)$")) {
                    val = $0
                    sub("^[[:space:]]+"key":[[:space:]]*", "", val)
                    gsub(/^["'\''"]|["'\''"]$/, "", val)
                    print val
                    exit
                }
            }
        }
    }' "$file"
}

# parse_yaml_array extracts array elements from a YAML field.
# Supports both inline ["a","b"] and multi-line "- a\n- b" formats.
# $1 = file, $2 = section, $3 = key
parse_yaml_array() {
    local file="$1" section="$2" key="$3"
    awk -v section="$section" -v key="$key" '
    BEGIN { in_section=0; in_array=0; key_indent=-1 }
    {
        if ($0 ~ "^"section":") { in_section=1; next }
        if (in_section) {
            if (/^[a-zA-Z]/ && $0 !~ "^"section":") { in_section=0; next }

            if (!in_array && match($0, "^[[:space:]]+"key":")) {
                line = $0
                sub("^[[:space:]]+"key":[[:space:]]*", "", line)

                if (line != "" && line != "|" && line != ">") {
                    gsub(/\[/, "", line)
                    gsub(/\]/, "", line)
                    gsub(/"/, "", line)
                    gsub(/'\''/, "", line)
                    n = split(line, items, /[[:space:]]*,[[:space:]]*/)
                    for (i=1; i<=n; i++) {
                        gsub(/^[[:space:]]+|[[:space:]]+$/, "", items[i])
                        if (items[i] != "") print items[i]
                    }
                    exit
                }
                in_array = 1
                tmp = $0; gsub(/[^ ].*/, "", tmp); key_indent = length(tmp)
                next
            }

            if (in_array) {
                if (/^[[:space:]]*$/) next
                tmp = $0; gsub(/[^ ].*/, "", tmp)
                if (length(tmp) <= key_indent && $0 !~ "^[[:space:]]*-") { exit }
                if (match($0, /^[[:space:]]*-[[:space:]]*(.*)$/)) {
                    val = $0
                    sub(/^[[:space:]]*-[[:space:]]*/, "", val)
                    gsub(/^["'\''"]|["'\''"]$/, "", val)
                    if (val != "") print val
                }
            }
        }
    }' "$file"
}

# discover_config finds the first config file containing the target section.
# $1 = section name (e.g. "etcd")
discover_config() {
    local section="$1"
    for f in "${CONF_DIR}"/*_conf.yaml; do
        [ -f "$f" ] || continue
        if grep -q "^${section}:" "$f" 2>/dev/null; then
            echo "$f"
            return 0
        fi
    done
    return 1
}

# urlencode encodes a string for safe use in URIs (e.g. MongoDB connection string).
# Falls back to raw output with a warning if python3 is unavailable.
urlencode() {
    if command -v python3 >/dev/null 2>&1; then
        python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1], safe=''))" "$1"
    else
        echo "Warning: python3 not found, credentials with special characters may cause connection errors" >&2
        printf '%s' "$1"
    fi
}
