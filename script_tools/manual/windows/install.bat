@echo off
:: Manual Installation Script for Windows
:: 
:: Usage:
::   powershell -ExecutionPolicy Bypass -Command "& {Invoke-WebRequest -Uri 'http://your-server/api/v3/backend/callback/workflow/node_install/get_manual_script/windows/OPER_INST_ID?action=ACTION_NAME' -OutFile install.bat; .\install.bat}"
::
:: Note: Placeholders in the script will be replaced by the server with actual values
::   - __BK_NODEMGR_OPERATION_INSTANCE_ID__: Operation instance ID
::   - __BK_NODEMGR_ACTION_NAME_REPORT_DETECT_INFO__: Action name for reporting detect info
::   - __BK_NODEMGR_ACTION_NAME_GET_EXEC_COMMAND__: Action name for getting execute command
::   - __BK_NODEMGR_CALLBACK_ADDRESS__: Callback server address
::   - __BK_NODEMGR_DOWNLOAD_ADDRESS__: Download server address for downloading installer

setlocal enabledelayedexpansion

:: Get operation instance ID (placeholder in script template will be replaced)
set "OPER_INST_ID=__BK_NODEMGR_OPERATION_INSTANCE_ID__"

:: Get action names (placeholders in script template will be replaced)
set "ACTION_NAME_REPORT_DETECT_INFO=__BK_NODEMGR_ACTION_NAME_REPORT_DETECT_INFO__"
set "ACTION_NAME_GET_EXEC_COMMAND=__BK_NODEMGR_ACTION_NAME_GET_EXEC_COMMAND__"

:: Get callback server address (placeholder in script template will be replaced)
set "CALLBACK_SERVER_URL=__BK_NODEMGR_CALLBACK_ADDRESS__"

:: Remove trailing slash from URL
if "!CALLBACK_SERVER_URL:~-1!"=="\" set "CALLBACK_SERVER_URL=!CALLBACK_SERVER_URL:~0,-1!"
if "!CALLBACK_SERVER_URL:~-1!"=="/" set "CALLBACK_SERVER_URL=!CALLBACK_SERVER_URL:~0,-1!"

:: Get download server address (placeholder in script template will be replaced)
set "DOWNLOAD_SERVER_URL=__BK_NODEMGR_DOWNLOAD_ADDRESS__"

:: Remove trailing slash from URL
if "!DOWNLOAD_SERVER_URL:~-1!"=="\" set "DOWNLOAD_SERVER_URL=!DOWNLOAD_SERVER_URL:~0,-1!"
if "!DOWNLOAD_SERVER_URL:~-1!"=="/" set "DOWNLOAD_SERVER_URL=!DOWNLOAD_SERVER_URL:~0,-1!"

:: Execute main function
call :main

