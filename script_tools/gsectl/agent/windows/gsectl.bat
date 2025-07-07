@echo off

setlocal EnableDelayedExpansion
set agent_config_file=gse_agent.conf
set cu_date=%date:~0,4%-%date:~5,2%-%date:~8,2%
set cu_time=%time:~0,8%
set gse_agent_daemon_path=%cd%
set gse_agent_path=%cd%
set gse_winagent_home=%cd:~0,-10%

set COMMAND=%1

set INSTALL_USER=
set INSTALL_PASSWORD=
if "%COMMAND%"=="start" goto parse_args
if "%COMMAND%"=="restart" goto parse_args
goto skip_parse

:parse_args
shift
:parse_loop
if "%1"=="" goto end_parse
set arg=%1
if "%arg:~0,11%"=="--installer=" (
    set INSTALL_USER=%arg:~11%
) else if "%arg:~0,11%"=="--password=" (
    set INSTALL_PASSWORD=%arg:~11%
)
shift
goto parse_loop
:end_parse

if not "%INSTALL_USER%"=="" (
    if "%INSTALL_PASSWORD%"=="" (
        call :log Error: Password required for user %INSTALL_USER% -- parameter check/check process
        exit /b 1
    )
)

:skip_parse
for %%a in ("%gse_winagent_home%") do ( set service_id=%%~nxa)
if %service_id%=="gse" (set _service_id=) else (set _service_id=_%service_id%)
set gse_winagent_home=%gse_winagent_home:\=\\%
set gse_agent_config_path=%gse_winagent_home:"=%\\agent\\etc\\%agent_config_file%

:: set gse_agent_daemon_path
set TMP_DIR=C:\tmp
set gse_agent_restart_log=%TMP_DIR%\restart_gse_agent_%service_id%.log
set gsectl_log=%TMP_DIR%\gsectl_%service_id%_%cu_date%.log

if not exist "%TMP_DIR%" mkdir "%TMP_DIR%"

echo ================ [%cu_date% %cu_time%] "%COMMAND%"  ================= >> %gsectl_log%

if "%COMMAND%"=="" goto :usage
if /i "%COMMAND%"=="start" goto :start
if /i "%COMMAND%"=="stop" goto :stop
if /i "%COMMAND%"=="restart" goto :restart
if /i "%COMMAND%"=="status" goto :status
if /i "%COMMAND%"=="version" goto :version
goto :usage

:ensure_gse_agent
    :: check if gse_agent_daemon service exists
    sc query gse_agent_daemon_%service_id% > nul 2>&1
    if %errorlevel% equ 0 (
        call :log gse_agent_daemon_%service_id% exists -- check service/check process
        goto :EOF
    )
    :: if not exists, install gse_agent_daemon service
    call :install_gse_agent_service
goto :EOF

:install_gse_agent_service
    if "%INSTALL_USER%" == "" (
        %gse_winagent_home%\\agent\\bin\\gse_agent_daemon.exe --install -f %gse_winagent_home%\\agent\\etc\\gse_agent.conf --name gse_agent_daemon_%service_id% >>%gsectl_log% 2>&1
        if %errorlevel% equ 0 (
            call :log create gse_agent service without special user succeed -- install service/install process
        ) else if %errorlevel% equ 1060 (
            call :log service already exists -- install service/install process
        ) else (
            call :log create gse_agent service without special user failed, errorlevel: %errorlevel% -- install service/install process
        )
    ) else (
        %gse_winagent_home%\\agent\\bin\\gse_agent_daemon.exe --install -f %gse_winagent_home%\\agent\\etc\\gse_agent.conf --name gse_agent_daemon_%service_id% --user .\\%INSTALL_USER% --pwd %INSTALL_PASSWORD% --encode >>%gsectl_log% 2>&1
        if %errorlevel% equ 0 (
            call :log create gse_agent service with special user: %INSTALL_USER% succeed -- install service/install process
        ) else if %errorlevel% equ 1060 (
            call :log service already exists -- install service/install process
        )  else (
            call :log create gse_agent service with special user %INSTALL_USER% failed -- install service/install process
        )
    )
goto :EOF

