@echo off
setlocal EnableDelayedExpansion
set prog_name=%1
set cu_date=%date:~0,4%-%date:~5,2%-%date:~8,2%
set cu_time=%time:~0,8%
set script_dir=%~dp0
cd /d %script_dir%

powershell -command "$expectedPath = '%script_dir%!prog_name!.exe' -replace '/', '\'; $process = Get-Process -Name '!prog_name!' -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $expectedPath }; if ($process) { exit 0 } else { exit 1 }" >nul 2>&1
if %errorlevel% neq 0 (
    echo [%cu_date% %cu_time%] !prog_name! already stopped in current directory
    goto EOF
)


rem 停止进程
rem 停止进程（精确匹配路径）
for /f "delims=" %%p in ('powershell -command "$expectedPath = '%script_dir%!prog_name!.exe' -replace '/', '\'; $process = Get-Process -Name '!prog_name!' -ErrorAction SilentlyContinue ^| Where-Object { $_.Path -eq $expectedPath }; if ($process) { $process.Id }"') do (
    taskkill /F /PID %%p
)
ping -n 2 127.0.0.1 >nul 2>&1
powershell -command "$expectedPath = '%script_dir%!prog_name!.exe' -replace '/', '\'; $process = Get-Process -Name '!prog_name!' -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $expectedPath }; if ($process) { exit 0 } else { exit 1 }" >nul 2>&1
if %errorlevel% equ 0 (
    echo [%cu_date% %cu_time%] stop !prog_name! fail
    goto EOF
)

echo [%cu_date% %cu_time%] stop !prog_name! done

:EOF