:msg
    set "msg_key=%~1"
    set "msg_arg1=%~2"
    set "msg_arg2=%~3"
    set "msg_arg3=%~4"
    if "!msg_key!"=="error_detect_os" (
        echo Error: Failed to detect operating system type >&2
    )
    if "!msg_key!"=="error_detect_cpu" (
        echo Error: Failed to detect CPU architecture >&2
    )
    if "!msg_key!"=="info_system_detected" (
        echo Detected system info: OS=!msg_arg1!, CPU=!msg_arg2!, DIR=!msg_arg3!
    )
    if "!msg_key!"=="error_connect_server" (
        echo Error: Unable to connect to server !msg_arg1!, HTTP status code: !msg_arg2! >&2
    )
    if "!msg_key!"=="error_report_failed" (
        echo Error: Failed to report detect info, HTTP status code: !msg_arg1! >&2
    )
    if "!msg_key!"=="error_report_response" (
        echo Response: !msg_arg1! >&2
    )
    if "!msg_key!"=="success_report_info" (
        echo Successfully reported detect info
    )
    if "!msg_key!"=="error_get_command_failed" (
        echo Error: Failed to get command, HTTP status code: !msg_arg1! >&2
    )
    if "!msg_key!"=="error_get_command_response" (
        echo Response: !msg_arg1! >&2
    )
    if "!msg_key!"=="error_get_command_empty" (
        echo Error: Command is empty or not ready yet >&2
    )
    if "!msg_key!"=="error_get_command_timeout" (
        echo Error: Timeout waiting for command (max timeout: !msg_arg1! seconds) >&2
    )
    if "!msg_key!"=="info_download_installer" (
        echo Downloading installer...
    )
    if "!msg_key!"=="error_download_installer_failed" (
        echo Error: Failed to download installer, HTTP status code: !msg_arg1! >&2
    )
    if "!msg_key!"=="error_download_installer_response" (
        echo Response: !msg_arg1! >&2
    )
    if "!msg_key!"=="error_create_dir_failed" (
        echo Error: Failed to create directory !msg_arg1! >&2
    )
    if "!msg_key!"=="success_download_installer" (
        echo Successfully downloaded installer to !msg_arg1!
    )
    if "!msg_key!"=="info_execute_command" (
        echo Executing command: !msg_arg1!
    )
    if "!msg_key!"=="success_command_executed" (
        echo Command executed successfully
    )
    if "!msg_key!"=="warning_command_exit_code" (
        echo "Warning: Command returned non-zero exit code: !msg_arg1!" >&2
    )
    if "!msg_key!"=="info_start_install" (
        echo Starting manual installation process...
    )
    if "!msg_key!"=="info_oper_inst_id" (
        echo Operation instance ID: !msg_arg1!
    )
    if "!msg_key!"=="info_callback_server_url" (
        echo Callback server URL: !msg_arg1!
    )
    if "!msg_key!"=="info_download_server_url" (
        echo Download server URL: !msg_arg1!
    )
    if "!msg_key!"=="error_detect_failed" (
        echo Error: System info detection failed >&2
    )
    if "!msg_key!"=="error_report_failed_main" (
        echo Error: Failed to report detect info >&2
    )
    if "!msg_key!"=="info_start_get_commands" (
        echo Starting to get and execute installation commands...
    )
    if "!msg_key!"=="info_getting_command" (
        echo Getting command from server...
    )
    if "!msg_key!"=="info_waiting_command" (
        echo Waiting for command... (waited !msg_arg1! seconds)
    )
    if "!msg_key!"=="error_get_execute_failed" (
        echo Error: Failed to get or execute command >&2
    )
    exit /b 0

:: Detect system information
:detect_system_info
    :: Detect OS type - Windows is always "windows"
    set "OS_TYPE=windows"

    :: Detect CPU architecture using PowerShell
    for /f "delims=" %%i in ('powershell -Command "Write-Output $env:PROCESSOR_ARCHITECTURE"') do set "CPU_ARCH_RAW=%%i"
    if "!CPU_ARCH_RAW!"=="" (
        :: Fallback: try wmic
        for /f "tokens=2 delims==" %%i in ('wmic os get osarchitecture /value 2^>nul') do (
            set "CPU_ARCH_RAW=%%i"
        )
    )

    :: Normalize CPU architecture
    set "CPU_ARCH=!CPU_ARCH_RAW!"
    if /i "!CPU_ARCH!"=="AMD64" set "CPU_ARCH=x86_64"
    if /i "!CPU_ARCH!"=="x64" set "CPU_ARCH=x86_64"
    if /i "!CPU_ARCH!"=="x86" set "CPU_ARCH=i386"
    if /i "!CPU_ARCH!"=="ARM64" set "CPU_ARCH=aarch64"

    if "!CPU_ARCH!"=="" (
        call :msg error_detect_cpu
        exit /b 1
    )

    :: Detect current working directory
    set "CONNECTION_DIR=%CD%"
    if "!CONNECTION_DIR!"=="" set "CONNECTION_DIR=%TEMP%"

    call :msg info_system_detected "!OS_TYPE!" "!CPU_ARCH!" "!CONNECTION_DIR!"
    exit /b 0

