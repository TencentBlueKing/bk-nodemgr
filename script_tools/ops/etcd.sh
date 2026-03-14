#!/bin/sh
# Ops helper: auto-discover etcd config and proxy etcdctl commands.
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

Auto-discover etcd configuration from ${CONF_DIR}/*_conf.yaml and
proxy commands to etcdctl with the correct connection parameters.

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

# --- Discover or validate config file ---
if [ -z "$CONFIG_FILE" ]; then
    CONFIG_FILE="$(discover_config "etcd")" || {
        echo "Error: no config file containing 'etcd:' found in ${CONF_DIR}/" >&2
        echo "Use -c <config_file> to specify manually." >&2
        exit 1
    }
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: config file not found: ${CONFIG_FILE}" >&2
    exit 1
fi

echo "Using config: ${CONFIG_FILE}" >&2

# --- Parse etcd configuration ---
ENDPOINTS=""
_eps=$(parse_yaml_array "$CONFIG_FILE" "etcd" "endpoints")
if [ -n "$_eps" ]; then
    while IFS= read -r ep; do
        [ -n "$ep" ] || continue
        [ -n "$ENDPOINTS" ] && ENDPOINTS="${ENDPOINTS},"
        ENDPOINTS="${ENDPOINTS}${ep}"
    done <<EOF
$_eps
EOF
fi

USERNAME="$(parse_yaml_value "$CONFIG_FILE" "etcd" "username")"
PASSWORD="$(parse_yaml_value "$CONFIG_FILE" "etcd" "password")"
CA_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "etcd" "tls" "caFile")"
CERT_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "etcd" "tls" "certFile")"
KEY_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "etcd" "tls" "keyFile")"

# --- Build etcdctl command via set -- (prepend flags before user args) ---
[ -n "$KEY_FILE" ] && [ -f "$KEY_FILE" ] && set -- "--key=$KEY_FILE" "$@"
[ -n "$CERT_FILE" ] && [ -f "$CERT_FILE" ] && set -- "--cert=$CERT_FILE" "$@"
[ -n "$CA_FILE" ] && [ -f "$CA_FILE" ] && set -- "--cacert=$CA_FILE" "$@"
[ -n "$USERNAME" ] && [ -n "$PASSWORD" ] && set -- "--user=${USERNAME}:${PASSWORD}" "$@"
[ -n "$ENDPOINTS" ] && set -- "--endpoints=$ENDPOINTS" "$@"
set -- etcdctl "$@"

# --- Execute with masked debug output ---
SAFE_CMD=$(echo "$*" | sed 's/--user=[^ ]*/--user=***:***/')
echo "Executing: ${SAFE_CMD}" >&2
exec "$@"
