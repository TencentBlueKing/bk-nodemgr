#!/bin/sh
# Ops helper: auto-discover Redis config and connect via redis-cli.
# Usage: redis.sh [-c config] [-h] [login|status]

set -eu

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONF_DIR="${CONF_DIR:-/bk-nodemgr/etc}"
CONFIG_FILE=""

. "${SCRIPT_DIR}/_common.sh"

usage() {
    cat <<EOF
Usage: $(basename "$0") [-c config_file] [-h] [login|status]

Auto-discover Redis configuration from ${CONF_DIR}/*_conf.yaml and
connect via redis-cli.

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

# split_host_port splits "host:port" into HOST and PORT variables.
split_host_port() {
    local addr="$1"
    case "$addr" in
        *:*)
            HOST="${addr%:*}"
            PORT="${addr##*:}"
            ;;
        *)
            HOST="$addr"
            PORT="6379"
            ;;
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
    CONFIG_FILE="$(discover_config "redis")" || {
        echo "Error: no config file containing 'redis:' found in ${CONF_DIR}/" >&2
        echo "Note: only backend_conf.yaml typically includes Redis configuration." >&2
        echo "Use -c <config_file> to specify manually." >&2
        exit 1
    }
fi

if [ ! -f "$CONFIG_FILE" ]; then
    echo "Error: config file not found: ${CONFIG_FILE}" >&2
    exit 1
fi

echo "Using config: ${CONFIG_FILE}" >&2

# --- Parse redis configuration ---
REDIS_TYPE="$(parse_yaml_value "$CONFIG_FILE" "redis" "type")"
REDIS_TYPE="${REDIS_TYPE:-standalone}"

ADDRS=$(parse_yaml_array "$CONFIG_FILE" "redis" "addrs")

PASSWORD="$(parse_yaml_value "$CONFIG_FILE" "redis" "password")"
DB="$(parse_yaml_value "$CONFIG_FILE" "redis" "db")"
MASTER_NAME="$(parse_yaml_value "$CONFIG_FILE" "redis" "masterName")"
CA_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "redis" "tls" "caFile")"
CERT_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "redis" "tls" "certFile")"
KEY_FILE="$(parse_yaml_nested_value "$CONFIG_FILE" "redis" "tls" "keyFile")"

if [ -z "$ADDRS" ]; then
    echo "Error: no Redis addresses found in config" >&2
    exit 1
fi

FIRST_ADDR=$(echo "$ADDRS" | head -1)

# --- Build common TLS arguments (space-separated string, unquoted when used) ---
TLS_ARGS=""
if [ -n "$CA_FILE" ] && [ -f "$CA_FILE" ]; then
    TLS_ARGS="--tls --cacert $CA_FILE"
fi
if [ -n "$CERT_FILE" ] && [ -f "$CERT_FILE" ]; then
    TLS_ARGS="$TLS_ARGS --cert $CERT_FILE"
fi
if [ -n "$KEY_FILE" ] && [ -f "$KEY_FILE" ]; then
    TLS_ARGS="$TLS_ARGS --key $KEY_FILE"
fi

# --- Set auth via environment variable (avoids password in ps output) ---
if [ -n "$PASSWORD" ]; then
    export REDISCLI_AUTH="$PASSWORD"
fi

# --- Helper: resolve sentinel master ---
resolve_sentinel_master() {
    local sentinel_addr="$1"
    local master="$2"
    split_host_port "$sentinel_addr"
    local result
    # shellcheck disable=SC2086
    result=$(redis-cli -h "$HOST" -p "$PORT" $TLS_ARGS \
        SENTINEL get-master-addr-by-name "$master" 2>/dev/null)
    if [ -z "$result" ]; then
        echo "Error: could not resolve master '${master}' from sentinel ${sentinel_addr}" >&2
        return 1
    fi
    local m_host m_port
    m_host=$(echo "$result" | head -1)
    m_port=$(echo "$result" | tail -1)
    echo "${m_host}:${m_port}"
}

# --- Execute command ---
# shellcheck disable=SC2086
case "$COMMAND" in
    login)
        case "$REDIS_TYPE" in
            standalone)
                split_host_port "$FIRST_ADDR"
                echo "Connecting to Redis standalone: ${HOST}:${PORT} (db: ${DB:-0})" >&2
                exec redis-cli -h "$HOST" -p "$PORT" -n "${DB:-0}" $TLS_ARGS
                ;;
            cluster)
                split_host_port "$FIRST_ADDR"
                echo "Connecting to Redis cluster: ${HOST}:${PORT}" >&2
                exec redis-cli -c -h "$HOST" -p "$PORT" $TLS_ARGS
                ;;
            sentinel)
                if [ -z "$MASTER_NAME" ]; then
                    echo "Error: sentinel mode requires masterName in config" >&2
                    exit 1
                fi
                echo "Resolving master '${MASTER_NAME}' via sentinel: ${FIRST_ADDR}" >&2
                MASTER_ADDR="$(resolve_sentinel_master "$FIRST_ADDR" "$MASTER_NAME")" || exit 1
                split_host_port "$MASTER_ADDR"
                echo "Connecting to Redis master: ${HOST}:${PORT} (db: ${DB:-0})" >&2
                exec redis-cli -h "$HOST" -p "$PORT" -n "${DB:-0}" $TLS_ARGS
                ;;
            *)
                echo "Error: unsupported Redis type '${REDIS_TYPE}'" >&2
                echo "Supported types: standalone, cluster, sentinel" >&2
                exit 1
                ;;
        esac
        ;;
    status)
        case "$REDIS_TYPE" in
            standalone)
                split_host_port "$FIRST_ADDR"
                echo "Checking Redis standalone: ${HOST}:${PORT}" >&2
                redis-cli -h "$HOST" -p "$PORT" -n "${DB:-0}" $TLS_ARGS PING
                redis-cli -h "$HOST" -p "$PORT" -n "${DB:-0}" $TLS_ARGS INFO server | grep -E "^(redis_version|uptime_in_seconds|connected_clients):"
                ;;
            cluster)
                split_host_port "$FIRST_ADDR"
                echo "Checking Redis cluster: ${HOST}:${PORT}" >&2
                redis-cli -c -h "$HOST" -p "$PORT" $TLS_ARGS CLUSTER INFO
                ;;
            sentinel)
                if [ -z "$MASTER_NAME" ]; then
                    echo "Error: sentinel mode requires masterName in config" >&2
                    exit 1
                fi
                echo "Checking sentinel master '${MASTER_NAME}': ${FIRST_ADDR}" >&2
                MASTER_ADDR="$(resolve_sentinel_master "$FIRST_ADDR" "$MASTER_NAME")" || exit 1
                split_host_port "$MASTER_ADDR"
                echo "Master resolved: ${HOST}:${PORT}" >&2
                redis-cli -h "$HOST" -p "$PORT" -n "${DB:-0}" $TLS_ARGS PING
                ;;
            *)
                echo "Error: unsupported Redis type '${REDIS_TYPE}'" >&2
                exit 1
                ;;
        esac
        ;;
    *)
        echo "Error: unknown command '${COMMAND}'" >&2
        echo "Available commands: login, status" >&2
        exit 1
        ;;
esac
