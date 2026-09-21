#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELMFILE_FILE="$ROOT_DIR/helmfile.yaml"
LOCAL_HELMFILE="$ROOT_DIR/bin/helmfile"

cd "$ROOT_DIR"

if [ -x "$LOCAL_HELMFILE" ]; then
	HELMFILE_BIN="$LOCAL_HELMFILE"
else
	HELMFILE_BIN="helmfile"
fi

function usage() {
	echo "Usage: $0 {diff|apply|sync|destroy} [environment]"
	echo "Example: $0 apply example"
	exit 1
}

function run_module() {
	local action=$1
	local environment=$2
	local module=$3

	"$HELMFILE_BIN" -f "$HELMFILE_FILE" -e "$environment" -l "module=$module" "$action"
}

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
	usage
fi

ACTION=$1
ENVIRONMENT=${2:-example}

case "$ACTION" in
diff | apply | sync)
	run_module "$ACTION" "$ENVIRONMENT" file
	run_module "$ACTION" "$ENVIRONMENT" backend
	run_module "$ACTION" "$ENVIRONMENT" application
	;;
destroy)
	run_module "$ACTION" "$ENVIRONMENT" application
	run_module "$ACTION" "$ENVIRONMENT" backend
	run_module "$ACTION" "$ENVIRONMENT" file
	;;
*)
	echo "Invalid operation: $ACTION"
	usage
	;;
esac