:start
    echo ================ [%date%-%time%] start ======================  >>%gse_agent_restart_log%

    call :ensure_gse_agent

    :: check if gse_agent_daemon%_service_id% is already running with established connection
    sc query gse_agent_daemon%_service_id% 2>&1 | findstr /r /i /C:"RUNNING" 1>nul 2>&1
    if %errorlevel% EQU 0 (
        netstat -an | findstr /r /C:":28668 *ESTABLISHED" 1>nul 2>&1
        if %errorlevel% EQU 0 (
            call :log "gse_agent_daemon%_service_id% is already running with established connection -- skip start/start process"
            goto :EOF
        )
    )

    echo [%cu_date% %cu_time%] Attempting to add auto-start registry entry -- registry config/config process >>%gse_agent_restart_log%
    call :log Attempting to add auto-start registry entry -- registry config/config process
    reg add "HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run" /v gse_agent_daemon%_service_id% /t reg_sz /d "%gse_winagent_home%\agent\bin\gsectl.bat start" /f >>%gsectl_log% 2>&1
    if %errorlevel% equ 0 (
        call :log Add auto-start registry succeeded -- registry config/config process
    ) else (
        call :log Add auto-start registry failed -- registry config/config process
    )

    sc query gse_agent_daemon%_service_id% 2>&1 | findstr /r /i /C:"RUNNING" 1>nul 2>&1
    if %errorlevel% NEQ 0 (
        %gse_winagent_home%\\agent\\bin\\gse_agent_daemon.exe --start --name gse_agent_daemon%_service_id% 1>nul 2>&1
    )

    ping -n 5 127.0.0.1 >nul 2>&1
    sc query gse_agent_daemon%_service_id% 2>&1 | findstr /r /i /C:"RUNNING" 1>nul 2>&1
    if %errorlevel% NEQ 0 (
        echo [%date%-%time%] gse_agent_daemon%_service_id% Service Status: NOT RUNNING >>%gse_agent_restart_log%
        call :log gse_agent_daemon%_service_id% Service Status: NOT RUNNING
        echo [%date%-%time%] start fail -- gse_agent start fail/start process >>%gse_agent_restart_log%
        call :log start fail -- gse_agent start fail/start process
        sc query gse_agent_daemon%_service_id% >>%gse_agent_restart_log%
        sc query gse_agent_daemon%_service_id% >>%gsectl_log%
        exit /b 1
    )

    netstat -an | findstr /r /C:":28668 *ESTABLISHED"
    if %errorlevel% EQU 0 (
        call :log start done -- start done/start process
        call :log "gse_agent_daemon%_service_id% Service Status: RUNNING and Network Connection: ESTABLISHED"
        sc query gse_agent_daemon%_service_id%
        netstat -an | findstr /r /C:":28668 *ESTABLISHED"
    ) else (
        call :log start fail -- gse_agent start fail/start process
        call :log "gse_agent_daemon%_service_id% Service Status: RUNNING and Network Connection: NOT ESTABLISHED"
        sc query gse_agent_daemon%_service_id%
        netstat -an | findstr /r /C:":28668 *ESTABLISHED"
        exit /b 1
    )
goto :EOF

