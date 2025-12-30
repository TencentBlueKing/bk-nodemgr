#!/bin/bash
# vim:ft=sh sts=4 ts=4 expandtab
#
# Manual Installation Script
# 
# Usage:
#   /bin/bash -c "$(curl -fsSL http://your-server/api/v3/callback/workflow/node_install/get_manual_script/darwin/OPER_INST_ID?action=ACTION_NAME)"
#
# Note: Placeholders in the script will be replaced by the server with actual values
#   - __BK_NODEMGR_OPERATION_INSTANCE_ID__: Operation instance ID
#   - __BK_NODEMGR_ACTION_NAME_REPORT_DETECT_INFO__: Action name for reporting detect info
#   - __BK_NODEMGR_ACTION_NAME_GET_EXEC_COMMAND__: Action name for getting execute command
#   - __BK_NODEMGR_CALLBACK_ADDRESS__: Callback server address
#   - __BK_NODEMGR_DOWNLOAD_ADDRESS__: Download server address for downloading installer

# Do not use set -e, handle errors manually for better control

# Message output function
msg() {
    local key="$1"
    shift
    
    case "$key" in
        error_detect_os) echo "Error: Failed to detect operating system type" >&2 ;;
        error_detect_cpu) echo "Error: Failed to detect CPU architecture" >&2 ;;
        info_system_detected) echo "Detected system info: OS=$1, CPU=$2, DIR=$3" ;;
        error_connect_server) echo "Error: Unable to connect to server $1" >&2 ;;
        error_report_failed) echo "Error: Failed to report detect info, HTTP status code: $1" >&2 ;;
        error_report_response) echo "Response: $1" >&2 ;;
        success_report_info) echo "Successfully reported detect info" ;;
        error_get_command_failed) echo "Error: Failed to get command, HTTP status code: $1" >&2 ;;
        error_get_command_response) echo "Response: $1" >&2 ;;
        error_get_command_empty) echo "Error: Command is empty or not ready yet" >&2 ;;
        error_get_command_timeout) echo "Error: Timeout waiting for command (max timeout: $1 seconds)" >&2 ;;
        info_download_installer) echo "Downloading installer..." ;;
        error_download_installer_failed) echo "Error: Failed to download installer, HTTP status code: $1" >&2 ;;
        error_download_installer_response) echo "Response: $1" >&2 ;;
        error_create_dir_failed) echo "Error: Failed to create directory $1" >&2 ;;
        success_download_installer) echo "Successfully downloaded installer to $1" ;;
        info_execute_command) echo "Executing command: $1" ;;
        success_command_executed) echo "Command executed successfully" ;;
        warning_command_exit_code) echo "Warning: Command returned non-zero exit code: $1" >&2 ;;
        info_start_install) echo "Starting manual installation process..." ;;
        info_oper_inst_id) echo "Operation instance ID: $1" ;;
        info_callback_server_url) echo "Callback server URL: $1" ;;
        info_download_server_url) echo "Download server URL: $1" ;;
        error_detect_failed) echo "Error: System info detection failed" >&2 ;;
        error_report_failed_main) echo "Error: Failed to report detect info" >&2 ;;
        info_start_get_commands) echo "Starting to get and execute installation commands..." ;;
        info_getting_command) echo "Getting command from server..." ;;
        info_no_command_ready) echo "Info: No command received for $1 consecutive times, installation may be completed or server is not ready yet" >&2 ;;
        info_exit_script) echo "Exiting script" >&2 ;;
        info_waiting_command) echo "Waiting for command... (waited ${1} seconds)" ;;
        error_get_execute_failed) echo "Error: Failed to get or execute command" >&2 ;;
        warning_max_attempts) echo "Warning: Reached maximum attempts ($1), exiting" >&2 ;;
    esac
}

# Get operation instance ID (placeholder in script template will be replaced)
OPER_INST_ID="__BK_NODEMGR_OPERATION_INSTANCE_ID__"

# Get action names (placeholders in script template will be replaced)
ACTION_NAME_REPORT_DETECT_INFO="__BK_NODEMGR_ACTION_NAME_REPORT_DETECT_INFO__"
ACTION_NAME_GET_EXEC_COMMAND="__BK_NODEMGR_ACTION_NAME_GET_EXEC_COMMAND__"

