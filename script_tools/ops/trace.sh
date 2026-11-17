#!/bin/sh
# Ops helper: find local logs by trace ID without an exporter.
# Usage: trace.sh <traceid>

set -eu

if [ "$#" -ne 1 ]; then
	echo "Usage: $(basename "$0") <traceid>" >&2
	exit 2
fi

trace_id="$1"
case "$trace_id" in
'' | *[!0123456789abcdef]*)
	echo "Error: traceid must be 32 lowercase hexadecimal characters" >&2
	exit 2
	;;
esac
if [ "${#trace_id}" -ne 32 ]; then
	echo "Error: traceid must be 32 lowercase hexadecimal characters" >&2
	exit 2
fi

log_dir="/bk-nodemgr/log"
if [ ! -d "$log_dir" ] || [ ! -r "$log_dir" ] || [ ! -x "$log_dir" ]; then
	echo "Error: log directory missing or unreadable: $log_dir" >&2
	exit 2
fi

status=1
for file in "$log_dir"/*; do
	# Current-log symlinks point to files already included in this scan.
	[ ! -L "$file" ] && [ -f "$file" ] || continue
	if grep -FnH -e "trace-id($trace_id)" "$file"; then
		if [ "$status" -ne 2 ]; then
			status=0
		fi
	else
		result=$?
		if [ "$result" -gt 1 ]; then
			status=2
		fi
	fi
done

case "$status" in
1) echo "No matching logs in $log_dir for traceid $trace_id" >&2 ;;
2) echo "Error: log search incomplete; some files could not be read" >&2 ;;
esac
exit "$status"