:stop
    :: remove auto-start registry entry
    call :log "Attempting to remove auto-start registry entry -- registry config/config process"
    reg delete "HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run" /v gse_agent_daemon%_service_id% /f >>%gsectl_log% 2>&1
    if %errorlevel% equ 0 (
        call :log "Remove auto-start registry succeeded -- registry config/config process"
    ) else (
        call :log "Remove auto-start registry failed -- registry config/config process"
    )

    :: check if gse_agent_daemon service exists
    sc query gse_agent_daemon%_service_id% | findstr /i "SERVICE_NAME" 1>nul 2>&1
    if %errorlevel% neq 0 (
        echo [%date%-%time%] stop fail -- service not exist >>%gse_agent_restart_log%
        call :log stop fail -- service not exist
        goto :EOF
    )

    :: check if gse_agent_daemon service stopped
    sc query gse_agent_daemon%_service_id% | findstr /r /i /C:" *STOPPED" 1>nul 2>&1
    if %errorlevel% equ 0 (
        echo [%date%-%time%] already stoped >>%gse_agent_restart_log%
        call :log already stoped
        goto :EOF
    )

    %gse_winagent_home%\\agent\\bin\\gse_agent_daemon.exe --quit --name gse_agent_daemon%_service_id% 1>nul 2>&1
    if %errorlevel% neq 0 (
        echo [%date%-%time%] stop Service gse_agent_daemon%_service_id% failed, then use wmic to terminate process >>%gse_agent_restart_log%
        call :log stop Service gse_agent_daemon%_service_id% failed, then use wmic to terminate process
        ping -n 5 127.0.0.1 1>nul 2>&1

        :: stop related processes
        wmic process where name="'gse_agent_daemon.exe' and ExecutablePath='%gse_winagent_home:"=%\\agent\\bin\\gse_agent_daemon.exe'" call terminate 1>nul 2>&1
        wmic process where name="'gse_agent.exe' and ExecutablePath='%gse_winagent_home:"=%\\agent\\bin\\gse_agent.exe'" call terminate 1>nul 2>&1
    )

    :: check if gse_agent_daemon service stopped
    sc query gse_agent_daemon%_service_id% | findstr /r /i /C:" *STOPPED" 1>nul 2>&1
    if %errorlevel% equ 0 (
        echo [%date%-%time%] stop Service gse_agent_daemon%_service_id% success >>%gse_agent_restart_log%
        call :log stop Service gse_agent_daemon%_service_id% success
    ) else (
        sc stop gse_agent_daemon%_service_id%
        ping -n 5 127.0.0.1 1>nul 2>&1

        sc query gse_agent_daemon%_service_id% | findstr /r /i /C:" *STOPPED" 1>nul 2>&1
        if %errorlevel% equ 0 (
            echo [%date%-%time%] stop Service gse_agent_daemon%_service_id% success >>%gse_agent_restart_log%
            call :log stop Service gse_agent_daemon%_service_id% success
        ) else (
            echo [%date%-%time%] stop Service gse_agent_daemon%_service_id% failed >>%gse_agent_restart_log%
            call :log stop Service gse_agent_daemon%_service_id% failed
        )
    )

    :: uninstall service
    call :log "Attempting to uninstall service -- service uninstall/uninstall process"
    %gse_winagent_home%\\agent\\bin\\gse_agent_daemon.exe --uninstall --name gse_agent_daemon%_service_id% >>%gsectl_log% 2>&1
    if %errorlevel% equ 0 (
        call :log "Service uninstall succeeded -- service uninstall/uninstall process"
    ) else (
        call :log "Service uninstall failed -- service uninstall/uninstall process"
    )

goto :EOF

:restart
    call :stop
    call :start
goto :EOF

:status
    sc query gse_agent_daemon%_service_id% 2>&1 | findstr /r /i /C:"RUNNING" 1>nul 2>&1
    if %errorlevel% NEQ 0 (
        echo [%date%-%time%] gse_agent_daemon%_service_id% Service Status: NOT RUNNING >>%gse_agent_restart_log%
        call :log gse_agent_daemon%_service_id% Service Status: NOT RUNNING
        sc query gse_agent_daemon%_service_id%
        exit /b 1
    )

    ping -n 5 127.0.0.1 1>nul 2>&1
    netstat -an | findstr /r /C:":28668 *ESTABLISHED" 1>nul 2>&1
    if %errorlevel% EQU 0 (
        echo [%date%-%time%] gse_agent_daemon%_service_id% Service Status: RUNNING and Network Connection: ESTABLISHED >>%gse_agent_restart_log%
        call :log gse_agent_daemon%_service_id% Service Status: RUNNING and Network Connection: ESTABLISHED
        sc query gse_agent_daemon%_service_id%
        netstat -an | findstr /r /C:":28668 *ESTABLISHED"
    ) else (
        echo [%date%-%time%] gse_agent_daemon%_service_id% Service Status: RUNNING and Network Connection: NOT ESTABLISHED >>%gse_agent_restart_log%
        call :log gse_agent_daemon%_service_id% Service Status: RUNNING and Network Connection: NOT ESTABLISHED
        sc query gse_agent_daemon%_service_id%
        netstat -an | findstr /r /C:":28668 *ESTABLISHED"
        exit /b 1
    )
goto :EOF

:version
%gse_agent_daemon_path%\gse_agent.exe --version
goto :EOF

:usage
    echo Usage: gsectl.bat COMMAND [PARAMETERS]
    echo.
    echo COMMANDS:
    echo    start    start the gse_agent
    echo    stop     stop the gse_agent
    echo    restart  restart the gse_agent
    echo    status   status of the gse_agent
    echo    version  get the gse_agent version
    echo.
    echo PARAMETERS ^(only valid for start and restart commands^):
    echo    --installer=USERNAME    Specify the service installer username
    echo    --password=PASSWORD     Specify the installer password
    echo.
    echo Examples:
    echo    gsectl.bat start
    echo    gsectl.bat start --installer=serviceuser --password=pass123
    echo    gsectl.bat restart --installer=serviceuser --password=pass123
    echo    gsectl.bat stop
    echo    gsectl.bat status
goto :EOF

:log
    setlocal
    echo [%cu_date% %cu_time%] %* >>%gsectl_log%
    echo [%cu_date% %cu_time%] %*
    endlocal
goto :EOF