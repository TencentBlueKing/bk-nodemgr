#!/bin/sh

# trap signals
trap "exit" SIGTERM

# monitor modules
# Using a list of modules with their subcommands
MODULES="application:webserver backend: file:"

# init installer tools.
cd /bk-nodemgr/workspace
mkdir -p temp cache local installer
cp /bk-nodemgr/file/tools/installer_* installer/

while true
do
    for module_setting in $MODULES
    do
        # Split module and subcommand
        module=$(echo "$module_setting" | cut -d: -f1)
        subcommand=$(echo "$module_setting" | cut -d: -f2)
        
        # Count running processes
        num=$(ps -ef | grep "bk-nodemgr-$module" | grep -v grep | wc -l)
        
        if [ "$num" -eq 0 ]; then
            # Start the module if not running
            cd /bk-nodemgr/bin && ./bk-nodemgr-"$module" $subcommand -f "/bk-nodemgr/etc/bk-nodemgr-${module}.yml" &
        fi
    done

    sleep 1
done