# Get callback server address (placeholder in script template will be replaced)
CALLBACK_SERVER_URL="__BK_NODEMGR_CALLBACK_ADDRESS__"

# Remove trailing slash from URL
CALLBACK_SERVER_URL="${CALLBACK_SERVER_URL%/}"

# Get download server address (placeholder in script template will be replaced)
DOWNLOAD_SERVER_URL="__BK_NODEMGR_DOWNLOAD_ADDRESS__"

# Remove trailing slash from URL
DOWNLOAD_SERVER_URL="${DOWNLOAD_SERVER_URL%/}"

# Detect system information
detect_system_info() {
    # Detect OS type
    local os_type_raw
    os_type_raw=$(uname -s 2>/dev/null || echo "")
    if [ -z "$os_type_raw" ]; then
        msg error_detect_os
        return 1
    fi
    OS_TYPE=$(echo "$os_type_raw" | tr '[:upper:]' '[:lower:]' | tr -d '\n\r')
    
    # Detect CPU architecture
    local cpu_arch_raw
    cpu_arch_raw=$(uname -m 2>/dev/null || echo "")
    if [ -z "$cpu_arch_raw" ]; then
        msg error_detect_cpu
        return 1
    fi
    CPU_ARCH=$(echo "$cpu_arch_raw" | tr '[:upper:]' '[:lower:]' | tr -d '\n\r')
    
    # Detect current working directory
    CONNECTION_DIR=$(pwd 2>/dev/null || echo "")
    if [ -z "$CONNECTION_DIR" ]; then
        CONNECTION_DIR="/tmp"
    fi
    
    msg info_system_detected "$OS_TYPE" "$CPU_ARCH" "$CONNECTION_DIR"
}

# Report detect information
report_detect_info() {
    local action_name="${ACTION_NAME_REPORT_DETECT_INFO}"
    local url="${CALLBACK_SERVER_URL}/api/v3/callback/workflow/node_install/report_detect_info"
    
    # Build JSON request body (use high-compatibility method, avoid using jq)
    local json_body
    json_body=$(cat <<EOF
{
  "oper_inst_id": "${OPER_INST_ID}",
  "action_name": "${action_name}",
  "os_type": "${OS_TYPE}",
  "cpu_arch": "${CPU_ARCH}",
  "connection_dir": "${CONNECTION_DIR}",
  "err_msg": ""
}
EOF
)
    
    # Send POST request
    local http_code
    local response
    response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$json_body" \
        "$url" 2>/dev/null || echo "")
    
    if [ -z "$response" ]; then
        msg error_connect_server "$url"
        return 1
    fi
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n 1)
    
    # Check HTTP status code
    if [ "$http_code" != "200" ] && [ "$http_code" != "201" ]; then
        msg error_report_failed "$http_code"
        msg error_report_response "$(echo "$response" | head -n -1)"
        return 1
    fi
    
    msg success_report_info
    return 0
}


# Download installer to specified path
download_installer() {
    local installer_path="$1"
    local url="${DOWNLOAD_SERVER_URL}/api/v3/download/installer"
    
    # Extract directory from installer path
    local installer_dir
    installer_dir=$(dirname "$installer_path")
    
    # Create directory if it doesn't exist
    if [ ! -d "$installer_dir" ]; then
        mkdir -p "$installer_dir" 2>/dev/null || {
            msg error_create_dir_failed "$installer_dir"
            return 1
        }
    fi
    
    # Build JSON request body
    local json_body
    json_body=$(cat <<EOF
{
  "os_type": "${OS_TYPE}",
  "cpu_arch": "${CPU_ARCH}"
}
EOF
)
    
    # Download installer file
    local http_code
    http_code=$(curl -s -w "%{http_code}" -o "$installer_path" -X POST \
        -H "Content-Type: application/json" \
        -d "$json_body" \
        "$url" 2>/dev/null || echo "000")
    
    # Check HTTP status code
    if [ "$http_code" != "200" ]; then
        msg error_download_installer_failed "$http_code"
        if [ -f "$installer_path" ]; then
            local error_content
            error_content=$(cat "$installer_path" 2>/dev/null || echo "")
            msg error_download_installer_response "$error_content"
            rm -f "$installer_path"
        fi
        return 1
    fi
    
    # Make installer executable
    chmod +x "$installer_path" 2>/dev/null || true
    
    msg success_download_installer "$installer_path"
    return 0
}