:: Report detect information
:report_detect_info
    set "url=!CALLBACK_SERVER_URL!/callback/workflow/node_install/report_detect_info"

    :: Escape CONNECTION_DIR for JSON (escape backslashes manually)
    set "CONNECTION_DIR_ESCAPED=!CONNECTION_DIR!"
    :: Replace backslashes with double backslashes
    :: In batch, we need to use a workaround: replace \ with a placeholder, then replace placeholder with \\
    set "CONNECTION_DIR_ESCAPED=!CONNECTION_DIR_ESCAPED:\=__BSLASH__!"
    set "CONNECTION_DIR_ESCAPED=!CONNECTION_DIR_ESCAPED:__BSLASH__=\\!"
    :: Replace quotes with escaped quotes (if any)
    set "CONNECTION_DIR_ESCAPED=!CONNECTION_DIR_ESCAPED:"=\"!"

    :: Build JSON request body manually
    set "json_body={"oper_inst_id":"!OPER_INST_ID!","action_name":"!ACTION_NAME_REPORT_DETECT_INFO!","os_type":"!OS_TYPE!","cpu_arch":"!CPU_ARCH!","connection_dir":"!CONNECTION_DIR_ESCAPED!","err_msg":""}"

    :: Send POST request using PowerShell
    set "temp_file=%TEMP%\nodemgr_report_%RANDOM%.json"
    set "http_code="
    set "response_content="
    set "response_line="
    set "NODEMGR_JSON_BODY=!json_body!"
    powershell -NoProfile -Command "$jsonBody = $env:NODEMGR_JSON_BODY; Set-Content -Path \"!temp_file!\" -Value $jsonBody -Encoding UTF8" >nul 2>&1
    set "NODEMGR_JSON_BODY="

    :: Use PowerShell with proper error handling to get HTTP status code and response
    set "temp_response_file=%TEMP%\nodemgr_response_%RANDOM%.txt"
    set "NODEMGR_URL=!url!"
    set "NODEMGR_TEMP_FILE=!temp_file!"
    powershell -NoProfile -Command "$url = $env:NODEMGR_URL; $bodyFile = $env:NODEMGR_TEMP_FILE; try { $body = (Get-Content $bodyFile) -join \"`n\"; $response = Invoke-WebRequest -Uri $url -Method POST -ContentType 'application/json' -Body $body -UseBasicParsing -ErrorAction Stop; Write-Output $response.StatusCode; Write-Output $response.Content } catch { if ($_.Exception.Response) { Write-Output $_.Exception.Response.StatusCode.value__; $stream = $_.Exception.Response.GetResponseStream(); $reader = New-Object System.IO.StreamReader($stream); Write-Output $reader.ReadToEnd() } else { Write-Output '000'; Write-Output $_.Exception.Message } }" > "!temp_response_file!" 2>nul
    set "NODEMGR_URL="
    set "NODEMGR_TEMP_FILE="

    :: Read HTTP status code and response content from temp file
    set "line_count=0"
    if exist "!temp_response_file!" (
        for /f "usebackq delims=" %%i in ("!temp_response_file!") do (
            set /a line_count+=1
            set "response_line=%%i"
            if !line_count! equ 1 (
                set "http_code=!response_line!"
            ) else (
                if defined response_content (
                    set "response_content=!response_content! !response_line!"
                ) else (
                    set "response_content=!response_line!"
                )
            )
        )
    )

    del "!temp_file!" 2>nul
    del "!temp_response_file!" 2>nul

    if "!http_code!"=="" (
        call :msg error_connect_server "!url!" "000"
        exit /b 1
    )

    if not "!http_code!"=="200" if not "!http_code!"=="201" (
        call :msg error_report_failed "!http_code!"
        if defined response_content call :msg error_report_response "!response_content!"
        exit /b 1
    )

    call :msg success_report_info
    exit /b 0

