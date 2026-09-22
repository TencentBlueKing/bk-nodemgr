#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELMFILE_FILE="$ROOT_DIR/helmfile.yaml.gotmpl"
LOCAL_HELMFILE="$ROOT_DIR/bin/helmfile"
LOCAL_BIN="$ROOT_DIR/bin"
LOCAL_HELM_PLUGINS="$ROOT_DIR/bin/helm-plugins"

cd "$ROOT_DIR"

export PATH="$LOCAL_BIN:$PATH"

if [ -d "$LOCAL_HELM_PLUGINS" ]; then
	export HELM_PLUGINS="$LOCAL_HELM_PLUGINS"
fi

if [ -x "$LOCAL_HELMFILE" ]; then
	HELMFILE_BIN="$LOCAL_HELMFILE"
else
	HELMFILE_BIN="helmfile"
fi

function usage() {
	echo "Usage: $0 {diff|apply|sync|destroy|render-values} [environment] [deploy args...] [helmfile args...]"
	echo "Example: $0 apply example"
	echo "Example: $0 render-values example"
	echo "Example: $0 render-values example --modules backend"
	echo "Example: $0 diff example --skip-deps --debug"
	echo "Example: $0 diff --skip-deps --debug"
	echo "Example: $0 diff example --skip-module monitoring --skip-deps"
	echo "Example: $0 diff example --modules file,backend --skip-deps"
	exit 1
}

function valid_modules() {
	echo "deps file backend application monitoring"
}

function is_valid_module() {
	local module=$1
	case "$module" in
	deps | file | backend | application | monitoring)
		return 0
		;;
	*)
		return 1
		;;
	esac
}

function ensure_valid_module() {
	local module=$1
	if ! is_valid_module "$module"; then
		echo "Invalid module: $module"
		echo "Valid modules: $(valid_modules)"
		exit 1
	fi
}

function contains_module() {
	local module=$1
	shift
	local item
	for item in "$@"; do
		if [ "$item" = "$module" ]; then
			return 0
		fi
	done
	return 1
}

function should_run_module() {
	local module=$1
	if [ "${#ONLY_MODULES[@]}" -gt 0 ] && ! contains_module "$module" "${ONLY_MODULES[@]}"; then
		return 1
	fi
	if contains_module "$module" "${SKIP_MODULES[@]}"; then
		return 1
	fi
	return 0
}

function run_module() {
	local action=$1
	local environment=$2
	local module=$3
	shift 3

	"$HELMFILE_BIN" -f "$HELMFILE_FILE" -e "$environment" -l "module=$module" "$action" "$@"
}

function selected_modules() {
	local module
	for module in "$@"; do
		if should_run_module "$module"; then
			echo "$module"
		fi
	done
}

function ensure_same_release_name() {
	local environment=$1
	shift
	local release_name=""
	local module
	local module_release_name
	for module in "$@"; do
		module_release_name=$("$HELMFILE_BIN" -f "$HELMFILE_FILE" -e "$environment" -l "module=$module" list --skip-charts | awk 'NR == 2 {print $1}')
		if [ -z "$module_release_name" ]; then
			echo "Failed to resolve release name for module: $module"
			exit 1
		fi
		if [ -z "$release_name" ]; then
			release_name=$module_release_name
		elif [ "$release_name" != "$module_release_name" ]; then
			echo "render-values requires all selected modules to use the same release name"
			echo "Expected release name: $release_name"
			echo "Module $module uses release name: $module_release_name"
			exit 1
		fi
	done
}

function has_helmfile_arg() {
	local expected=$1
	local arg
	for arg in "${HELMFILE_ARGS[@]}"; do
		if [ "$arg" = "$expected" ]; then
			return 0
		fi
	done
	return 1
}

function render_values() {
	local environment=$1
	shift
	local modules=("$@")
	local output_template="rendered-values/$environment/values/{{ .Release.Namespace }}.yaml"
	mkdir -p "$ROOT_DIR/rendered-values/$environment/values"
	ensure_same_release_name "$environment" "${modules[@]}"
	if ! has_helmfile_arg "--skip-deps"; then
		HELMFILE_ARGS+=("--skip-deps")
	fi
	HELMFILE_ARGS+=("--output-file-template" "$output_template")
	run_modules "write-values" "$environment" "${modules[@]}"
}

if [ "$#" -lt 1 ]; then
	usage
fi

ACTION=$1
shift

ENVIRONMENT=example
if [ "$#" -gt 0 ] && [[ "$1" != -* ]]; then
	ENVIRONMENT=$1
	shift
fi

ONLY_MODULES=()
SKIP_MODULES=()
HELMFILE_ARGS=()

while [ "$#" -gt 0 ]; do
	case "$1" in
	--modules)
		if [ "$#" -lt 2 ]; then
			echo "Missing value for --modules"
			usage
		fi
		IFS=',' read -ra MODULES <<<"$2"
		for module in "${MODULES[@]}"; do
			ensure_valid_module "$module"
			ONLY_MODULES+=("$module")
		done
		shift 2
		;;
	--skip-module)
		if [ "$#" -lt 2 ]; then
			echo "Missing value for --skip-module"
			usage
		fi
		ensure_valid_module "$2"
		SKIP_MODULES+=("$2")
		shift 2
		;;
	*)
		HELMFILE_ARGS+=("$1")
		shift
		;;
	esac
done

function run_modules() {
	local action=$1
	local environment=$2
	shift 2
	local module
	for module in "$@"; do
		if should_run_module "$module"; then
			run_module "$action" "$environment" "$module" "${HELMFILE_ARGS[@]}"
		fi
	done
}

case "$ACTION" in
diff | apply | sync)
	run_modules "$ACTION" "$ENVIRONMENT" deps file backend application monitoring
	;;
render-values)
	mapfile -t SELECTED_MODULES < <(selected_modules deps file backend application monitoring)
	if [ "${#SELECTED_MODULES[@]}" -eq 0 ]; then
		echo "No modules selected"
		exit 1
	fi
	render_values "$ENVIRONMENT" "${SELECTED_MODULES[@]}"
	;;
destroy)
	run_modules "$ACTION" "$ENVIRONMENT" monitoring application backend file deps
	;;
*)
	echo "Invalid operation: $ACTION"
	usage
	;;
esac
