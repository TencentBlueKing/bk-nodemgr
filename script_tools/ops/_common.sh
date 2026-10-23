#!/bin/sh
# Shared helper functions for ops scripts.
# Source this file: . "${SCRIPT_DIR}/_common.sh"

# find_adminclient locates the bk-nodemgr-adminclient binary.
# Lookup order: ADMINCLIENT_BIN env, /bk-nodemgr/bin, PATH.
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
