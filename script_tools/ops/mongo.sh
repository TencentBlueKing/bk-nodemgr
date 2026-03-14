#!/bin/sh
# Ops helper: auto-discover MongoDB config and connect via mongo shell.
# Usage: mongo.sh [-c config] [-h] [login|status]

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONF_DIR="${CONF_DIR:-/bk-nodemgr/etc}"
CONFIG_FILE=""
DEFAULT_MONGO_PORT="27017"

. "${SCRIPT_DIR}/_common.sh"

usage() {
    cat <<EOF
Usage: $(basename "$0") [-c config_file] [-h] [login|status]

Auto-discover MongoDB configuration from ${CONF_DIR}/*_conf.yaml and
connect via the mongo shell.

Commands:
  login    Start an interactive mongo shell session (default)
  status   Check MongoDB server status and exit

Options:
  -c CONFIG  Specify config file path (default: auto-discover)
  -h         Show this help message

Examples:
  $(basename "$0")
  $(basename "$0") login
  $(basename "$0") status
  $(basename "$0") -c /bk-nodemgr/etc/backend_conf.yaml login
EOF
}

# ensure_port appends default port if host has no port component.
ensure_port() {
    local host="$1"
    case "$host" in
        *:*) echo "$host" ;;
        *)   echo "${host}:${DEFAULT_MONGO_PORT}" ;;
    esac
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

COMMAND="${1:-login}"

# --- Discover or validate config file ---
if [ -z "$CONFIG_FILE" ]; then
    CONFIG_FILE="$(discover_config "mongodb")" || {
        echo "Error: no config file containing 'mongodb:' found in ${CONF_DIR}/" >&2
        echo "Use -c <config_file> to specify manually." >&2
        exit 1
    }
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: config file not found: ${CONFIG_FILE}" >&2
    exit 1
fi

echo "Using config: ${CONFIG_FILE}" >&2

# --- Parse mongodb configuration ---
USERNAME="$(parse_yaml_value "$CONFIG_FILE" "mongodb" "username")"
PASSWORD="$(parse_yaml_value "$CONFIG_FILE" "mongodb" "password")"
DATABASE="$(parse_yaml_value "$CONFIG_FILE" "mongodb" "database")"
AUTH_SOURCE="$(parse_yaml_value "$CONFIG_FILE" "mongodb" "authSource")"
AUTH_MECHANISM="$(parse_yaml_value "$CONFIG_FILE" "mongodb" "authMechanism")"
CA_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "mongodb" "tls" "caFile")"
CERT_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "mongodb" "tls" "certFile")"

# Build hosts string with port defaults
HOSTS=""
_hosts_raw=$(parse_yaml_array "$CONFIG_FILE" "mongodb" "hosts")
if [ -n "$_hosts_raw" ]; then
    while IFS= read -r h; do
        [ -n "$h" ] || continue
        hp=$(ensure_port "$h")
        [ -n "$HOSTS" ] && HOSTS="${HOSTS},"
        HOSTS="${HOSTS}${hp}"
    done <<EOF
$_hosts_raw
EOF
fi

if [ -z "$HOSTS" ]; then
    echo "Error: no MongoDB hosts found in config" >&2
    exit 1
fi

# --- Build connection URI with URL-encoded credentials ---
CONN_URI="mongodb://"

if [ -n "$USERNAME" ] && [ -n "$PASSWORD" ]; then
    ENCODED_USER="$(urlencode "$USERNAME")"
    ENCODED_PASS="$(urlencode "$PASSWORD")"
    CONN_URI="${CONN_URI}${ENCODED_USER}:${ENCODED_PASS}@"
fi

CONN_URI="${CONN_URI}${HOSTS}/${DATABASE:-nodemgr}"

QUERY_PARAMS=""
if [ -n "$AUTH_SOURCE" ]; then
    QUERY_PARAMS="authSource=${AUTH_SOURCE}"
fi
if [ -n "$AUTH_MECHANISM" ]; then
    [ -n "$QUERY_PARAMS" ] && QUERY_PARAMS="${QUERY_PARAMS}&"
    QUERY_PARAMS="${QUERY_PARAMS}authMechanism=${AUTH_MECHANISM}"
fi
if [ -n "$QUERY_PARAMS" ]; then
    CONN_URI="${CONN_URI}?${QUERY_PARAMS}"
fi

# --- Build TLS arguments (space-separated string, unquoted when used) ---
TLS_ARGS=""
if [ -n "$CA_FILE" ] && [ -f "$CA_FILE" ]; then
    TLS_ARGS="--tls --tlsCAFile $CA_FILE"
fi
if [ -n "$CERT_FILE" ] && [ -f "$CERT_FILE" ]; then
    TLS_ARGS="$TLS_ARGS --tlsCertificateKeyFile $CERT_FILE"
fi

# --- Execute command ---
case "$COMMAND" in
    login)
        echo "Connecting to MongoDB: ${HOSTS} (db: ${DATABASE:-nodemgr})" >&2
        # shellcheck disable=SC2086
        exec mongo "$CONN_URI" $TLS_ARGS
        ;;
    status)
        echo "Checking MongoDB status: ${HOSTS}" >&2
        # shellcheck disable=SC2086
        mongo "$CONN_URI" $TLS_ARGS --quiet --eval "
            var status = db.serverStatus();
            print('Server OK: ' + status.ok);
            print('Version: ' + status.version);
            print('Uptime: ' + status.uptime + 's');
            try { var rs = rs.status(); print('ReplicaSet: ' + rs.set + ' (members: ' + rs.members.length + ')'); } catch(e) { print('ReplicaSet: N/A (standalone)'); }
        "
        ;;
    *)
        echo "Error: unknown command '${COMMAND}'" >&2
        echo "Available commands: login, status" >&2
        exit 1
        ;;
esac
