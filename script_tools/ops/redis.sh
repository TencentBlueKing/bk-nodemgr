#!/bin/sh
# Ops helper: connect to Redis via bk-nodemgr-adminclient.
# Config loading (including kms secret override) and connection parameters
# are handled by adminclient.
# Usage: redis.sh [-c config] [-h] [login|status]

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONF_DIR="${CONF_DIR:-/bk-nodemgr/etc}"
CONFIG_FILE=""

. "${SCRIPT_DIR}/_common.sh"

usage() {
    cat <<EOF
Usage: $(basename "$0") [-c config_file] [-h] [login|status]

Connect to Redis via bk-nodemgr-adminclient with connection parameters from
the service config file in ${CONF_DIR} (kms secret override applied when kms
is enabled in the config).

Commands:
  login    Start an interactive redis-cli session (default)
  status   Check Redis connectivity and exit

Options:
  -c CONFIG  Specify config file path (default: auto-discover)
  -h         Show this help message

Supported Redis types: standalone, cluster, sentinel

Examples:
  $(basename "$0")
  $(basename "$0") login
  $(basename "$0") status
  $(basename "$0") -c /bk-nodemgr/etc/backend_conf.yaml login
EOF
}

# --- Parse script options ---
while [ $# -gt 0 ]; do
    case "$1" in
        -c)
            shift
            CONFIG_FILE="${1:-}"
            [ -z "$CONFIG_FILE" ] && { echo "Error: -c requires a config file path" >&2; exit 1; }
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            break
            ;;
    esac
done

ADMINCLIENT="$(find_adminclient)" || {
    echo "Error: bk-nodemgr-adminclient not found" >&2
    exit 1
}

if [ -n "$CONFIG_FILE" ]; then
    [ -f "$CONFIG_FILE" ] || { echo "Error: config file not found: ${CONFIG_FILE}" >&2; exit 1; }
    set -- -f "$CONFIG_FILE" "$@"
fi
set -- --conf-dir "$CONF_DIR" "$@"

exec "$ADMINCLIENT" database redis "$@"