:: Download installer to specified path
:download_installer
    set "installer_path=%~1"
    set "url=!DOWNLOAD_SERVER_URL!/download/installer"

    :: Extract directory from installer path
    set "installer_dir="
    for %%i in ("!installer_path!") do set "installer_dir=%%~dpi"
    :: If no drive/path extracted, use parent directory
    if "!installer_dir!"=="" (
        for %%i in ("!installer_path!") do set "installer_dir=%%~pi"
    )
    :: If still empty, use current directory
    if "!installer_dir!"=="" set "installer_dir=.\"

    :: Create directory if it doesn't exist
    if not exist "!installer_dir!" (
        mkdir "!installer_dir!" 2>nul
        if errorlevel 1 (
            call :msg error_create_dir_failed "!installer_dir!"
            exit /b 1
        )
    )

    :: Build JSON request body manually
    set "json_body={"os_type":"!OS_TYPE!","cpu_arch":"!CPU_ARCH!"}"

    :: Download installer file using PowerShell
    set "temp_file=%TEMP%\nodemgr_download_%RANDOM%.json"
    set "http_code="
    set "error_content="
    set "NODEMGR_JSON_BODY=!json_body!"
    powershell -NoProfile -Command "$jsonBody = $env:NODEMGR_JSON_BODY; Set-Content -Path \"!temp_file!\" -Value $jsonBody -Encoding UTF8" >nul 2>&1
    set "NODEMGR_JSON_BODY="

    :: Use PowerShell with proper error handling to get HTTP status code
    set "temp_status_file=%TEMP%\nodemgr_download_status_%RANDOM%.txt"
    set "NODEMGR_URL=!url!"
    set "NODEMGR_TEMP_FILE=!temp_file!"
    set "NODEMGR_INSTALLER_PATH=!installer_path!"
    powershell -NoProfile -Command "$url = $env:NODEMGR_URL; $bodyFile = $env:NODEMGR_TEMP_FILE; $outFile = $env:NODEMGR_INSTALLER_PATH; $body = (Get-Content $bodyFile) -join \"`n\"; $response = $null; try { $response = Invoke-WebRequest -Uri $url -Method POST -ContentType 'application/json' -Body $body -OutFile $outFile -UseBasicParsing -ErrorAction Stop } catch { $statusCode = '000'; if ($_.Exception.Response) { $statusCode = $_.Exception.Response.StatusCode.value__ } if (Test-Path $outFile) { Remove-Item $outFile -Force }; Write-Output $statusCode; exit }; if ($response -and $response.StatusCode) { Write-Output $response.StatusCode } else { if (Test-Path $outFile) { Write-Output '200' } else { Write-Output '000' } }" > "!temp_status_file!" 2>nul
    set "NODEMGR_URL="
    set "NODEMGR_TEMP_FILE="
    set "NODEMGR_INSTALLER_PATH="

    :: Read HTTP status code from temp file
    if exist "!temp_status_file!" (
        for /f "usebackq delims=" %%i in ("!temp_status_file!") do set "http_code=%%i"
    )

    del "!temp_file!" 2>nul
    del "!temp_status_file!" 2>nul

    if "!http_code!"=="" set "http_code=000"

    if not "!http_code!"=="200" (
        :: Delete downloaded file if exists (may contain error response, don't print binary content)
        if exist "!installer_path!" del "!installer_path!" 2>nul
        call :msg error_download_installer_failed "!http_code!"
        exit /b 1
    )

    call :msg success_download_installer "!installer_path!"
    exit /b 0

:: Get command from server
:get_command
    set "url=!CALLBACK_SERVER_URL!/callback/workflow/node_install/get_manual_install_exec_command"

    :: Build JSON request body manually
    set "json_body={"oper_inst_id":"!OPER_INST_ID!","action_name":"!ACTION_NAME_GET_EXEC_COMMAND!"}"

    :: Send POST request using PowerShell
    set "temp_file=%TEMP%\nodemgr_getcmd_%RANDOM%.json"
    set "http_code="
    set "response_content="
    set "response_line="
    set "NODEMGR_JSON_BODY=!json_body!"
    powershell -NoProfile -Command "$jsonBody = $env:NODEMGR_JSON_BODY; Set-Content -Path \"!temp_file!\" -Value $jsonBody -Encoding UTF8" >nul 2>&1
    set "NODEMGR_JSON_BODY="

    :: Use PowerShell with proper error handling to get HTTP status code and response
    set "temp_response_file=%TEMP%\nodemgr_getcmd_response_%RANDOM%.txt"
    set "NODEMGR_URL=!url!"
    set "NODEMGR_TEMP_FILE=!temp_file!"
    powershell -NoProfile -Command "$url = $env:NODEMGR_URL; $bodyFile = $env:NODEMGR_TEMP_FILE; try { $body = (Get-Content $bodyFile) -join \"`n\"; $response = Invoke-WebRequest -Uri $url -Method POST -ContentType 'application/json' -Body $body -UseBasicParsing -ErrorAction Stop; Write-Output $response.StatusCode; Write-Output $response.Content } catch { if ($_.Exception.Response) { Write-Output $_.Exception.Response.StatusCode.value__; $stream = $_.Exception.Response.GetResponseStream(); $reader = New-Object System.IO.StreamReader($stream); Write-Output $reader.ReadToEnd() } else { Write-Output '000'; Write-Output $_.Exception.Message } }" > "!temp_response_file!" 2>nul
    set "NODEMGR_URL="
    set "NODEMGR_TEMP_FILE="

    :: Read HTTP status code and response content from temp file
    set "line_count=0"
    set "installer_path="
    set "command="
    if exist "!temp_response_file!" (
        for /f "usebackq delims=" %%i in ("!temp_response_file!") do (
            set /a line_count+=1
            set "response_line=%%i"
            if !line_count! equ 1 (
                set "http_code=!response_line!"
            ) else if !line_count! equ 2 (
                :: First line after HTTP code is installer_path
                set "installer_path=!response_line!"
            ) else if !line_count! equ 3 (
                :: Second line after HTTP code is command
                set "command=!response_line!"
            ) else (
                :: Append remaining lines to command (in case command spans multiple lines)
                if defined command (
                    set "command=!command! !response_line!"
                ) else (
                    set "command=!response_line!"
                )
            )
        )
    )

    del "!temp_file!" 2>nul
    del "!temp_response_file!" 2>nul

    if "!http_code!"=="" (
        call :msg error_connect_server "!url!" "000"
        exit /b 1
    )

    if not "!http_code!"=="200" (
        call :msg error_get_command_failed "!http_code!"
        if defined response_content call :msg error_get_command_response "!response_content!"
        exit /b 1
    )

    :: Trim whitespace from installer_path and command
    if defined installer_path (
        for /f "tokens=* delims= " %%a in ("!installer_path!") do set "installer_path=%%a"
        for /l %%a in (1,1,100) do if "!installer_path:~-1!"==" " set "installer_path=!installer_path:~0,-1!"
    )
    if defined command (
        for /f "tokens=* delims= " %%a in ("!command!") do set "command=%%a"
        for /l %%a in (1,1,100) do if "!command:~-1!"==" " set "command=!command:~0,-1!"
    )

    :: If installer_path or command is empty, it means not ready yet, return 2 to indicate need to continue waiting
    if "!installer_path!"=="" exit /b 2
    if "!command!"=="" exit /b 2
    if "!installer_path!"=="null" exit /b 2
    if "!command!"=="null" exit /b 2
    if "!installer_path!"=="\"\"" exit /b 2
    if "!command!"=="\"\"" exit /b 2

    exit /b 0

:: Main function
:main
    call :msg info_start_install
    call :msg info_oper_inst_id "!OPER_INST_ID!"
    call :msg info_callback_server_url "!CALLBACK_SERVER_URL!"
    call :msg info_download_server_url "!DOWNLOAD_SERVER_URL!"

    :: Detect system information
    call :detect_system_info
    if errorlevel 1 (
        call :msg error_detect_failed
        exit /b 1
    )

    :: Report detect information
    call :report_detect_info
    if errorlevel 1 (
        call :msg error_report_failed_main
        exit /b 1
    )

    :: Periodically get command until timeout or command is ready
    call :msg info_start_get_commands
    call :msg info_getting_command

    set "max_timeout=300"
    set "wait_times=0"
    set "sleep_interval=1"
    set "command="

:wait_loop
    :: Check timeout by wait times.
    set /a wait_times+=1

    if !wait_times! geq !max_timeout! (
        call :msg error_get_command_timeout "!max_timeout!"
        exit /b 1
    )

    :: Try to get command
    set "installer_path="
    set "command="
    set "get_result="
    call :get_command
    set "get_result=!errorlevel!"
    if "!get_result!"=="0" (
        if defined installer_path if defined command (
            goto :command_ready
        )
    ) else if "!get_result!"=="2" (
        :: Command not ready yet, continue waiting
        :: Output prompt every 10 seconds
        set /a remainder=!elapsed! %% 10
        if !remainder! equ 0 if !elapsed! gtr 0 (
            call :msg info_waiting_command "!elapsed!"
        )
        :: Use ping for sleep (compatible with all Windows versions)
        :: ping sends packets with ~1 second interval, so for N seconds we need N+1 packets
        set /a ping_count=!sleep_interval!+1
        ping 127.0.0.1 -n !ping_count! >nul 2>&1
        goto :wait_loop
    ) else (
        :: Error occurred
        call :msg error_get_execute_failed
        exit /b 1
    )

:command_ready
    if "!installer_path!"=="" (
        call :msg error_get_execute_failed
        exit /b 1
    )
    if "!command!"=="" (
        call :msg error_get_execute_failed
        exit /b 1
    )

    :: Download installer to the path specified in command
    call :msg info_download_installer
    call :download_installer "!installer_path!"
    if errorlevel 1 (
        exit /b 1
    )

    :: Execute command
    call :msg info_execute_command "!command!"
    cmd /c "!command!"
    set "exec_result=!errorlevel!"

    :: Exit based on execution result
    if !exec_result! equ 0 (
        call :msg success_command_executed
        exit /b 0
    ) else (
        call :msg warning_command_exit_code "!exec_result!"
        exit /b !exec_result!
    )