# Get command from server
get_command() {
    local action_name="${ACTION_NAME_GET_EXEC_COMMAND}"
    local url="${CALLBACK_SERVER_URL}/api/v3/callback/workflow/node_install/get_manual_install_exec_command"
    
    # Build JSON request body
    local json_body
    json_body=$(cat <<EOF
{
  "oper_inst_id": "${OPER_INST_ID}",
  "action_name": "${action_name}"
}
EOF
)
    
    # Send POST request to get command
    local http_code
    local response
    response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$json_body" \
        "$url" 2>/dev/null || echo "")
    
    if [ -z "$response" ]; then
        msg error_connect_server "$url"
        return 1
    fi
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n 1)
    
    # Check HTTP status code
    if [ "$http_code" != "200" ]; then
        msg error_get_command_failed "$http_code"
        msg error_get_command_response "$(echo "$response" | head -n -1)"
        return 1
    fi
    
    # Extract installer_path and command (remove last line HTTP status code)
    # Note: The API returns plain text, first line is installer_path, second line is command
    local response_body
    response_body=$(echo "$response" | head -n -1)
    
    # Extract first line (installer_path) and second line (command)
    local installer_path
    local command
    installer_path=$(echo "$response_body" | sed -n '1p' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    command=$(echo "$response_body" | sed -n '2p' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    # If installer_path or command is empty, it means not ready yet, return 2 to indicate need to continue waiting
    if [ -z "$installer_path" ] || [ -z "$command" ] || [ "$installer_path" = "null" ] || [ "$command" = "null" ] || [ "$installer_path" = '""' ] || [ "$command" = '""' ]; then
        return 2
    fi
    
    # Output installer_path and command (first line installer_path, second line command)
    echo "$installer_path"
    echo "$command"
    return 0
}

# Main function
main() {
    msg info_start_install
    msg info_oper_inst_id "$OPER_INST_ID"
    msg info_callback_server_url "$CALLBACK_SERVER_URL"
    msg info_download_server_url "$DOWNLOAD_SERVER_URL"
    
    # Detect system information
    if ! detect_system_info; then
        msg error_detect_failed
        exit 1
    fi
    
    # Report detect information
    if ! report_detect_info; then
        msg error_report_failed_main
        exit 1
    fi
    
    # Periodically get command until timeout or command is ready
    msg info_start_get_commands
    msg info_getting_command
    
    local max_timeout=300  # Maximum timeout in seconds
    local sleep_interval=1  # Wait 1 second between attempts
    local start_time
    start_time=$(date +%s)
    local command=""
    
    while true; do
        # Check timeout
        local current_time
        current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        
        if [ $elapsed -ge $max_timeout ]; then
            msg error_get_command_timeout "$max_timeout"
            exit 1
        fi
        
        # Try to get command
        local result
        result=$(get_command)
        local get_result=$?
        
        if [ $get_result -eq 0 ]; then
            # Command is ready - extract installer_path and command from result
            # get_command returns two lines: first line is installer_path, second line is command
            installer_path=$(echo "$result" | sed -n '1p')
            command=$(echo "$result" | sed -n '2p')
            break
        elif [ $get_result -eq 2 ]; then
            # Command not ready yet, continue waiting
            # Output prompt every 10 seconds
            if [ $((elapsed % 10)) -eq 0 ] && [ $elapsed -gt 0 ]; then
                msg info_waiting_command "$elapsed"
            fi
            sleep $sleep_interval
        else
            # Error occurred
            msg error_get_execute_failed
            exit 1
        fi
    done
    
    # installer_path and command are already extracted from get_command result
    if [ -z "$installer_path" ]; then
        msg error_get_execute_failed
        exit 1
    fi
    if [ -z "$command" ]; then
        msg error_get_execute_failed
        exit 1
    fi
    
    # Download installer to the path specified in command
    msg info_download_installer
    if ! download_installer "$installer_path"; then
        exit 1
    fi
    
    # Execute command once
    msg info_execute_command "$command"
    eval "$command"
    local exec_result=$?
    
    # Exit based on execution result
    if [ $exec_result -eq 0 ]; then
        msg success_command_executed
        exit 0
    else
        msg warning_command_exit_code "$exec_result"
        exit $exec_result
    fi
}

# Execute main function
main "$@"

