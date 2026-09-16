#!/bin/sh
# Ops helper: sync unassigned agent hosts to recommended network units via backend admin API.
# Usage: sync_unassigned_network_unit.sh [-c config] [--login-name user] (--bk-biz-id ids | --all)

set -eu

CONF_DIR="${CONF_DIR:-/bk-nodemgr/etc}"
CONFIG_FILE=""
LOGIN_NAME="admin"
BK_BIZ_ID=""
ALL="false"

usage() {
	cat <<EOF
Usage: $(basename "$0") [-c config_file] [--login-name user] (--bk-biz-id ids | --all)

Sync agent hosts whose network unit is unassigned to recommended network units.
The command prints the backend admin API result JSON to stdout.

Options:
  -c, --config CONFIG  Specify backend config file path (default: backend config in CONF_DIR)
  --login-name USER    Login name for backend admin authentication (default: admin)
  --bk-biz-id IDS      Comma-separated bk biz ids to sync, for example 2,3
  --all                Explicitly sync all businesses
  -h, --help           Show this help message

Examples:
  $(basename "$0") --bk-biz-id 2,3
  $(basename "$0") --all
  $(basename "$0") -c /bk-nodemgr/etc/backend_conf.yaml --login-name admin --bk-biz-id 2
EOF
}

discover_backend_config() {
	for candidate in "${CONF_DIR}/backend_conf.yaml" "${CONF_DIR}/bk-nodemgr-backend.yml"; do
		if [ -f "$candidate" ]; then
			echo "$candidate"
			return 0
		fi
	done

	return 1
}

find_adminclient() {
	if [ -n "${ADMINCLIENT_BIN:-}" ]; then
		echo "$ADMINCLIENT_BIN"
		return 0
	fi

	if [ -x "/bk-nodemgr/bin/bk-nodemgr-adminclient" ]; then
		echo "/bk-nodemgr/bin/bk-nodemgr-adminclient"
		return 0
	fi

	command -v bk-nodemgr-adminclient 2>/dev/null || return 1
}

# --- Parse script options ---
while [ $# -gt 0 ]; do
	case "$1" in
	-c | --config)
		shift
		CONFIG_FILE="${1:-}"
		[ -z "$CONFIG_FILE" ] && {
			echo "Error: -c/--config requires a config file path" >&2
			exit 1
		}
		shift
		;;
	--login-name)
		shift
		LOGIN_NAME="${1:-}"
		[ -z "$LOGIN_NAME" ] && {
			echo "Error: --login-name requires a value" >&2
			exit 1
		}
		shift
		;;
	--bk-biz-id)
		shift
		BK_BIZ_ID="${1:-}"
		[ -z "$BK_BIZ_ID" ] && {
			echo "Error: --bk-biz-id requires comma-separated ids" >&2
			exit 1
		}
		shift
		;;
	--all)
		ALL="true"
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Error: unknown option '$1'" >&2
		echo "" >&2
		usage >&2
		exit 1
		;;
	esac
done

if [ "$ALL" = "true" ] && [ -n "$BK_BIZ_ID" ]; then
	echo "Error: --all and --bk-biz-id cannot be used together" >&2
	exit 1
fi

if [ "$ALL" != "true" ] && [ -z "$BK_BIZ_ID" ]; then
	echo "Error: one of --all or --bk-biz-id is required" >&2
	echo "" >&2
	usage >&2
	exit 1
fi

# --- Discover or validate config file ---
if [ -z "$CONFIG_FILE" ]; then
	CONFIG_FILE="$(discover_backend_config)" || {
		echo "Error: backend config not found in ${CONF_DIR}/" >&2
		echo "Expected backend_conf.yaml or bk-nodemgr-backend.yml; use -c <config_file> to specify manually." >&2
		exit 1
	}
fi

if [ ! -f "$CONFIG_FILE" ]; then
	echo "Error: config file not found: ${CONFIG_FILE}" >&2
	exit 1
fi

ADMINCLIENT="$(find_adminclient)" || {
	echo "Error: bk-nodemgr-adminclient not found. Set ADMINCLIENT_BIN or add it to PATH." >&2
	exit 1
}

if [ ! -x "$ADMINCLIENT" ]; then
	echo "Error: adminclient is not executable: ${ADMINCLIENT}" >&2
	exit 1
fi

echo "Using config: ${CONFIG_FILE}" >&2

set -- "$ADMINCLIENT" backend --file "$CONFIG_FILE" --login-name "$LOGIN_NAME" \
	node agent sync-unassigned-network-unit

if [ "$ALL" = "true" ]; then
	set -- "$@" --all
else
	set -- "$@" --bk-biz-id "$BK_BIZ_ID"
fi

exec "$@"
