#!/bin/sh
# Ops helper: proxy etcdctl commands via bk-nodemgr-adminclient.
# Config loading (including kms secret override) and connection parameters
# are handled by adminclient.
# Usage: etcd.sh [-c config] [-h] <etcdctl-subcommand> [args...]
# Example: etcd.sh member list
#          etcd.sh endpoint health
#          etcd.sh get /some/key

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONF_DIR="${CONF_DIR:-/bk-nodemgr/etc}"
CONFIG_FILE=""

. "${SCRIPT_DIR}/_common.sh"

usage() {
    cat <<EOF
Usage: $(basename "$0") [-c config_file] [-h] <etcdctl-subcommand> [args...]

Proxy etcdctl commands via bk-nodemgr-adminclient with connection parameters
from the service config file in ${CONF_DIR} (kms secret override applied when
kms is enabled in the config).

Options:
  -c CONFIG  Specify config file path (default: auto-discover)
  -h         Show this help message

Examples:
  $(basename "$0") member list
  $(basename "$0") endpoint health
  $(basename "$0") get /bk-nodemgr/some/key
  $(basename "$0") -c /bk-nodemgr/etc/backend_conf.yaml endpoint status
EOF
}

# --- Parse script options (before etcdctl subcommand) ---
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

if [ $# -eq 0 ]; then
    echo "Error: no etcdctl subcommand provided" >&2
    echo "" >&2
    usage >&2
    exit 1
fi

ADMINCLIENT="$(find_adminclient)" || {
    echo "Error: bk-nodemgr-adminclient not found" >&2
    exit 1
}

# --- Translate to adminclient args: database etcd [--conf-dir dir] [-f config] -- <etcdctl args> ---
if [ -n "$CONFIG_FILE" ]; then
    [ -f "$CONFIG_FILE" ] || { echo "Error: config file not found: ${CONFIG_FILE}" >&2; exit 1; }
    set -- -f "$CONFIG_FILE" -- "$@"
else
    set -- -- "$@"
fi
set -- --conf-dir "$CONF_DIR" "$@"

exec "$ADMINCLIENT" database etcd "$@"
